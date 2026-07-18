package graph

import "github.com/contact-daniel-me/kyver/internal/analyzer"

type GraphEngine interface {
	Callers(symbol string, opts CallersOptions) (*CallersResult, error)
	Callees(symbol string, opts CalleesOptions) (*CalleesResult, error)
	References(symbol string, opts ReferenceOptions) (*ReferenceResult, error)
	Impacts(symbol string, opts ImpactOptions) (*ImpactResult, error)
	PackageGraph(pkg string) (*PackageGraphResult, error)
	FileGraph(file string) (*FileGraphResult, error)
	Implements(symbol string) (*InterfaceResult, error)
	Interface(symbol string) (*InterfaceResult, error)
	Graph(symbol string, opts GraphOptions) (*SymbolGraphResult, error)
}

type CalleesOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	SortBy        string
	TreeFormat    bool
}

type CalleesResult struct {
	Symbol        *analyzer.Symbol    `json:"symbol,omitempty"`
	Candidates    []analyzer.Symbol   `json:"candidates,omitempty"`
	MultipleFound bool                `json:"multipleFound,omitempty"`
	Groups        []CalleeGroup       `json:"groups,omitempty"`
	Stats         CalleesStats        `json:"stats"`
}

type CalleeGroup struct {
	Package string       `json:"package"`
	Callees []CalleeNode `json:"callees,omitempty"`
}

type CalleeNode struct {
	analyzer.Symbol
}

type CalleesStats struct {
	TotalCallees int     `json:"totalCallees"`
	TotalFiles   int     `json:"totalFiles"`
	TotalPkgs    int     `json:"totalPackages"`
	LookupTimeMs float64 `json:"lookupTimeMs"`
}

type CallersOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	SortBy        string
	TreeFormat    bool
}

type CallersResult struct {
	Symbol        *analyzer.Symbol    `json:"symbol,omitempty"`
	Candidates    []analyzer.Symbol   `json:"candidates,omitempty"`
	MultipleFound bool                `json:"multipleFound,omitempty"`
	Groups        []CallerGroup  `json:"groups,omitempty"`
	Stats         CallersStats   `json:"stats"`
}

type CallerGroup struct {
	Package string       `json:"package"`
	Callers []CallerNode `json:"callers,omitempty"`
}

type CallerNode struct {
	analyzer.Symbol
	IsRecursive bool `json:"isRecursive"`
}

type CallersStats struct {
	TotalCallers int     `json:"totalCallers"`
	TotalFiles   int     `json:"totalFiles"`
	TotalPkgs    int     `json:"totalPackages"`
	LookupTimeMs float64 `json:"lookupTimeMs"`
}

type SymbolResult struct {
	Symbol     *analyzer.Symbol
	Candidates []analyzer.Symbol
	Found      []analyzer.Symbol
}

type ReferenceOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	FileFilter    string
	SortBy        string
	TreeFormat    bool
	Interactive   bool
}

type ReferenceResult struct {
	Symbol        *analyzer.Symbol  `json:"symbol,omitempty"`
	Candidates    []analyzer.Symbol `json:"candidates,omitempty"`
	MultipleFound bool              `json:"multipleFound,omitempty"`
	Groups        []ReferenceGroup  `json:"groups,omitempty"`
	Stats         ReferenceStats    `json:"stats"`
	LookupTimeMs  float64           `json:"lookupTimeMs"`
}

type ReferenceGroup struct {
	Package    string      `json:"package"`
	References []Reference `json:"references,omitempty"`
}

type Reference struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Package      string `json:"package"`
	File         string `json:"file"`
	Line         int    `json:"line"`
	Context      string `json:"context"`
	Receiver     string `json:"receiver,omitempty"`
	Relationship string `json:"relationship,omitempty"`
}

type ReferenceStats struct {
	TotalReferences int `json:"totalReferences"`
	TotalFiles      int `json:"totalFiles"`
	TotalPkgs       int `json:"totalPackages"`
}

type ImpactOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	SortBy        string
	TreeFormat    bool
	Interactive   bool
}

type ImpactResult struct {
	Symbol            *analyzer.Symbol  `json:"symbol,omitempty"`
	Candidates        []analyzer.Symbol `json:"candidates,omitempty"`
	MultipleFound     bool              `json:"multipleFound,omitempty"`
	DirectCallers     []ImpactCaller    `json:"directCallers,omitempty"`
	IndirectCallers   []ImpactCaller    `json:"indirectCallers,omitempty"`
	AffectedFiles     []string          `json:"affectedFiles,omitempty"`
	AffectedPackages  []string          `json:"affectedPackages,omitempty"`
	Tests             []ImpactCaller    `json:"tests,omitempty"`
	Severity          string            `json:"severity"`
	SeverityRationale string            `json:"-"`
	Stats             ImpactStats       `json:"stats"`
	LookupTimeMs      float64           `json:"lookupTimeMs"`
}

