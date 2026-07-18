package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	kyverContext "github.com/contact-daniel-me/kyver/internal/context"
)

// mockContextBuilder simulates kyver's context engine for testing
type mockContextBuilder struct{}

func (m *mockContextBuilder) BuildSymbolContext(name string, opts kyverContext.ContextOptions) (*kyverContext.ContextResult, error) {
	return &kyverContext.ContextResult{
		Symbol: &analyzer.Symbol{Name: name, Package: "mockpkg", Kind: "Struct"},
		Documentation: "Mock documentation.",
	}, nil
}
func (m *mockContextBuilder) BuildFileContext(file string, opts kyverContext.ContextOptions) (*kyverContext.ContextResult, error) {
	return nil, nil
}
func (m *mockContextBuilder) BuildPackageContext(pkg string, opts kyverContext.ContextOptions) (*kyverContext.ContextResult, error) {
	return nil, nil
}
func (m *mockContextBuilder) BuildFunctionContext(function string, opts kyverContext.ContextOptions) (*kyverContext.ContextResult, error) {
	return nil, nil
}

func TestMockProviderAsk(t *testing.T) {
	provider := NewMockProvider()
	res, err := provider.Ask(context.Background(), "QUESTION: Explain Engine\nSYMBOL: Engine")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TargetSymbol != "Engine" {
		t.Errorf("expected TargetSymbol Engine, got %s", res.TargetSymbol)
	}
	if !strings.Contains(res.Explanation, "Engine structs are used throughout Kyver") {
		t.Errorf("expected explanation about Engine, got %s", res.Explanation)
	}
}

func TestPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()
	ctxRes := &kyverContext.ContextResult{
		Symbol: &analyzer.Symbol{Name: "Init", Package: "repository", Kind: "Method"},
	}
	
	promptStr, err := pb.Build("Who calls Init?", "Init", ctxRes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if !strings.Contains(promptStr, "QUESTION: Who calls Init?") {
		t.Errorf("missing question in prompt: %s", promptStr)
	}
	if !strings.Contains(promptStr, "SYMBOL: Init") {
		t.Errorf("missing symbol in prompt: %s", promptStr)
	}
	if !strings.Contains(promptStr, `"name": "Init"`) {
		t.Errorf("missing context JSON in prompt: %s", promptStr)
	}
}

func TestEngineIntegration(t *testing.T) {
	provider := NewMockProvider()
	ctxBuild := &mockContextBuilder{}
	engine := NewEngine(provider, ctxBuild)
	
	res, err := engine.Ask("Explain GoAnalyzer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if res.Question != "Explain GoAnalyzer" {
		t.Errorf("expected Question Explain GoAnalyzer, got %s", res.Question)
	}
	if res.TargetSymbol != "GoAnalyzer" {
		t.Errorf("expected TargetSymbol GoAnalyzer, got %s", res.TargetSymbol)
	}
	if !strings.Contains(res.Summary, "GoAnalyzer parses Go code") {
		t.Errorf("expected specific summary for GoAnalyzer, got %s", res.Summary)
	}
}
