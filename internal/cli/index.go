package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/contact-daniel-me/kyver/internal/indexer"
	"github.com/contact-daniel-me/kyver/internal/util"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func runIndex(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	verbose := false
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		}
	}

	// Detect Git repository
	gitPath := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return fmt.Errorf("repository not initialized: .git directory not found in %s", cwd)
	}

	fmt.Println("Scanning repository...")

	idx := indexer.New(cwd)
	res, err := idx.Index(verbose)
	if err != nil {
		if errors.Is(err, indexer.ErrPermissionDenied) {
			return fmt.Errorf("permission denied: unable to create or write to .kyver directory")
		}
		if errors.Is(err, indexer.ErrCorruptedGit) {
			return fmt.Errorf("repository error: git structure appears invalid or corrupted")
		}
		return fmt.Errorf("indexing failed: %w", err)
	}

	printSummary(cwd, res, verbose)

	if err := validateIndex(res.IndexPath, res.FilesIndexed); err != nil {
		return fmt.Errorf("index validation failed: %w", err)
	}
	fmt.Println("\n✓ Index validation passed")

	return nil
}

func printSummary(cwd string, res indexer.Result, verbose bool) {
	p := message.NewPrinter(language.English)
	repoName := filepath.Base(cwd)

	fmt.Println("\nRepository Summary")
	fmt.Println("------------------")
	fmt.Printf("Repository Name : %s\n", repoName)
	fmt.Printf("Repository Root : %s\n\n", cwd)

	fmt.Println("Files Indexed")
	fmt.Println("-------------")
	for lang, count := range res.Stats.Languages {
		p.Printf("%-14s: %d\n", lang, count)
	}
	fmt.Println()
	p.Printf("Total Source Files : %d\n", res.FilesIndexed)
	p.Printf("Directories        : %d\n\n", res.Stats.Directories)
	
	fmt.Println("Note: 'Total Source Files' counts only supported code files, ignoring binaries and unknown extensions.")
	fmt.Println()

	fmt.Println("Excluded")
	fmt.Println("--------")
	for _, dir := range res.Stats.ExcludedDirs {
		fmt.Printf("✓ %s\n", dir)
	}
	fmt.Println("✓ unsupported extensions")
	fmt.Println()

	fmt.Println("Performance")
	fmt.Println("-----------")
	p.Printf("Indexing Time     : %v\n", res.Duration)
	fmt.Printf("Index Size        : %s\n", util.FormatSize(res.IndexSize))
	speed := 0.0
	if res.Duration.Seconds() > 0 {
		speed = float64(res.FilesIndexed) / res.Duration.Seconds()
	}
	p.Printf("Processing Speed  : %.0f files/sec\n", speed)
	fmt.Printf("Memory Used       : %s\n", util.FormatSize(int64(res.MemoryUsed)))
	fmt.Println()

	if verbose {
		fmt.Println("Verbose Information")
		fmt.Println("-------------------")
		if len(res.Stats.SkippedFiles) > 0 {
			for _, skipped := range res.Stats.SkippedFiles {
				fmt.Printf("Skipped: %s (%s)\n", skipped.Path, skipped.Reason)
			}
		} else {
			fmt.Println("No files skipped due to errors or invalid names.")
		}
		fmt.Println()
	}
}

func validateIndex(indexPath string, expectedCount int) error {
	file, err := os.Open(indexPath)
	if err != nil {
		return fmt.Errorf("cannot read generated index file: %w", err)
	}
	defer file.Close()

	var data indexer.IndexData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("index JSON is malformed: %w", err)
	}

	if len(data.Files) != expectedCount {
		return fmt.Errorf("file count mismatch: index contains %d files, expected %d", len(data.Files), expectedCount)
	}

	if data.Repository == "" {
		return fmt.Errorf("missing repository name in index")
	}

	return nil
}
