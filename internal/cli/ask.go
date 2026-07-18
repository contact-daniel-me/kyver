package cli

import (
	"fmt"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/ai"
	"github.com/contact-daniel-me/kyver/internal/context"
)

func runAsk(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver ask \"<question>\"")
	}

	query := strings.Join(args, " ")

	// Instantiate Semantic Context Engine
	ctxEngine, err := context.NewEngine(".")
	if err != nil {
		return err
	}

	// Instantiate AI Provider (Mock for MVP)
	provider := ai.NewMockProvider()

	// Instantiate AI Engine
	aiEngine := ai.NewEngine(provider, ctxEngine)

	// Execute AI Query
	res, err := aiEngine.Ask(query)
	if err != nil {
		return err
	}

	// Format and Print Output
	ai.PrintAskResult(res)

	return nil
}
