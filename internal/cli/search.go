package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/search"
)

func runSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query is required. Usage: kyver search <query> [options]")
	}

	q := search.Query{
		Limit: 0,
	}

	jsonOutput := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") && q.Text == "" {
			q.Text = arg
			continue
		}

		switch arg {
		case "--kind":
			if i+1 < len(args) {
				// Normalize kind to capitalized title case for groups (e.g. struct -> Struct)
				kindStr := strings.ToLower(args[i+1])
				if len(kindStr) > 0 {
					kindStr = strings.ToUpper(kindStr[:1]) + kindStr[1:]
				}
				q.Kind = kindStr
				i++
			}
		case "--package":
			if i+1 < len(args) {
				q.Package = args[i+1]
				i++
			}
		case "--receiver":
			if i+1 < len(args) {
				q.Receiver = args[i+1]
				i++
			}
		case "--language":
			if i+1 < len(args) {
				q.Language = args[i+1]
				i++
			}
		case "--file":
			if i+1 < len(args) {
				q.File = args[i+1]
				i++
			}
		case "--imports":
			if i+1 < len(args) {
				q.Imports = args[i+1]
				i++
			}
		case "--docs":
			if i+1 < len(args) {
				q.Docs = args[i+1]
				i++
			}
		case "--limit":
			if i+1 < len(args) {
				if l, err := strconv.Atoi(args[i+1]); err == nil {
					q.Limit = l
				}
				i++
			}
		case "--exported":
			q.Exported = true
		case "--definition":
			q.Definition = true
		case "--references":
			q.References = true
		case "--symbol":
			if i+1 < len(args) {
				q.Symbol = true
				q.Text = args[i+1]
				i++
			}
		case "--json":
			jsonOutput = true
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !jsonOutput {
		fmt.Println("Searching...")
	}

	start := time.Now()
	repo, err := search.LoadRepository(cwd)
	if err != nil {
		return fmt.Errorf("failed to load semantic index. run 'kyver analyze' first: %w", err)
	}

	engine := search.NewKeywordSearch(repo)

	var results []search.Result
	var related []search.Result
	var references []string

	if q.Definition {
		results, err = engine.Definition(q.Text)
	} else if q.References {
		// If only references are requested, we can still fetch them to populate the output
		references, err = engine.References(q.Text)
		// Fetch definition to show what we are referencing
		results, _ = engine.Definition(q.Text) 
	} else {
		results, err = engine.Search(q)
		if err == nil && len(results) > 0 && q.Text != "" {
			// Enrich standard search with top result context
			topName := results[0].Name
			related, _ = engine.Related(topName)
			references, _ = engine.References(topName)
		}
	}

	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if jsonOutput {
		return search.PrintJSON(q, results, related, references, time.Since(start))
	}

	search.PrintPlain(q, results, related, references, time.Since(start))

	return nil
}
