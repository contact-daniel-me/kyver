package chat

import (
	"context"
	"fmt"
	"os"
)

type ChatRequest struct {
	Messages []Message
	Stream   bool
}

type ChatResponse struct {
	Content string
}

type LLMProvider interface {
	Chat(ctx context.Context, request ChatRequest, streamChan chan<- string) (*ChatResponse, error)
	Name() string
}

// MockProvider provides dummy responses for testing without an API key
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Chat(ctx context.Context, request ChatRequest, streamChan chan<- string) (*ChatResponse, error) {
	resp := "This is a mock response from the LLM based on the retrieved context."
	if request.Stream && streamChan != nil {
		streamChan <- resp
		close(streamChan)
	}
	return &ChatResponse{Content: resp}, nil
}

// OpenAIProvider is a skeleton for the OpenAI API
type OpenAIProvider struct {
	APIKey string
}

func NewOpenAIProvider() (*OpenAIProvider, error) {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is missing")
	}
	return &OpenAIProvider{APIKey: key}, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Chat(ctx context.Context, request ChatRequest, streamChan chan<- string) (*ChatResponse, error) {
	// TODO: Implement actual HTTP call to OpenAI API here.
	// For now, return a placeholder to satisfy the adapter skeleton.
	resp := "[OpenAI] Implementation pending API integration."
	if request.Stream && streamChan != nil {
		streamChan <- resp
		close(streamChan)
	}
	return &ChatResponse{Content: resp}, nil
}
