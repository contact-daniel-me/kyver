package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/graph"
)

func runCallers(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver callers <symbol> [--json] [--tree] [--max <n>] [--package <pkg>] [--sort <line|name>] [--select <index>] [--interactive]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.CallersOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--tree":
			opts.TreeFormat = true
		case "--interactive":
			interactive = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.Callers(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintCallersResult(res, opts, jsonOutput)
		for {
			fmt.Print("\nEnter choice: ")
			var choice int
			_, err := fmt.Scanln(&choice)
			if err != nil || choice < 1 || choice > len(res.Candidates) {
				fmt.Println("Invalid choice. Try again.")
				continue
			}
			
			opts.SelectIndex = choice
			res, err = engine.Callers(symbol, opts)
			if err != nil {
				return err
			}
			break
		}
	}

	graph.PrintCallersResult(res, opts, jsonOutput)
	return nil
}

func runCallees(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver callees <symbol> [--json] [--tree] [--max <n>] [--package <pkg>] [--sort <line|name>] [--select <index>] [--interactive]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.CalleesOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--tree":
			opts.TreeFormat = true
		case "--interactive":
			interactive = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.Callees(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintCalleesResult(res, opts, jsonOutput)
		for {
			fmt.Printf("\nEnter choice (1-%d, q to cancel): ", len(res.Candidates))
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				return fmt.Errorf("cancelled")
			}
			
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "q" || input == "quit" || input == "exit" {
				return fmt.Errorf("cancelled")
			}
			
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(res.Candidates) {
				fmt.Println("Invalid choice. Try again.")
				continue
			}
			
			opts.SelectIndex = choice
			res, err = engine.Callees(symbol, opts)
			if err != nil {
				return err
			}
			break
		}
	}

	graph.PrintCalleesResult(res, opts, jsonOutput)
	return nil
}

func runReferences(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver references <symbol> [--json] [--tree] [--max <n>] [--package <pkg>] [--file <file>] [--sort <line|name>] [--select <index>] [--interactive]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.ReferenceOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--tree":
			opts.TreeFormat = true
		case "--interactive":
			interactive = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--file":
			if i+1 < len(args) {
				opts.FileFilter = args[i+1]
				i++
			}
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.References(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintReferenceResult(res, opts, jsonOutput)
		for {
			fmt.Printf("\nEnter choice (1-%d, q to cancel): ", len(res.Candidates))
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				return fmt.Errorf("cancelled")
			}
			
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "q" || input == "quit" || input == "exit" {
				return fmt.Errorf("cancelled")
			}
			
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(res.Candidates) {
				fmt.Println("Invalid choice. Try again.")
				continue
			}
			
			opts.SelectIndex = choice
			res, err = engine.References(symbol, opts)
			if err != nil {
				return err
			}
			break
		}
	}

	graph.PrintReferenceResult(res, opts, jsonOutput)
	return nil
}

func runRefs(args []string) error {
	return runReferences(args)
}

func runImpacts(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver impacts <symbol> [--json] [--tree] [--max <n>] [--package <pkg>] [--sort <line|name>] [--select <index>] [--interactive]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.ImpactOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--interactive":
			interactive = true
		case "--tree":
			opts.TreeFormat = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}
	opts.Interactive = interactive

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.Impacts(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintImpactResult(res, opts, jsonOutput)
		choice, err := promptSymbolChoice(len(res.Candidates))
		if err != nil {
			return err
		}
		opts.SelectIndex = choice
		res, err = engine.Impacts(symbol, opts)
		if err != nil {
			return err
		}
	}

	graph.PrintImpactResult(res, opts, jsonOutput)
	return nil
}

func runImplements(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver implements <interface> [--tree] [--json] [--interactive] [--select <index>] [--package <pkg>] [--sort <line|name>] [--max <n>]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.ImplementsOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--interactive":
			interactive = true
		case "--tree":
			opts.TreeFormat = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}
	opts.Interactive = interactive

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.Implements(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintImplementsResult(res, opts, jsonOutput)
		choice, err := promptSymbolChoice(len(res.Candidates))
		if err != nil {
			return err
		}
		opts.SelectIndex = choice
		res, err = engine.Implements(symbol, opts)
		if err != nil {
			return err
		}
	}

	graph.PrintImplementsResult(res, opts, jsonOutput)
	return nil
}

