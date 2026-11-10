package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekSuggest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("authorization") != "Bearer k" {
			t.Errorf("missing bearer token, got %q", r.Header.Get("authorization"))
		}
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		if req["model"] != DefaultDeepSeekModel {
			t.Errorf("model = %v, want %v", req["model"], DefaultDeepSeekModel)
		}
		rf, _ := req["response_format"].(map[string]any)
		if rf["type"] != "json_object" {
			t.Errorf("response_format = %v, want json_object", req["response_format"])
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"{\"command\":\"ls -la\",\"steps\":[{\"part\":\"-la\",\"explanation\":\"long format, include hidden\"}]}"}}]}`))
	}))
	defer srv.Close()

	d := &DeepSeek{APIKey: "k", URL: srv.URL}
	got, err := d.Suggest(context.Background(), Request{Task: "list files"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "ls -la" || len(got.Steps) != 1 {
		t.Fatalf("unexpected suggestion: %+v", got)
	}
}

func TestDeepSeekExplicitModel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		if req["model"] != "deepseek-v4-pro" {
			t.Errorf("model = %v, want deepseek-v4-pro", req["model"])
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"{\"command\":\"pwd\",\"steps\":[]}"}}]}`))
	}))
	defer srv.Close()

	if _, err := (&DeepSeek{APIKey: "k", Model: "deepseek-v4-pro", URL: srv.URL}).Suggest(context.Background(), Request{}); err != nil {
		t.Fatal(err)
	}
}

func TestDeepSeekErrors(t *testing.T) {
	if _, err := (&DeepSeek{}).Suggest(context.Background(), Request{}); err == nil ||
		!strings.Contains(err.Error(), "DEEPSEEK_API_KEY") {
		t.Fatalf("want missing-key error, got %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad key", http.StatusUnauthorized)
	}))
	defer srv.Close()
	if _, err := (&DeepSeek{APIKey: "k", URL: srv.URL}).Suggest(context.Background(), Request{}); err == nil {
		t.Fatal("want HTTP error")
	}

	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[]}`))
	}))
	defer empty.Close()
	if _, err := (&DeepSeek{APIKey: "k", URL: empty.URL}).Suggest(context.Background(), Request{}); err == nil {
		t.Fatal("want no-choices error")
	}
}
