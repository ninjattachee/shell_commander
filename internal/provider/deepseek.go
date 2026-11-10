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
	DefaultDeepSeekModel = "deepseek-flash"
	defaultDeepSeekURL   = "https://api.deepseek.com/chat/completions"
)

// DeepSeek talks to the DeepSeek chat API over plain HTTP. The API is
// OpenAI-compatible, so a bare chat-completions request is enough.
type DeepSeek struct {
	APIKey string
	Model  string       // defaults to DefaultDeepSeekModel
	URL    string       // defaults to the public API; overridden in tests
	Client *http.Client // defaults to http.DefaultClient
}

func (d *DeepSeek) Suggest(ctx context.Context, r Request) (Suggestion, error) {
	if d.APIKey == "" {
		return Suggestion{}, errors.New("DEEPSEEK_API_KEY is not set: run `shell-commander auth deepseek` or export it")
	}
	model, url, client := d.Model, d.URL, d.Client
	if model == "" {
		model = DefaultDeepSeekModel
	}
	if url == "" {
		url = defaultDeepSeekURL
	}
	if client == nil {
		client = http.DefaultClient
	}

	// JSON mode requires the word "json" in the prompt (jsonInstruction
	// provides it) and a generous token limit so the object isn't cut off.
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt(r) + jsonInstruction()},
			{"role": "user", "content": r.Task},
		},
		"response_format": map[string]any{"type": "json_object"},
		"max_tokens":      2048,
		"stream":          false,
	})
	if err != nil {
		return Suggestion{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return Suggestion{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("authorization", "Bearer "+d.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return Suggestion{}, fmt.Errorf("deepseek request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Suggestion{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return Suggestion{}, fmt.Errorf("deepseek: %s: %s", resp.Status, raw)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return Suggestion{}, fmt.Errorf("decode deepseek response: %w", err)
	}
	if len(out.Choices) == 0 {
		return Suggestion{}, errors.New("deepseek response contained no choices")
	}
	s, err := parseSuggestion(out.Choices[0].Message.Content)
	if err != nil {
		return Suggestion{}, err
	}
	if s.Command == "" {
		return Suggestion{}, errors.New("deepseek returned an empty command")
	}
	return s, nil
}
