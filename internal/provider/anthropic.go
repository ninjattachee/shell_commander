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
	DefaultAnthropicModel = "claude-sonnet-5"
	defaultAnthropicURL   = "https://api.anthropic.com/v1/messages"
	anthropicVersion      = "2023-06-01"
)

// Anthropic talks to the Claude Messages API over plain HTTP.
type Anthropic struct {
	APIKey string
	Model  string       // defaults to DefaultAnthropicModel
	URL    string       // defaults to the public API; overridden in tests
	Client *http.Client // defaults to http.DefaultClient
}

func (a *Anthropic) Suggest(ctx context.Context, r Request) (Suggestion, error) {
	if a.APIKey == "" {
		return Suggestion{}, errors.New("ANTHROPIC_API_KEY is not set")
	}
	model, url, client := a.Model, a.URL, a.Client
	if model == "" {
		model = DefaultAnthropicModel
	}
	if url == "" {
		url = defaultAnthropicURL
	}
	if client == nil {
		client = http.DefaultClient
	}

	// Force Claude to "call" our one tool so the reply is JSON matching schema.
	body, err := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": 1024,
		"system":     systemPrompt(r),
		"messages":   []map[string]any{{"role": "user", "content": r.Task}},
		"tools": []map[string]any{{
			"name":         toolName,
			"description":  "Report the suggested shell command and its explanation.",
			"input_schema": schema,
		}},
		"tool_choice": map[string]any{"type": "tool", "name": toolName},
	})
	if err != nil {
		return Suggestion{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Suggestion{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	resp, err := client.Do(httpReq)
	if err != nil {
		return Suggestion{}, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Suggestion{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Suggestion{}, fmt.Errorf("anthropic: %s: %s", resp.Status, raw)
	}

	var out struct {
		Content []struct {
			Type  string          `json:"type"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return Suggestion{}, fmt.Errorf("decode anthropic response: %w", err)
	}
	for _, c := range out.Content {
		if c.Type != "tool_use" {
			continue
		}
		var s Suggestion
		if err := json.Unmarshal(c.Input, &s); err != nil {
			return Suggestion{}, fmt.Errorf("decode suggestion: %w", err)
		}
		if s.Command == "" {
			return Suggestion{}, errors.New("anthropic returned an empty command")
		}
		return s, nil
	}
	return Suggestion{}, errors.New("anthropic response contained no tool_use block")
}
