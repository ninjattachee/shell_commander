// Command shell-commander suggests a shell command for a plain-English
// request, explains it step by step, and can run it on request.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/ninjattachee/shell_commander/internal/credentials"
	"github.com/ninjattachee/shell_commander/internal/provider"
	"github.com/ninjattachee/shell_commander/internal/runner"
	"github.com/ninjattachee/shell_commander/internal/ui"
	"golang.org/x/term"
)

// apiKeyEnv maps providers that need a key to the environment variable that
// can supply it.
var apiKeyEnv = map[string]string{
	"deepseek":  "DEEPSEEK_API_KEY",
	"anthropic": "ANTHROPIC_API_KEY",
}

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		return auth(os.Args[2:], os.Stdin)
	}

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: shell-commander [flags] <what you want to do>")
		fmt.Fprintln(os.Stderr, "       shell-commander auth [provider]")
		flag.PrintDefaults()
	}
	provName := flag.String("provider", envOr("SHELL_COMMANDER_PROVIDER", "deepseek"), "LLM provider: deepseek, anthropic or ollama")
	model := flag.String("model", "", "model name (provider default if empty)")
	yes := flag.Bool("yes", false, "run without prompting (refused if the command is flagged as risky)")
	explainOnly := flag.Bool("explain-only", false, "show the command and explanation, never prompt or run")
	flag.Parse()

	task := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if task == "" {
		flag.Usage()
		return 2
	}

	p, err := newProvider(*provName, *model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cwd, _ := os.Getwd()
	shell := runner.Shell()
	fmt.Fprintln(os.Stderr, "Thinking...")
	s, err := p.Suggest(ctx, provider.Request{Task: task, OS: runtime.GOOS, Shell: shell, Cwd: cwd})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	ui.Render(os.Stdout, s)
	if *explainOnly {
		return 0
	}

	cmd := s.Command
	if *yes {
		if s.Warning != "" {
			fmt.Fprintln(os.Stderr, "refusing --yes: command is flagged as risky; run again without --yes")
			return 1
		}
	} else {
		action, edited, err := ui.Prompt(bufio.NewReader(os.Stdin), os.Stdout, s)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		if action != ui.RunIt {
			return 0
		}
		cmd = edited
	}

	code, err := runner.Run(ctx, shell, cmd, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return code
}

func newProvider(name, model string) (provider.Provider, error) {
	switch name {
	case "deepseek":
		return &provider.DeepSeek{APIKey: apiKey(name), Model: model}, nil
	case "anthropic":
		return &provider.Anthropic{APIKey: apiKey(name), Model: model}, nil
	case "ollama":
		return &provider.Ollama{Model: model}, nil
	}
	return nil, fmt.Errorf("unknown provider %q (want deepseek, anthropic or ollama)", name)
}

// apiKey returns the key for a provider, preferring the environment over the
// OS keyring so scripts and CI can override a stored key.
func apiKey(providerName string) string {
	if v := os.Getenv(apiKeyEnv[providerName]); v != "" {
		return v
	}
	v, err := credentials.Get(providerName)
	if err != nil {
		return ""
	}
	return v
}

// auth prompts once for a provider's API key and stores it in the OS keyring.
func auth(args []string, in *os.File) int {
	name := "deepseek"
	if len(args) > 0 {
		name = args[0]
	}
	envName, ok := apiKeyEnv[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: %q does not use an API key (want deepseek or anthropic)\n", name)
		return 2
	}

	fmt.Fprintf(os.Stderr, "Paste your %s API key (input hidden): ", name)
	key, err := readSecret(in)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: reading key:", err)
		return 1
	}
	key = strings.TrimSpace(key)
	if key == "" {
		fmt.Fprintln(os.Stderr, "error: empty key, nothing saved")
		return 1
	}
	if err := credentials.Set(name, key); err != nil {
		fmt.Fprintln(os.Stderr, "error: saving key in the OS keyring:", err)
		fmt.Fprintf(os.Stderr, "tip: export %s instead\n", envName)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Saved the %s key in the OS keyring. It is used when %s is not set.\n", name, envName)
	return 0
}

// readSecret reads one line with echo disabled when stdin is a terminal, so
// the key can also be piped in from another program.
func readSecret(in *os.File) (string, error) {
	if term.IsTerminal(int(in.Fd())) {
		b, err := term.ReadPassword(int(in.Fd()))
		return string(b), err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return line, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
