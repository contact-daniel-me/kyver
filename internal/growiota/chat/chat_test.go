package chat

import (
	"context"
	"testing"
)

func TestMockProvider(t *testing.T) {
	provider := NewMockProvider()
	
	if provider.Name() != "mock" {
		t.Errorf("Expected provider name 'mock', got %s", provider.Name())
	}

	req := ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "Hello"}},
		Stream:   false,
	}

	resp, err := provider.Chat(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Content == "" {
		t.Errorf("Expected non-empty response content")
	}
}

func TestHistory(t *testing.T) {
	h := NewHistory()
	h.AddMessage(RoleSystem, "System prompt")
	h.AddMessage(RoleUser, "User prompt")

	if len(h.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(h.Messages))
	}

	if h.Messages[0].Role != RoleSystem {
		t.Errorf("Expected System role")
	}
	if h.Messages[1].Role != RoleUser {
		t.Errorf("Expected User role")
	}
}

func TestPromptBuilder(t *testing.T) {
	// Test nil bundle doesn't panic
	prompt := BuildSystemPrompt(nil)
	if prompt == "" {
		t.Errorf("Expected fallback system prompt for nil bundle")
	}
}
