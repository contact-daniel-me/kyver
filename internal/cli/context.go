package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/context"
)

func runContext(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver context <symbol> [--summary] [--full] [--sections <secs>] [--tree] [--json] [--interactive] [--select <index>] [--package <pkg>]")
	}

	symbol := args[0]
	opts := context.ContextOptions{}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			opts.JSONOutput = true
		case "--tree":
			opts.TreeFormat = true
		case "--interactive":
			opts.Interactive = true
		case "--summary":
			opts.Summary = true
		case "--full":
			opts.Full = true
		case "--select":
			if i+1 < len(args) {
				if idx, err := strconv.Atoi(args[i+1]); err == nil {
					opts.SelectIndex = idx
				}
				i++
			}
		case "--package":
			if i+1 < len(args) {
				opts.PackageFilter = args[i+1]
				i++
			}
		case "--sections":
			if i+1 < len(args) {
				sections := strings.Split(args[i+1], ",")
				for _, s := range sections {
					s = strings.TrimSpace(s)
					if s != "" {
						valid := false
						for _, v := range []string{"callers", "callees", "references", "graph", "impacts", "implements", "interface", "methods", "documentation", "package"} {
							if strings.EqualFold(s, v) {
								valid = true
								break
							}
						}
						if !valid {
							return fmt.Errorf("Unknown section %q", s)
						}
						opts.Sections = append(opts.Sections, s)
					}
				}
				i++
			}
		}
	}

	engine, err := context.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.BuildSymbolContext(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && opts.Interactive && !opts.JSONOutput {
		context.PrintContextResult(res, opts)
		choice, err := promptSymbolChoice(len(res.Candidates))
		if err != nil {
			return err
		}
		
		opts.SelectIndex = choice
		res, err = engine.BuildSymbolContext(symbol, opts)
		if err != nil {
			return err
		}
	}

	context.PrintContextResult(res, opts)
	return nil
}
