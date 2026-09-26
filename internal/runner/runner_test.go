package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	var out bytes.Buffer
	code, err := Run(context.Background(), "/bin/sh", "echo hi | tr a-z A-Z && exit 3", strings.NewReader(""), &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if code != 3 {
		t.Errorf("exit code = %d, want 3", code)
	}
	if strings.TrimSpace(out.String()) != "HI" {
		t.Errorf("output = %q, want HI", out.String())
	}
}

func TestRunMissingShell(t *testing.T) {
	code, err := Run(context.Background(), "/no/such/shell", "true", strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || code != -1 {
		t.Fatalf("want start error, got code=%d err=%v", code, err)
	}
}