func runGraph(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver graph <symbol> [--package <pkg>] [--file <file>] [--json] [--interactive] [--select <index>] [--tree]")
	}

	var symbol, pkg, file string
	jsonOutput := false
	interactive := false
	treeFormat := false
	selectIndex := 0

	// graph can be called with --package or --file without a symbol
	if args[0] != "--package" && args[0] != "--file" && args[0] != "--json" && args[0] != "--interactive" && args[0] != "--select" && args[0] != "--tree" {
		symbol = args[0]
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--interactive":
			interactive = true
		case "--tree":
			treeFormat = true
		case "--package":
			if i+1 < len(args) {
				pkg = args[i+1]
				i++
			}
		case "--file":
			if i+1 < len(args) {
				file = args[i+1]
				i++
			}
		case "--select":
			if i+1 < len(args) {
				idx, err := strconv.Atoi(args[i+1])
				if err == nil {
					selectIndex = idx
				}
				i++
			}
		}
	}

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	start := time.Now()

	if pkg != "" {
		res, err := engine.PackageGraph(pkg)
		if err != nil {
			return err
		}
		graph.PrintPackageGraph(res, jsonOutput)
	} else if file != "" {
		res, err := engine.FileGraph(file)
		if err != nil {
			return err
		}
		graph.PrintFileGraph(res, jsonOutput)
	} else {
		opts := graph.GraphOptions{
			SelectIndex: selectIndex,
			TreeFormat:  treeFormat,
			Interactive: interactive,
		}
		res, err := engine.Graph(symbol, opts)
		if err != nil {
			return err
		}

		if res.MultipleFound && interactive && !jsonOutput {
			graph.PrintGraph(res, opts, jsonOutput)
			for {
				fmt.Printf("\nEnter choice (1-%d, q to cancel): ", len(res.Candidates))
				var input string
				_, err := fmt.Scanln(&input)
				if err != nil {
					return fmt.Errorf("cancelled")
				}
				
				input = strings.TrimSpace(strings.ToLower(input))
				if input == "q" || input == "quit" || input == "exit" {
					return fmt.Errorf("cancelled")
				}
				
				choice, err := strconv.Atoi(input)
				if err != nil || choice < 1 || choice > len(res.Candidates) {
					fmt.Println("Invalid choice. Try again.")
					continue
				}
				
				opts.SelectIndex = choice
				res, err = engine.Graph(symbol, opts)
				if err != nil {
					return err
				}
				break
			}
		}

		graph.PrintGraph(res, opts, jsonOutput)
	}

	if !jsonOutput && !interactive && symbol != "" && pkg == "" && file == "" {
		// Outputting completed in ... not needed, Graph stats outputs it natively if we follow callers/callees.
		// Wait, if it's package or file, it prints completed. For graph it should use the stats time like callers/callees.
	} else if !jsonOutput {
		if pkg != "" || file != "" {
			fmt.Printf("\nCompleted in %v\n", time.Since(start))
		}
	}
	return nil
}

func runInterface(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver interface <struct> [--tree] [--json] [--interactive] [--select <index>] [--package <pkg>] [--sort <line|name>] [--max <n>]")
	}

	symbol := args[0]
	jsonOutput := false
	interactive := false
	opts := graph.InterfaceOptions{
		SortBy: "line",
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--interactive":
			interactive = true
		case "--tree":
			opts.TreeFormat = true
		case "--max":
			if i+1 < len(args) {
				if max, err := strconv.Atoi(args[i+1]); err == nil {
					opts.Max = max
				}
				i++
			}
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
		case "--sort":
			if i+1 < len(args) {
				opts.SortBy = args[i+1]
				i++
			}
		}
	}
	opts.Interactive = interactive

	engine, err := graph.NewEngine(".")
	if err != nil {
		return err
	}

	res, err := engine.Interface(symbol, opts)
	if err != nil {
		return err
	}

	if res.MultipleFound && interactive && !jsonOutput {
		graph.PrintInterfaceResult(res, opts, jsonOutput)
		choice, err := promptSymbolChoice(len(res.Candidates))
		if err != nil {
			return err
		}
		opts.SelectIndex = choice
		res, err = engine.Interface(symbol, opts)
		if err != nil {
			return err
		}
	}

	graph.PrintInterfaceResult(res, opts, jsonOutput)
	return nil
}
