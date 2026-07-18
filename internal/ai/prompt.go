package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	kyverContext "github.com/contact-daniel-me/kyver/internal/context"
)

// PromptBuilder constructs the final prompt string to send to the AI provider.
type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// Build combines the user query with the relevant context payload.
func (p *PromptBuilder) Build(query string, targetSymbol string, ctxRes *kyverContext.ContextResult) (string, error) {
	// Serialize the context result into JSON to feed to the LLM
	ctxBytes, err := json.MarshalIndent(ctxRes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize context: %v", err)
	}

	var sb strings.Builder
	
	// System Instructions
	sb.WriteString("You are Kyver AI, an expert semantic code analysis assistant.\n")
	sb.WriteString("Use the provided JSON context to answer the user's question accurately.\n")
	sb.WriteString("Do not guess or hallucinate. Rely strictly on the context provided.\n\n")
	
	// Payload Markers (Helpful for MockProvider heuristics as well)
	sb.WriteString(fmt.Sprintf("QUESTION: %s\n", query))
	sb.WriteString(fmt.Sprintf("SYMBOL: %s\n\n", targetSymbol))
	
	// Context
	sb.WriteString("CONTEXT_JSON:\n")
	sb.WriteString(string(ctxBytes))
	sb.WriteString("\n\n")
	
	// Output Instructions
	sb.WriteString("Return your response in a structured format containing: Summary, Detailed Explanation, Related Files, and Confidence.\n")

	return sb.String(), nil
}
