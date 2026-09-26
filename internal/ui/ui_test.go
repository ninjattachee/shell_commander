package ui

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/ninjattachee/shell_commander/internal/provider"
)

func prompt(t *testing.T, input string, s provider.Suggestion) (Action, string, string) {
	t.Helper()
	var out bytes.Buffer
	a, cmd, err := Prompt(bufio.NewReader(strings.NewReader(input)), &out, s)
	if err != nil {
		t.Fatal(err)
	}
	return a, cmd, out.String()
}

func TestPrompt(t *testing.T) {
	s := provider.Suggestion{Command: "ls"}

	if a, _, _ := prompt(t, "r\n", s); a != RunIt {
		t.Error("r should run")
	}
	if a, _, _ := prompt(t, "q\n", s); a != Quit {
		t.Error("q should quit")
	}
	if a, _, _ := prompt(t, "", s); a != Quit {
		t.Error("EOF should quit, never run")
	}
	if a, cmd, _ := prompt(t, "e\nls -la\nr\n", s); a != RunIt || cmd != "ls -la" {
		t.Errorf("edit then run: got %v %q", a, cmd)
	}
	if _, _, out := prompt(t, "x\nq\n", s); !strings.Contains(out, "Please answer") {
		t.Error("invalid input should re-prompt")
	}
}

func TestPromptWarningNeedsYes(t *testing.T) {
	s := provider.Suggestion{Command: "rm -rf x", Warning: "deletes files"}

	if a, _, _ := prompt(t, "r\nno\nq\n", s); a != Quit {
		t.Error("declining confirmation must not run")
	}
	if a, _, _ := prompt(t, "r\nyes\n", s); a != RunIt {
		t.Error("typing yes should run")
	}
}

func TestRender(t *testing.T) {
	var out bytes.Buffer
	Render(&out, provider.Suggestion{
		Command: "deno init",
		Steps:   []provider.Step{{Part: "deno init", Explanation: "scaffold a project"}},
		Warning: "careful",
	})
	for _, want := range []string{"deno init", "scaffold a project", "WARNING: careful"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, out.String())
		}
	}
}
