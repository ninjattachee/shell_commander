package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	DefaultOllamaModel = "gemma4:31b-cloud"
	defaultOllamaURL   = "http://localhost:11434/api/chat"
)

// Ollama talks to a local Ollama server. No API key is needed.
type Ollama struct {
	Model  string       // defaults to DefaultOllamaModel
	URL    string       // defaults to the local server; overridden in tests
	Client *http.Client // defaults to http.DefaultClient
}

func (o *Ollama) Suggest(ctx context.Context, r Request) (Suggestion, error) {
	model, url, client := o.Model, o.URL, o.Client
	if model == "" {
		model = DefaultOllamaModel
	}
	if url == "" {
		url = defaultOllamaURL
	}
	if client == nil {
		client = http.DefaultClient
	}

	// "format" constrains the model's output to our JSON Schema.
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt(r) + jsonInstruction()},
			{"role": "user", "content": r.Task},
		},
		"format": schema,
		"stream": false,
	})
	if err != nil {
		return Suggestion{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Suggestion{}, err
	}
	httpReq.Header.Set("content-type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return Suggestion{}, fmt.Errorf("ollama request (is the server running?): %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Suggestion{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Suggestion{}, fmt.Errorf("ollama: %s: %s", resp.Status, raw)
	}

	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return Suggestion{}, fmt.Errorf("decode ollama response: %w", err)
	}
	s, err := parseSuggestion(out.Message.Content)
	if err != nil {
		return Suggestion{}, err
	}
	if s.Command == "" {
		return Suggestion{}, errors.New("ollama returned an empty command")
	}
	return s, nil
}
