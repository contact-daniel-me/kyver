package ai

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	kyverContext "github.com/contact-daniel-me/kyver/internal/context"
)

// Engine encapsulates the AI intelligence layer for Kyver.
type Engine struct {
	provider AIProvider
	ctxBuild kyverContext.ContextBuilder
	promptBuilder *PromptBuilder
}

func NewEngine(provider AIProvider, ctxBuild kyverContext.ContextBuilder) *Engine {
	return &Engine{
		provider:      provider,
		ctxBuild:      ctxBuild,
		promptBuilder: NewPromptBuilder(),
	}
}

// Ask orchestrates the end-to-end flow of an AI question.
func (e *Engine) Ask(query string) (*AskResult, error) {
	start := time.Now()

	// 1. Extract Target Symbol from Question (MVP heuristic)
	// In the future, an LLM call will parse intent or use RAG directly.
	targetSymbol := extractSymbolHeuristic(query)
	if targetSymbol == "" {
		return nil, fmt.Errorf("could not determine target symbol from query. Try asking 'Explain <Symbol>'")
	}

	// 2. Fetch Semantic Context
	// We request a full context to feed to the LLM.
	opts := kyverContext.ContextOptions{
		Full: true,
	}
	ctxRes, err := e.ctxBuild.BuildSymbolContext(targetSymbol, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to gather context for %s: %v", targetSymbol, err)
	}

	// 3. Build Prompt
	promptStr, err := e.promptBuilder.Build(query, targetSymbol, ctxRes)
	if err != nil {
		return nil, err
	}

	// 4. Request AI inference
	res, err := e.provider.Ask(context.Background(), promptStr)
	if err != nil {
		return nil, err
	}
	
	// Ensure the result preserves the target symbol and correctly factors lookup time
	if res.TargetSymbol == "" {
		res.TargetSymbol = targetSymbol
	}
	res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)

	return res, nil
}

// extractSymbolHeuristic is a naive MVP approach to pull the symbol from known prompt formats.
func extractSymbolHeuristic(query string) string {
	q := strings.TrimSpace(query)
	
	// Example: "Explain Engine" -> "Engine"
	if strings.HasPrefix(strings.ToLower(q), "explain ") {
		return strings.TrimSpace(q[8:])
	}
	// Example: "Who calls Init?" -> "Init"
	if strings.HasPrefix(strings.ToLower(q), "who calls ") {
		return strings.TrimRight(strings.TrimSpace(q[10:]), "?")
	}
	// Example: "What does GoAnalyzer do?" -> "GoAnalyzer"
	if strings.HasPrefix(strings.ToLower(q), "what does ") {
		parts := strings.Split(q[10:], " ")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	if strings.Contains(strings.ToLower(q), "indexing implemented") {
		return "Indexer"
	}
	
	// Fallback to searching for capitalized words assuming they are symbols?
	// For MVP, just return the last word if no question mark, or just return the query if simple.
	parts := strings.Split(q, " ")
	for _, p := range parts {
		if len(p) > 0 && p[0] >= 'A' && p[0] <= 'Z' && p != "What" && p != "Who" && p != "How" {
			return strings.TrimRight(p, "?")
		}
	}
	
	// Absolute fallback
	return "Unknown"
}
