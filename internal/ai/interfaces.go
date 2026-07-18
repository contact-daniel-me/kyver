package ai

import (
	"context"
)

// AskResult represents the structured output of an AI query.
type AskResult struct {
	Question     string  `json:"question"`
	TargetSymbol string  `json:"targetSymbol"`
	Summary      string  `json:"summary"`
	Explanation  string  `json:"explanation"`
	RelatedFiles []string `json:"relatedFiles,omitempty"`
	Confidence   string  `json:"confidence"`
	LookupTimeMs float64 `json:"lookupTimeMs"`
}

// AIProvider abstracts the underlying AI models (Gemini, Claude, Ollama, etc.)
type AIProvider interface {
	Ask(ctx context.Context, prompt string) (*AskResult, error)
}
