package analyzer

import "github.com/contact-daniel-me/kyver/internal/indexer"

// AnalysisResult contains all extracted intelligence from a single file.
type AnalysisResult struct {
	Symbols     []Symbol
	Imports     []string
	Metrics     FileMetrics
	PackageName string
	FilePath    string
}

// Analyzer defines the contract for language-specific parsers.
type Analyzer interface {
	// Language returns the string identifier for this analyzer (e.g. "Go")
	Language() string

	// Analyze extracts symbols and metrics from a file.
	Analyze(fullPath string, file indexer.FileMetadata) (*AnalysisResult, error)
}
