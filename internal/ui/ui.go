// Package ui renders a Suggestion and asks the user what to do with it.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ninjattachee/shell_commander/internal/provider"
)

// Action is what the user chose to do with the suggested command.
type Action int

const (
	Quit Action = iota
	RunIt
)

// Render prints the command, a step-by-step walkthrough, and any warning.
func Render(w io.Writer, s provider.Suggestion) {
	fmt.Fprintf(w, "\nCommand:\n  %s\n", s.Command)
	if len(s.Steps) > 0 {
		fmt.Fprintln(w, "\nWhat it does:")
		for i, st := range s.Steps {
			fmt.Fprintf(w, "  %d. %s\n     %s\n", i+1, st.Part, st.Explanation)
		}
	}
	if s.Warning != "" {
		fmt.Fprintf(w, "\n⚠ WARNING: %s\n", s.Warning)
	}
	fmt.Fprintln(w)
}

// Prompt asks what to do until it gets a usable answer. It returns the action
// and the (possibly edited) command. Commands with a warning must be confirmed
// by typing "yes" instead of a single keypress.
func Prompt(in *bufio.Reader, w io.Writer, s provider.Suggestion) (Action, string, error) {
	cmd := s.Command
	for {
		fmt.Fprint(w, "[r]un  [e]dit  [c]opy  [q]uit > ")
		line, err := in.ReadString('\n')
		if err != nil && line == "" {
			return Quit, cmd, nil // EOF: treat as quit, never run
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "r", "run":
			if s.Warning != "" {
				fmt.Fprint(w, `This command is flagged. Type "yes" to run it: `)
				ans, _ := in.ReadString('\n')
				if strings.TrimSpace(ans) != "yes" {
					fmt.Fprintln(w, "Not confirmed.")
					continue
				}
			}
			return RunIt, cmd, nil
		case "e", "edit":
			fmt.Fprint(w, "Edit command: ")
			edited, _ := in.ReadString('\n')
			if edited = strings.TrimSpace(edited); edited != "" {
				cmd = edited
				fmt.Fprintf(w, "Now: %s\n", cmd)
			}
		case "c", "copy":
			if err := copyToClipboard(cmd); err != nil {
				fmt.Fprintf(w, "Could not copy: %v\n", err)
			} else {
				fmt.Fprintln(w, "Copied.")
			}
		case "q", "quit":
			return Quit, cmd, nil
		default:
			fmt.Fprintln(w, "Please answer r, e, c, or q.")
		}
	}
}

// copyToClipboard tries the common clipboard tools in turn.
func copyToClipboard(text string) error {
	for _, tool := range [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"pbcopy"}} {
		if _, err := exec.LookPath(tool[0]); err != nil {
			continue
		}
		c := exec.Command(tool[0], tool[1:]...)
		c.Stdin = strings.NewReader(text)
		return c.Run()
	}
	return fmt.Errorf("no clipboard tool found (wl-copy, xclip, pbcopy)")
}
