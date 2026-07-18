package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/status"
	"github.com/contact-daniel-me/kyver/internal/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func runAnalyze(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	verbose := false
	targetLanguage := ""
	jsonOutput := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--verbose" {
			verbose = true
		} else if arg == "--json" {
			jsonOutput = true
		} else if arg == "--language" && i+1 < len(args) {
			targetLanguage = args[i+1]
			i++ // skip next
		}
	}

	if !jsonOutput {
		fmt.Println("Analyzing repository...")
		fmt.Println()
	}

	indexPath := filepath.Join(cwd, ".kyver", "index.json")
	indexData, err := status.ReadIndex(indexPath)
	if err != nil || indexData == nil {
		return fmt.Errorf("failed to read repository index. run 'kyver index' first")
	}

	if !jsonOutput {
		fmt.Println("✓ Loading index")
		if verbose {
			fmt.Printf("Found %d files in index\n", len(indexData.Files))
		}
	}

	engine := analyzer.NewEngine(cwd, status.KyverVersion)

	if !jsonOutput {
		fmt.Println("✓ Parsing symbols")
		fmt.Println("✓ Building dependency graph")
		fmt.Println("✓ Computing architecture")
		fmt.Println("✓ Computing metrics")
		fmt.Println("✓ Writing output")
	}

	summary, err := engine.Analyze(indexData, targetLanguage, verbose)
	if err != nil {
		if errors.Is(err, analyzer.ErrInvalidIndex) {
			return fmt.Errorf("invalid or missing index: please run 'kyver index' first")
		}
		if errors.Is(err, analyzer.ErrMissingRepo) {
			return fmt.Errorf("repository missing or inaccessible")
		}
		if errors.Is(err, analyzer.ErrUnsupportedLang) {
			return fmt.Errorf("the specified language is not supported for analysis")
		}
		return fmt.Errorf("analysis failed: %w", err)
	}

	err = validateArtifacts(cwd)
	
	if !jsonOutput {
		printAnalyzeSummary(cwd, summary, err)
	} else {
		if verbose {
			fmt.Printf("{\"status\":\"success\", \"duration\":\"%v\", \"symbols\":%d}\n", summary.Duration, summary.TotalSymbols)
		}
	}

	if err != nil {
		return fmt.Errorf("validation of generated artifacts failed: %w", err)
	}

	return nil
}

func printAnalyzeSummary(cwd string, summary analyzer.AnalysisSummary, validationErr error) {
	p := message.NewPrinter(language.English)
	repoName := filepath.Base(cwd)

	fmt.Println("\nAnalysis Summary")
	fmt.Println()
	fmt.Printf("%-20s: %s\n", "Repository", repoName)
	fmt.Printf("%-20s: %s\n", "Repository Root", cwd)
	p.Printf("%-20s: %d\n", "Packages", summary.Packages)
	p.Printf("%-20s: %d\n", "Indexed Files", summary.TotalIndexedFiles)
	p.Printf("%-20s: %d\n", "Analyzed Files", summary.FilesAnalyzed)
	
	langStr := ""
	for lang, count := range summary.Languages {
		if langStr != "" {
			langStr += ", "
		}
		langStr += fmt.Sprintf("%s (%d)", lang, count)
	}
	fmt.Printf("%-20s: %s\n", "Languages", langStr)

	fmt.Println("\nSemantic Model")
	fmt.Println()
	p.Printf("%-20s: %d\n", "Symbols", summary.TotalSymbols)
	p.Printf("%-20s: %d\n", "Exported Symbols", summary.ExportedSymbols)
	p.Printf("%-20s: %d\n", "Functions", summary.Functions)
	p.Printf("%-20s: %d\n", "Methods", summary.Methods)
	p.Printf("%-20s: %d\n", "Structs", summary.Structs)
	p.Printf("%-20s: %d\n", "Interfaces", summary.Interfaces)
	p.Printf("%-20s: %d\n", "Constants", summary.Constants)
	p.Printf("%-20s: %d\n", "Variables", summary.Variables)
	p.Printf("%-20s: %d\n", "Type Aliases", summary.Aliases)

	fmt.Println("\nDependency Graph")
	fmt.Println()
	p.Printf("%-20s: %d\n", "Packages", summary.Packages)
	p.Printf("%-20s: %d\n", "Edges", summary.DependencyEdges)
	p.Printf("%-20s: %d\n", "Internal Imports", summary.InternalImports)
	p.Printf("%-20s: %d\n", "External Imports", summary.ExternalImports)
	
	if summary.CyclesDetected == 0 {
		fmt.Printf("%-20s: 0 ✓\n", "Cycles")
	} else {
		p.Printf("%-20s: %d\n", "Cycles", summary.CyclesDetected)
	}

	fmt.Println("\nArtifacts")
	fmt.Println()
	fmt.Println("✓ index.json")
	fmt.Println("✓ symbols.json")
	fmt.Println("✓ dependency_graph.json")
	fmt.Println("✓ architecture.json")
	fmt.Println("✓ metrics.json")

	fmt.Println("\nPerformance")
	fmt.Println()
	p.Printf("%-20s: %v\n", "Analysis Time", summary.Duration)
	
	speed := 0.0
	if summary.Duration.Seconds() > 0 {
		speed = float64(summary.FilesAnalyzed) / summary.Duration.Seconds()
	}
	p.Printf("%-20s: %.0f\n", "Files/sec", speed)
	fmt.Printf("%-20s: %s\n", "Output Size", util.FormatSize(summary.ArtifactsSize))
	p.Printf("%-20s: %d\n", "Workers", summary.WorkersCount)

	fmt.Println()
	if validationErr != nil {
		fmt.Printf("Validation          : Failed\n\nReason: %v\n\n", validationErr)
		fmt.Println("Analysis completed with validation errors.")
	} else {
		fmt.Printf("Validation          : Passed ✓\n\n")
		fmt.Println("Analysis completed successfully.")
	}
}

func validateArtifacts(cwd string) error {
	artifacts := []string{
		"index.json",
		"symbols.json",
		"dependency_graph.json",
		"architecture.json",
		"metrics.json",
	}

	for _, artifact := range artifacts {
		path := filepath.Join(cwd, ".kyver", artifact)
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("missing artifact %s: %w", artifact, err)
		}
		
		info, err := file.Stat()
		if err != nil {
			file.Close()
			return fmt.Errorf("cannot stat artifact %s: %w", artifact, err)
		}
		
		if info.Size() == 0 {
			file.Close()
			return fmt.Errorf("artifact %s is empty", artifact)
		}

		// A very fast structural validation by decoding into an empty interface
		// This ensures valid JSON braces and quotes without building specific typed trees
		var dummy interface{}
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&dummy); err != nil {
			file.Close()
			return fmt.Errorf("artifact %s contains invalid JSON: %w", artifact, err)
		}
		
		file.Close()
	}

	return nil
}
