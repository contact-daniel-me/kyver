package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// promptSymbolChoice reads an interactive 1-based index from stdin.
// Supports q, quit, exit, and treats EOF / Ctrl+C (Scanln error) as cancel.
func promptSymbolChoice(n int) (int, error) {
	for {
		fmt.Printf("\nEnter choice (1-%d, q to cancel): ", n)
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return 0, fmt.Errorf("cancelled")
		}

		input = strings.TrimSpace(strings.ToLower(input))
		if input == "q" || input == "quit" || input == "exit" {
			return 0, fmt.Errorf("cancelled")
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > n {
			fmt.Println("Invalid choice. Try again.")
			continue
		}
		return choice, nil
	}
}
