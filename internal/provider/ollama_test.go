package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaSuggest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// message.content is itself a JSON document encoded as a string.
		w.Write([]byte(`{"message":{"content":"{\"command\":\"ls -la\",\"steps\":[{\"part\":\"-la\",\"explanation\":\"long format, include hidden\"}]}"}}`))
	}))
	defer srv.Close()

	got, err := (&Ollama{URL: srv.URL}).Suggest(context.Background(), Request{Task: "list files"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "ls -la" || len(got.Steps) != 1 {
		t.Fatalf("unexpected suggestion: %+v", got)
	}
}

func TestOllamaBadContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":{"content":"not json"}}`))
	}))
	defer srv.Close()
	if _, err := (&Ollama{URL: srv.URL}).Suggest(context.Background(), Request{}); err == nil {
		t.Fatal("want decode error")
	}
}

func TestParseSuggestionFenced(t *testing.T) {
	in := "Here you go:\n```json\n{\"command\":\"deno init enguy\",\"steps\":[]}\n```"
	got, err := parseSuggestion(in)
	if err != nil || got.Command != "deno init enguy" {
		t.Fatalf("got %+v, err %v", got, err)
	}
}
