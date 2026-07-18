package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/contact-daniel-me/kyver/internal/retrieval"
)

func runRetrieve(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver retrieve <question> [--symbol <sym>] [--package <pkg>] [--file <file>] [--json] [--verbose]")
	}

	query := retrieval.Query{}

	// If the first argument doesn't start with --, treat it as the question
	if args[0][:2] != "--" {
		query.Text = args[0]
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			query.JSON = true
		case "--verbose":
			query.Verbose = true
		case "--symbol":
			if i+1 < len(args) {
				query.Symbol = args[i+1]
				i++
			}
		case "--package":
			if i+1 < len(args) {
				query.Package = args[i+1]
				i++
			}
		case "--file":
			if i+1 < len(args) {
				query.File = args[i+1]
				i++
			}
		}
	}

	engine, err := retrieval.NewEngine(".")
	if err != nil {
		return err
	}

	bundle, err := engine.Retrieve(query)
	if err != nil {
		return err
	}

	// Save bundle
	err = engine.SaveBundle(bundle)
	if err != nil {
		fmt.Printf("Warning: failed to save retrieval bundle: %v\n", err)
	}

	if query.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.Encode(bundle)
		return nil
	}

	fmt.Printf("=== Retrieval Results ===\n\n")
	fmt.Printf("Confidence Score: %.2f\n", bundle.Confidence)
	fmt.Printf("Retrieval Time: %s\n\n", bundle.RetrievalTime)

	if len(bundle.TopSymbols) > 0 {
		fmt.Println("Top Symbols:")
		for i, s := range bundle.TopSymbols {
			if i >= 5 {
				fmt.Println("- ...")
				break
			}
			fmt.Printf("- %s (%s)\n", s.Name, s.Kind)
		}
		fmt.Println()
	}

	if len(bundle.TopPackages) > 0 {
		fmt.Println("Top Packages:")
		for _, p := range bundle.TopPackages {
			fmt.Printf("- %s\n", p)
		}
		fmt.Println()
	}

	if len(bundle.TopFiles) > 0 {
		fmt.Println("Top Files:")
		for i, f := range bundle.TopFiles {
			if i >= 5 {
				fmt.Println("- ...")
				break
			}
			fmt.Printf("- %s\n", f)
		}
		fmt.Println()
	}
	
	if len(bundle.TopContexts) > 0 {
		fmt.Printf("Context bundles generated: %d\n", len(bundle.TopContexts))
	}

	return nil
}
