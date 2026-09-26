package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

// toolName is the name of the single "tool" we force Anthropic to call so
// that its reply is always JSON matching schema.
const toolName = "suggest_command"

// schema is the JSON Schema for Suggestion. Both providers use it to force
// structured output instead of parsing free-form text.
var schema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"command": map[string]any{
			"type":        "string",
			"description": "A single shell command line. Chain with && if several steps are needed.",
		},
		"steps": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"part":        map[string]any{"type": "string"},
					"explanation": map[string]any{"type": "string"},
				},
				"required": []string{"part", "explanation"},
			},
		},
		"warning": map[string]any{
			"type":        "string",
			"description": "Only set if the command deletes data, overwrites files, needs root, or is otherwise hard to undo.",
		},
	},
	"required": []string{"command", "steps"},
}

// systemPrompt tells the model its job and the environment it is targeting.
func systemPrompt(r Request) string {
	var b strings.Builder
	b.WriteString("You suggest shell commands and explain them to someone learning the shell.\n")
	b.WriteString("Return exactly one command line that accomplishes the task. ")
	b.WriteString("Then break it into steps: each step quotes one part of the command and explains what it does and why it is needed, in plain language.\n")
	b.WriteString("Never just restate the part (e.g. \"the init part\"); say what the tool does. Example: part \"ls -la\" -> \"lists files in long format, including hidden ones\".\n")
	b.WriteString("Set `warning` only for destructive, irreversible, or privileged commands.\n")
	b.WriteString("Prefer common, widely available tools and never invent flags.\n")
	fmt.Fprintf(&b, "Environment: OS=%s, shell=%s, cwd=%s.\n", r.OS, r.Shell, r.Cwd)
	return b.String()
}

// parseSuggestion decodes model text into a Suggestion. Some models ignore
// the JSON-only constraint and wrap the object in a markdown fence or add
// prose, so we decode from the first '{' to the last '}'.
func parseSuggestion(text string) (Suggestion, error) {
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return Suggestion{}, fmt.Errorf("no JSON object in model output: %q", text)
	}
	var s Suggestion
	if err := json.Unmarshal([]byte(text[start:end+1]), &s); err != nil {
		return Suggestion{}, fmt.Errorf("decode suggestion: %w", err)
	}
	return s, nil
}

// jsonInstruction is appended for providers that can't be forced to return
// structured output (Ollama's `format` is ignored by some models).
func jsonInstruction() string {
	sch, _ := json.Marshal(schema)
	return "Respond with ONLY a JSON object matching this JSON Schema. No markdown, no prose outside the JSON:\n" + string(sch) + "\n"
}
