package context

import (
	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/graph"
)

type ContextBuilder interface {
	BuildSymbolContext(name string, opts ContextOptions) (*ContextResult, error)
	BuildFileContext(file string, opts ContextOptions) (*ContextResult, error)
	BuildPackageContext(pkg string, opts ContextOptions) (*ContextResult, error)
	BuildFunctionContext(function string, opts ContextOptions) (*ContextResult, error)
}

type ContextOptions struct {
	Sections      []string
	Summary       bool
	Full          bool
	Interactive   bool
	SelectIndex   int
	PackageFilter string
	TreeFormat    bool
	JSONOutput    bool
}

type ContextStats struct {
	Methods    int `json:"methods"`
	Interfaces int `json:"interfaces"`
	Callers    int `json:"callers"`
	Callees    int `json:"callees"`
	References int `json:"references"`
}

type ContextPackage struct {
	Name string `json:"name"`
}

// ContextResult represents the unified JSON structure that bundles context
type ContextResult struct {
	Type            string                       `json:"type,omitempty"`
	Target          string                       `json:"target,omitempty"`
	MultipleFound   bool                         `json:"multipleFound"`
	Candidates      []analyzer.Symbol            `json:"candidates,omitempty"`
	Symbol          *analyzer.Symbol             `json:"symbol,omitempty"`
	Documentation   string                       `json:"documentation,omitempty"`
	Package         *ContextPackage              `json:"package,omitempty"`
	Methods         []analyzer.Symbol            `json:"methods,omitempty"`
	Interfaces      []graph.ImplementedInterface `json:"interfaces,omitempty"`
	Implementations []graph.Implementation       `json:"implementations,omitempty"`
	Callers         *graph.CallersResult         `json:"callers,omitempty"`
	Callees         *graph.CalleesResult         `json:"callees,omitempty"`
	References      *graph.ReferenceResult       `json:"references,omitempty"`
	Impact          *graph.ImpactResult          `json:"impact,omitempty"`
	Dependencies    []string                     `json:"dependencies,omitempty"` // for retrieval package
	Dependents      []string                     `json:"dependents,omitempty"`   // for retrieval package
	Statistics      ContextStats                 `json:"statistics"`
	LookupTimeMs    float64                      `json:"lookupTimeMs"`
}
