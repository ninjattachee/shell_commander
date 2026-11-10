package main

import (
	"os"
	"testing"

	"github.com/ninjattachee/shell_commander/internal/credentials"
	"github.com/zalando/go-keyring"
)

func TestAPIKeyPrefersEnvironment(t *testing.T) {
	keyring.MockInit()
	if err := credentials.Set("deepseek", "from-keyring"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEEPSEEK_API_KEY", "from-env")

	if got := apiKey("deepseek"); got != "from-env" {
		t.Fatalf("apiKey = %q, want from-env", got)
	}
}

func TestAPIKeyFallsBackToKeyring(t *testing.T) {
	keyring.MockInit()
	if err := credentials.Set("deepseek", "from-keyring"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEEPSEEK_API_KEY", "")

	if got := apiKey("deepseek"); got != "from-keyring" {
		t.Fatalf("apiKey = %q, want from-keyring", got)
	}
}

func TestAuthSavesPipedKey(t *testing.T) {
	keyring.MockInit()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	go func() {
		w.WriteString("sk-piped\n")
		w.Close()
	}()

	if code := auth([]string{"deepseek"}, r); code != 0 {
		t.Fatalf("auth exited with %d", code)
	}
	got, err := credentials.Get("deepseek")
	if err != nil {
		t.Fatal(err)
	}
	if got != "sk-piped" {
		t.Fatalf("stored key = %q, want sk-piped", got)
	}
}

func TestAuthRejectsKeylessProvider(t *testing.T) {
	if code := auth([]string{"ollama"}, nil); code != 2 {
		t.Fatalf("auth exited with %d, want 2", code)
	}
}
