// Command shell-commander suggests a shell command for a plain-English
// request, explains it step by step, and can run it on request.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"

	"github.com/ninjattachee/shell_commander/internal/provider"
	"github.com/ninjattachee/shell_commander/internal/runner"
	"github.com/ninjattachee/shell_commander/internal/ui"
)

func main() {
	os.Exit(run())
}

func run() int {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: shell-commander [flags] <what you want to do>")
		flag.PrintDefaults()
	}
	provName := flag.String("provider", envOr("SHELL_COMMANDER_PROVIDER", "ollama"), "LLM provider: anthropic or ollama")
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
	case "anthropic":
		return &provider.Anthropic{APIKey: os.Getenv("ANTHROPIC_API_KEY"), Model: model}, nil
	case "ollama":
		return &provider.Ollama{Model: model}, nil
	}
	return nil, fmt.Errorf("unknown provider %q (want anthropic or ollama)", name)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
