package search

import "github.com/contact-daniel-me/kyver/internal/analyzer"

// Query represents the search query and its parameters
type Query struct {
	Text       string
	Kind       string
	Package    string
	Receiver   string
	Language   string
	File       string
	Imports    string
	Exported   bool
	Docs       string
	Limit      int
	Symbol     bool
	Definition bool
	References bool
}

// Result represents a matched symbol with its computed relevance score
type Result struct {
	Symbol analyzer.Symbol `json:"-"`
	
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Receiver      string `json:"receiver,omitempty"`
	Package       string `json:"package"`
	File          string `json:"file"`
	Line          int    `json:"line"`
	Signature     string `json:"signature,omitempty"`
	Documentation string  `json:"documentation,omitempty"`
	Score         float64 `json:"-"`
	Reason        string  `json:"-"`
}

func NewResult(symbol analyzer.Symbol, score float64, reason string) Result {
	return Result{
		Symbol:        symbol,
		Kind:          symbol.Kind,
		Name:          symbol.Name,
		Receiver:      symbol.Receiver,
		Package:       symbol.Package,
		File:          symbol.FilePath,
		Line:          symbol.StartLine,
		Signature:     symbol.Signature,
		Documentation: symbol.Documentation,
		Score:         score,
		Reason:        reason,
	}
}

// DefinitionProvider locates the exact definition of a symbol.
type DefinitionProvider interface {
	Definition(symbol string) ([]Result, error)
}

// ReferenceProvider locates references to a symbol.
type ReferenceProvider interface {
	References(symbol string) ([]string, error)
}

// RelationProvider locates related symbols.
type RelationProvider interface {
	Related(symbol string) ([]Result, error)
}

// SearchContext provides a unified interface for the developer search engine.
type SearchContext interface {
	SearchEngine
	DefinitionProvider
	ReferenceProvider
	RelationProvider
}

// SearchEngine defines the interface for different search implementations
type SearchEngine interface {
	Search(q Query) ([]Result, error)
}
