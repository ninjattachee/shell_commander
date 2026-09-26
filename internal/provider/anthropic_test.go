package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnthropicSuggest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "k" {
			t.Errorf("missing api key header")
		}
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		if req["tool_choice"] == nil {
			t.Errorf("tool_choice not set")
		}
		w.Write([]byte(`{"content":[{"type":"tool_use","input":{
			"command":"mkdir enguy && cd enguy && deno init",
			"steps":[{"part":"mkdir enguy","explanation":"create the folder"}],
			"warning":""}}]}`))
	}))
	defer srv.Close()

	a := &Anthropic{APIKey: "k", URL: srv.URL}
	got, err := a.Suggest(context.Background(), Request{Task: "make enguy with deno"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "mkdir enguy && cd enguy && deno init" || len(got.Steps) != 1 {
		t.Fatalf("unexpected suggestion: %+v", got)
	}
}

func TestAnthropicErrors(t *testing.T) {
	if _, err := (&Anthropic{}).Suggest(context.Background(), Request{}); err == nil ||
		!strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("want missing-key error, got %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad key", http.StatusUnauthorized)
	}))
	defer srv.Close()
	if _, err := (&Anthropic{APIKey: "k", URL: srv.URL}).Suggest(context.Background(), Request{}); err == nil {
		t.Fatal("want HTTP error")
	}
}
