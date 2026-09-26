// Package runner executes a command line in the user's shell.
package runner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
)

// Run executes cmdline with `<shell> -c` so pipes, &&, and globs work as they
// would when typed. It returns the command's exit code. A non-zero exit is not
// an error; err is set only if the shell could not be started.
func Run(ctx context.Context, shell, cmdline string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	if shell == "" {
		shell = "/bin/sh"
	}
	c := exec.CommandContext(ctx, shell, "-c", cmdline)
	c.Stdin, c.Stdout, c.Stderr = stdin, stdout, stderr

	err := c.Run()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode(), nil
	}
	if err != nil {
		return -1, err
	}
	return 0, nil
}

// Shell returns the user's preferred shell, falling back to /bin/sh.
func Shell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/sh"
}
