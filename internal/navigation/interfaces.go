package navigation

import "github.com/contact-daniel-me/kyver/internal/analyzer"

type NavigationEngine interface {
	Goto(symbol string, selectIndex int) (*GotoResult, error)
	Peek(symbol string, selectIndex int, linesLimit int) (*PeekResult, error)
	Outline(file string, filter string, sortBy string) (*OutlineResult, error)
	Hierarchy(symbol string, selectIndex int) (*HierarchyResult, error)
}

type GotoResult struct {
	MultipleFound bool              `json:"multipleFound"`
	Candidates    []analyzer.Symbol `json:"candidates,omitempty"`
	Symbol        *analyzer.Symbol  `json:"symbol,omitempty"`
}

type PeekResult struct {
	MultipleFound  bool              `json:"multipleFound"`
	Candidates     []analyzer.Symbol `json:"candidates,omitempty"`
	Symbol         *analyzer.Symbol  `json:"symbol,omitempty"`
	Snippet        string            `json:"snippet,omitempty"`
	Methods        []analyzer.Symbol `json:"methods,omitempty"`
	Implements     []analyzer.Symbol `json:"implements,omitempty"`
	ReferencedBy   []string          `json:"referencedBy,omitempty"`
	Dependencies   []string          `json:"dependencies,omitempty"`
	Truncated      bool              `json:"truncated,omitempty"`
	TruncatedLines int               `json:"truncatedLines,omitempty"`
}

type OutlineResult struct {
	File     string        `json:"file"`
	Package  string        `json:"package"`
	Language string        `json:"language"`
	Imports  []string      `json:"imports"`
	Groups   []SymbolGroup `json:"groups"`
	Stats    OutlineStats  `json:"stats"`
}

type SymbolGroup struct {
	Kind    string            `json:"kind"`
	Symbols []analyzer.Symbol `json:"symbols"`
}

type OutlineStats struct {
	TotalSymbols int            `json:"totalSymbols"`
	TotalImports int            `json:"totalImports"`
	Counts       map[string]int `json:"counts"`
}

type HierarchyResult struct {
	MultipleFound bool              `json:"multipleFound"`
	Candidates    []analyzer.Symbol `json:"candidates,omitempty"`
	Symbol        *analyzer.Symbol  `json:"symbol,omitempty"`
	Methods       []analyzer.Symbol `json:"methods"`
	Related       []analyzer.Symbol `json:"related"`
	Dependencies  []string          `json:"dependencies"`
}
