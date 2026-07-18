package repository

import (
	"fmt"
	"os"
	"path/filepath"
)
// Engine manages the .kyver directory lifecycle and state.
type Engine struct {
	// Add repository state here
}

// New initializes a new repository engine.
func New() *Engine {
	return &Engine{}
}

// Init initializes a new Kyver repository in the current directory.
func (e *Engine) Init() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	kyverDir := filepath.Join(cwd, ".kyver")

	// Check if already initialized
	if _, err := os.Stat(kyverDir); err == nil {
		fmt.Println("Kyver repository already initialized.")
		fmt.Printf("Location: %s\n", kyverDir)
		return nil
	}

	// Create .kyver directory structure
	dirs := []string{
		kyverDir,
		filepath.Join(kyverDir, "adr"),
		filepath.Join(kyverDir, "commits"),
		filepath.Join(kyverDir, "context"),
		filepath.Join(kyverDir, "graph"),
		filepath.Join(kyverDir, "memory"),
		filepath.Join(kyverDir, "prompts"),
		filepath.Join(kyverDir, "requirements"),
		filepath.Join(kyverDir, "retrieval"),
		filepath.Join(kyverDir, "timeline"),
		filepath.Join(kyverDir, "traceability"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write config.json
	configPath := filepath.Join(kyverDir, "config.json")
	configContent := `{
  "version": "1.0",
  "kyverVersion": "v1.0.0",
  "initialized": true
}
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Write README
	readmePath := filepath.Join(kyverDir, "README.md")
	readmeContent := "# .kyver\n\nThis directory is managed by Kyver. Do not edit files here manually.\n\nRun 'kyver index' to build the semantic index.\n"
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	repoName := filepath.Base(cwd)
	fmt.Printf("✓ Initialized Kyver repository in %s\n", cwd)
	fmt.Printf("  Repository: %s\n", repoName)
	fmt.Printf("  Location  : %s\n", kyverDir)
	fmt.Println()
	fmt.Println("Next step: run 'kyver index' to build the semantic index.")

	return nil
}
