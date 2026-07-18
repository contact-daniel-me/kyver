package chat

import (
	"context"

	"github.com/contact-daniel-me/kyver/internal/retrieval"
)

type Engine struct {
	retrievalEngine *retrieval.Engine
	provider        LLMProvider
	history         *History
}

func NewEngine(rootDir string, providerName string) (*Engine, error) {
	retEngine, err := retrieval.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}

	var provider LLMProvider
	if providerName == "openai" {
		p, err := NewOpenAIProvider()
		if err != nil {
			return nil, err
		}
		provider = p
	} else {
		// default to mock provider for tests and safe defaults
		provider = NewMockProvider()
	}

	return &Engine{
		retrievalEngine: retEngine,
		provider:        provider,
		history:         NewHistory(),
	}, nil
}

func (e *Engine) Ask(ctx context.Context, question string, streamChan chan<- string) (*ChatResponse, error) {
	// 1. Retrieve Context
	query := retrieval.Query{Text: question}
	bundle, err := e.retrievalEngine.Retrieve(query)
	if err != nil {
		return nil, err
	}

	// 2. Build System Prompt if this is the first interaction
	if len(e.history.Messages) == 0 {
		sysPrompt := BuildSystemPrompt(bundle)
		e.history.AddMessage(RoleSystem, sysPrompt)
	} else {
		// Optionally, we could inject the new retrieval bundle as a system message
		// to update context for the specific question.
		sysPrompt := BuildSystemPrompt(bundle)
		e.history.AddMessage(RoleSystem, "Context update for new question:\n"+sysPrompt)
	}

	// 3. Add User Question
	e.history.AddMessage(RoleUser, question)

	// 4. Build Chat Request
	req := ChatRequest{
		Messages: e.history.Messages,
		Stream:   streamChan != nil,
	}

	// 5. Call LLM
	resp, err := e.provider.Chat(ctx, req, streamChan)
	if err != nil {
		return nil, err
	}

	// 6. Record Assistant Response
	e.history.AddMessage(RoleAssistant, resp.Content)

	return resp, nil
}
