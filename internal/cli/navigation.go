package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/contact-daniel-me/kyver/internal/navigation"
)

func runGoto(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver goto <symbol> [--select <index>] [--json] [--open]")
	}

	symbol := args[0]
	jsonOutput := false
	open := false
	interactive := false
	selectIdx := 0

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--open":
			open = true
		case "--interactive":
			interactive = true
		case "--select":
			if i+1 < len(args) {
				idx, err := strconv.Atoi(args[i+1])
				if err == nil {
					selectIdx = idx
				}
				i++
			}
		}
	}

	engine, err := navigation.NewEngine(".")
	if err != nil {
		return err
	}

	start := time.Now()
	res, err := engine.Goto(symbol, selectIdx)
	if err != nil {
		return err
	}
	
	if res.MultipleFound && interactive && !jsonOutput {
		navigation.PrintGoto(res, open, jsonOutput)
		for {
			fmt.Print("\nEnter choice: ")
			var choice int
			_, err := fmt.Scanln(&choice)
			if err != nil || choice < 1 || choice > len(res.Candidates) {
				fmt.Println("Invalid choice. Try again.")
				// flush stdin if needed, though Scanln handles newline
				continue
			}
			
			// perform goto with choice
			res, err = engine.Goto(symbol, choice)
			if err != nil {
				return err
			}
			break
		}
	}

	navigation.PrintGoto(res, open, jsonOutput)
	if !jsonOutput && !open {
		d := time.Since(start)
		if d.Microseconds() < 1000 {
			fmt.Printf("\nCompleted in %d µs\n", d.Microseconds())
		} else {
			ms := float64(d.Microseconds()) / 1000.0
			fmt.Printf("\nCompleted in %.2f ms\n", ms)
		}
	}
	return nil
}

func runPeek(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver peek <symbol> [--select <index>] [--lines <limit>] [--json]")
	}

	symbol := args[0]
	jsonOutput := false
	selectIdx := 0
	linesLimit := 40

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--select":
			if i+1 < len(args) {
				idx, err := strconv.Atoi(args[i+1])
				if err == nil {
					selectIdx = idx
				}
				i++
			}
		case "--lines":
			if i+1 < len(args) {
				limit, err := strconv.Atoi(args[i+1])
				if err == nil {
					linesLimit = limit
				}
				i++
			}
		}
	}

	engine, err := navigation.NewEngine(".")
	if err != nil {
		return err
	}

	start := time.Now()
	res, err := engine.Peek(symbol, selectIdx, linesLimit)
	if err != nil {
		return err
	}

	navigation.PrintPeek(res, jsonOutput)
	if !jsonOutput {
		d := time.Since(start)
		ms := float64(d.Microseconds()) / 1000.0
		fmt.Printf("\nLookup Time : %.2f ms\n", ms)
	}
	return nil
}

func runOutline(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver outline <file> [--json] [--filter <string>] [--sort <line|name>]")
	}

	file := args[0]
	jsonOutput := false
	filter := ""
	sortBy := "line"

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--filter":
			if i+1 < len(args) {
				filter = args[i+1]
				i++
			}
		case "--sort":
			if i+1 < len(args) {
				sortBy = args[i+1]
				i++
			}
		}
	}

	engine, err := navigation.NewEngine(".")
	if err != nil {
		return err
	}

	start := time.Now()
	res, err := engine.Outline(file, filter, sortBy)
	if err != nil {
		return err
	}

	navigation.PrintOutline(res, jsonOutput)
	if !jsonOutput {
		d := time.Since(start)
		if d.Microseconds() < 1000 {
			fmt.Printf("\nLookup Time : %d µs\n", d.Microseconds())
		} else {
			ms := float64(d.Microseconds()) / 1000.0
			fmt.Printf("\nLookup Time : %.2f ms\n", ms)
		}
	}
	return nil
}

func runHierarchy(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: kyver hierarchy <symbol> [--select <index>] [--json]")
	}

	symbol := args[0]
	jsonOutput := false
	selectIdx := 0

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		case "--select":
			if i+1 < len(args) {
				idx, err := strconv.Atoi(args[i+1])
				if err == nil {
					selectIdx = idx
				}
				i++
			}
		}
	}

	engine, err := navigation.NewEngine(".")
	if err != nil {
		return err
	}

	start := time.Now()
	res, err := engine.Hierarchy(symbol, selectIdx)
	if err != nil {
		return err
	}

	navigation.PrintHierarchy(res, jsonOutput)
	if !jsonOutput {
		fmt.Printf("\nCompleted in %v\n", time.Since(start))
	}
	return nil
}
