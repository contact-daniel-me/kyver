package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/contact-daniel-me/kyver/internal/status"
)

func runStatus(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	mode := "human"
	for _, arg := range args {
		if arg == "--json" {
			mode = "json"
		} else if arg == "--summary" {
			mode = "summary"
		}
	}

	indexPath := filepath.Join(cwd, ".kyver", "index.json")
	indexData, err := status.ReadIndex(indexPath)
	// We proceed even if err != nil, to report the missing index status

	gitInfo, _ := status.GetGitInfo(cwd)

	var stats *status.Statistics
	var recs []string
	if indexData != nil {
		stats = status.CalculateStats(indexData)
		recs = status.CheckHealth(cwd, indexData)
	} else {
		stats = &status.Statistics{}
		recs = []string{fmt.Sprintf("Index error: %v", err)}
	}

	report := status.GenerateReport(cwd, indexData, gitInfo, stats, recs)

	if mode == "json" {
		return status.PrintJSON(report)
	} else if mode == "summary" {
		status.PrintSummary(report)
	} else {
		status.PrintHumanReadable(report)
	}

	return nil
}
