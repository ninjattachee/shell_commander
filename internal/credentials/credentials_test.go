package credentials

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestSetGetDelete(t *testing.T) {
	keyring.MockInit()

	if err := Set("deepseek", "sk-test"); err != nil {
		t.Fatal(err)
	}
	got, err := Get("deepseek")
	if err != nil {
		t.Fatal(err)
	}
	if got != "sk-test" {
		t.Fatalf("got %q, want sk-test", got)
	}

	if err := Delete("deepseek"); err != nil {
		t.Fatal(err)
	}
	if _, err := Get("deepseek"); !errors.Is(err, keyring.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}
