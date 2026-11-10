# shell-commander

Describe what you want in plain English. shell-commander suggests a shell command, walks you through what each part does, and lets you run it right away.

```console
$ shell-commander create a project named enguy and initialize it with Deno

Command:
  mkdir enguy && cd enguy && deno init

What it does:
  1. mkdir enguy
     creates a new directory named enguy to hold the project files
  2. cd enguy
     changes the current working directory to the new folder
  3. deno init
     initializes a new Deno project with a basic file structure

[r]un  [e]dit  [c]opy  [q]uit >
```

## Install

Requires Go 1.27+.

```sh
go install github.com/ninjattachee/shell_commander@latest
```

This installs a binary named `shell_commander`. Rename or alias it to `shell-commander` if you prefer. Or build from a clone:

```sh
go build -o shell-commander .
```

## Usage

```sh
shell-commander [flags] <what you want to do>
shell-commander auth [provider]
```

Put flags **before** the request. Everything after the first non-flag word is treated as the request, so quoting is not needed.

| Flag | Description |
| --- | --- |
| `--provider` | `deepseek`, `anthropic` or `ollama`. Env: `SHELL_COMMANDER_PROVIDER`. Default: `deepseek`. |
| `--model` | Model name. Falls back to the provider's default. |
| `--explain-only` | Show the command and explanation; never prompt or run. |
| `--yes` | Run without prompting. Refused if the command is flagged as risky. |

At the prompt:

- `r` runs the command in your `$SHELL` (`$SHELL -c`), so pipes, `&&` and globs work as usual. The command's exit code becomes shell-commander's exit code.
- `e` lets you edit the command before running it.
- `c` copies it to the clipboard (`wl-copy`, `xclip` or `pbcopy`).
- `q` quits without running anything. End of input (Ctrl-D) also quits.

If the model flags a command as destructive, irreversible or privileged, a warning is shown and you must type `yes` to run it.

## Credentials

API keys are read from the environment first, then from the OS keyring. Save a key once instead of exporting it in every shell:

```sh
shell-commander auth              # DeepSeek (default)
shell-commander auth anthropic
```

You are prompted once with echo disabled, and the key is stored in Secret Service (Linux), Keychain (macOS) or Credential Manager (Windows). Environment variables still take precedence, which is handy for scripts and CI. Piped input also works:

```sh
printf '%s\n' "$DEEPSEEK_API_KEY" | shell-commander auth
```

## Providers

**DeepSeek** (default) needs an API key, either stored with `shell-commander auth` or exported as `DEEPSEEK_API_KEY`. The default model is `deepseek-flash`:

```sh
shell-commander find files changed in the last day
```

**Anthropic** needs an API key, either stored with `shell-commander auth anthropic` or exported as `ANTHROPIC_API_KEY`:

```sh
shell-commander --provider anthropic find files changed in the last day
```

The default Anthropic model is `claude-sonnet-5`.

**Ollama** needs a running server at `http://localhost:11434`. The default model is `gemma4:31b-cloud`. Models with a `-cloud` suffix are run by Ollama's hosted service, so your request leaves your machine. Use a local model to keep it private:

```sh
shell-commander --provider ollama --model llama3.1 list the 10 largest files here
```

## Safety

Model suggestions can be wrong. Read the command and explanation before you run anything. Nothing runs without your confirmation, except with `--yes` on a command the model did not flag, and the model's flagging is not a guarantee.

## Development

```sh
go vet ./... && go test ./...
```

Layout:

- `main.go`: flags and the suggest, render, prompt, run flow.
- `internal/provider`: the `Provider` interface, the shared prompt/schema, and the DeepSeek, Anthropic and Ollama implementations. Tests use `httptest` fakes, so no network or key is needed.
- `internal/credentials`: OS keyring storage for provider API keys.
- `internal/ui`: rendering and the interactive prompt.
- `internal/runner`: runs a command line through the shell.

## License

[MIT](LICENSE)
