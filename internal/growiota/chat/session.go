package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// RunInteractiveSession starts a REPL for conversational chat.
func (e *Engine) RunInteractiveSession(ctx context.Context, stream bool) error {
	fmt.Println("Kyver AI Chat Session Started. Type 'exit' or 'quit' to end.")
	fmt.Println("---------------------------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		var streamChan chan string
		if stream {
			streamChan = make(chan string)
			go printStream(streamChan)
		}

		resp, err := e.Ask(ctx, input, streamChan)
		if err != nil {
			fmt.Printf("\n[Error] %v\n", err)
			continue
		}

		if !stream {
			fmt.Printf("\nKyver: %s\n", resp.Content)
		} else {
			// Print a newline after the stream completes
			fmt.Println()
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func printStream(streamChan <-chan string) {
	fmt.Print("\nKyver: ")
	for chunk := range streamChan {
		fmt.Print(chunk)
	}
}