type ImpactCaller struct {
	analyzer.Symbol
	Calls string   `json:"calls,omitempty"` // immediate callee on the path to the target
	Chain []string `json:"chain,omitempty"`           // human-readable why explanation
}

type ImpactStats struct {
	DirectCallers   int     `json:"directCallers"`
	IndirectCallers int     `json:"indirectCallers"`
	AffectedFiles   int     `json:"affectedFiles"`
	AffectedPkgs    int     `json:"affectedPackages"`
	PublicAPIs      int     `json:"publicApis"`
}

type PackageGraphResult struct {
	Package  string   `json:"package"`
	Imports  []string `json:"imports,omitempty"`
	Imported []string `json:"imported_by,omitempty"`
}

type FileGraphResult struct {
	File     string            `json:"file"`
	Imports  []string          `json:"imports,omitempty"`
	Symbols  []analyzer.Symbol `json:"symbols"`
	Incoming []string          `json:"incoming_dependencies,omitempty"`
}

type InterfaceStats struct {
	TotalInterfaces int `json:"totalInterfaces"`
	TotalPackages   int `json:"totalPackages"`
}

type ImplementedInterface struct {
	Name               string   `json:"name"`
	Package            string   `json:"package"`
	RequiredMethods    []string `json:"requiredMethods,omitempty"`
	ImplementedMethods []string `json:"implementedMethods,omitempty"`
	MissingMethods     []string `json:"missingMethods,omitempty"`
	PointerReceiver    bool     `json:"pointerReceiver"`
}

type InterfaceOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	SortBy        string
	TreeFormat    bool
	Interactive   bool
}

type InterfaceResult struct {
	Symbol        *analyzer.Symbol       `json:"symbol,omitempty"`
	Candidates    []analyzer.Symbol      `json:"candidates,omitempty"`
	MultipleFound bool                   `json:"multipleFound,omitempty"`
	Interfaces    []ImplementedInterface `json:"interfaces,omitempty"`
	Stats         InterfaceStats         `json:"stats"`
	LookupTimeMs  float64                `json:"lookupTimeMs"`
}

type ImplementsOptions struct {
	SelectIndex   int
	Max           int
	PackageFilter string
	SortBy        string
	TreeFormat    bool
	Interactive   bool
}

type ImplementsResult struct {
	Symbol          *analyzer.Symbol  `json:"symbol,omitempty"`
	Candidates      []analyzer.Symbol `json:"candidates,omitempty"`
	MultipleFound   bool              `json:"multipleFound,omitempty"`
	RequiredMethods []string          `json:"requiredMethods,omitempty"`
	Implementations []Implementation  `json:"implementations,omitempty"`
	LookupTimeMs    float64           `json:"lookupTimeMs"`
}

type Implementation struct {
	Struct             string   `json:"struct"`
	Package            string   `json:"package"`
	PointerReceiver    bool     `json:"pointerReceiver"`
	ImplementedMethods []string `json:"implementedMethods,omitempty"`
	MissingMethods     []string `json:"missingMethods,omitempty"`
}

type GraphOptions struct {
	SelectIndex int
	TreeFormat  bool
	Interactive bool
}

type GraphStats struct {
	Methods      int     `json:"methods"`
	Callers      int     `json:"callers"`
	Callees      int     `json:"callees"`
	Dependencies int     `json:"dependencies"`
	References   int     `json:"references"`
	LookupTimeMs float64 `json:"lookupTimeMs"`
}

type SymbolGraphResult struct {
	Symbol        *analyzer.Symbol  `json:"symbol,omitempty"`
	Candidates    []analyzer.Symbol `json:"candidates,omitempty"`
	MultipleFound bool              `json:"multipleFound,omitempty"`
	Methods       []analyzer.Symbol `json:"methods"`
	Callers       []analyzer.Symbol `json:"callers"`
	Callees       []analyzer.Symbol `json:"callees"`
	Dependencies  []string          `json:"dependencies,omitempty"`
	References    []Reference       `json:"references,omitempty"`
	Stats         GraphStats        `json:"stats"`
}
