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
```

Put flags **before** the request. Everything after the first non-flag word is treated as the request, so quoting is not needed.

| Flag | Description |
| --- | --- |
| `--provider` | `anthropic` or `ollama`. Env: `SHELL_COMMANDER_PROVIDER`. Default: `ollama`. |
| `--model` | Model name. Falls back to the provider's default. |
| `--explain-only` | Show the command and explanation; never prompt or run. |
| `--yes` | Run without prompting. Refused if the command is flagged as risky. |

At the prompt:

- `r` runs the command in your `$SHELL` (`$SHELL -c`), so pipes, `&&` and globs work as usual. The command's exit code becomes shell-commander's exit code.
- `e` lets you edit the command before running it.
- `c` copies it to the clipboard (`wl-copy`, `xclip` or `pbcopy`).
- `q` quits without running anything. End of input (Ctrl-D) also quits.

If the model flags a command as destructive, irreversible or privileged, a warning is shown and you must type `yes` to run it.

## Providers

**Ollama** (default) needs a running server at `http://localhost:11434`. The default model is `gemma4:31b-cloud`. Models with a `-cloud` suffix are run by Ollama's hosted service, so your request leaves your machine. Use a local model to keep it private:

```sh
shell-commander --model llama3.1 list the 10 largest files here
```

**Anthropic** needs an API key:

```sh
export ANTHROPIC_API_KEY=...
shell-commander --provider anthropic find files changed in the last day
```

The default Anthropic model is `claude-sonnet-5`.

## Safety

Model suggestions can be wrong. Read the command and explanation before you run anything. Nothing runs without your confirmation, except with `--yes` on a command the model did not flag, and the model's flagging is not a guarantee.

## Development

```sh
go vet ./... && go test ./...
```

Layout:

- `main.go`: flags and the suggest, render, prompt, run flow.
- `internal/provider`: the `Provider` interface, the shared prompt/schema, and the Anthropic and Ollama implementations. Tests use `httptest` fakes, so no network or key is needed.
- `internal/ui`: rendering and the interactive prompt.
- `internal/runner`: runs a command line through the shell.
