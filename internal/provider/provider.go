// Package provider defines how shell-commander asks an LLM for a command.
package provider

import "context"

// Request is everything the model needs to know to suggest a command.
type Request struct {
	Task  string // the user's natural-language request
	OS    string // e.g. "linux"
	Shell string // e.g. "/bin/zsh"
	Cwd   string // current working directory
}

// Step explains one piece of the suggested command.
type Step struct {
	Part        string `json:"part"`        // a fragment of the command, e.g. "mkdir enguy"
	Explanation string `json:"explanation"` // what that fragment does
}

// Suggestion is the model's structured answer.
type Suggestion struct {
	Command string `json:"command"`
	Steps   []Step `json:"steps"`
	Warning string `json:"warning,omitempty"` // non-empty for destructive/irreversible commands
}

// Provider turns a Request into a Suggestion. Implementations hide the
// details of a particular LLM API.
type Provider interface {
	Suggest(ctx context.Context, req Request) (Suggestion, error)
}
