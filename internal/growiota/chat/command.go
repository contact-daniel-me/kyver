package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// RunChat executes the AI Chat Engine logic (Growiota Premium module).
func RunChat(args []string) error {
	var question string
	var stream bool
	var jsonOutput bool
	provider := "mock"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--stream":
			stream = true
		case "--json":
			jsonOutput = true
		case "--provider":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		default:
			if question == "" && args[i][:2] != "--" {
				question = args[i]
			}
		}
	}

	engine, err := NewEngine(".", provider)
	if err != nil {
		return fmt.Errorf("failed to initialize chat engine: %w", err)
	}

	ctx := context.Background()

	// If no specific question provided, start interactive session
	if question == "" {
		if jsonOutput {
			return fmt.Errorf("cannot use --json with interactive session")
		}
		return engine.RunInteractiveSession(ctx, stream)
	}

	// Single shot query
	var streamChan chan string
	if stream && !jsonOutput {
		streamChan = make(chan string)
		go func() {
			for chunk := range streamChan {
				fmt.Print(chunk)
			}
		}()
	}

	resp, err := engine.Ask(ctx, question, streamChan)
	if err != nil {
		return err
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(resp)
	} else if !stream {
		fmt.Printf("\n%s\n", resp.Content)
	} else {
		fmt.Println()
	}

	return nil
}
