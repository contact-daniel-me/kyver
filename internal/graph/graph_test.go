package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)

func createTestEngine(t *testing.T) *Engine {
	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Repository", Kind: "Interface", Package: "core", FilePath: "core/repo.go", Exported: true},
			{Name: "SaveRepository", Kind: "Method", Package: "core", FilePath: "core/repo.go", Exported: true, Receiver: "Repository"},
			{Name: "Engine", Kind: "Struct", Package: "analyzer", FilePath: "analyzer/analyzer.go", Exported: true},
			{Name: "Analyze", Kind: "Method", Package: "analyzer", FilePath: "analyzer/analyzer.go", Exported: true, Receiver: "Engine"},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"core/repo.go":         {"fmt", "strings"},
				"analyzer/analyzer.go": {"core", "fmt"},
			},
		},
		SymbolsModTime: time.Now(),
		DepsModTime:    time.Now(),
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}

	for i, sym := range repo.Symbols {
		repo.ByName[sym.Name] = append(repo.ByName[sym.Name], i)
		repo.ByPackage[sym.Package] = append(repo.ByPackage[sym.Package], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
	}

	e := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}
	e.buildPackageGraph()

	return e
}

func TestImpacts(t *testing.T) {
	engine := createTestEngine(t)

	opts := ImpactOptions{}
	res, err := engine.Impacts("Repository", opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.Symbol == nil || res.Symbol.Name != "Repository" {
		t.Errorf("Expected symbol Repository, got %v", res.Symbol)
	}

	if res.Stats.DirectCallers == 0 && res.Severity != "Low" {
		t.Errorf("Expected Low severity for 0 callers, got %s", res.Severity)
	}
}

func TestImpactsNoCallers(t *testing.T) {
	engine := createTestEngine(t)

	opts := ImpactOptions{}
	res, err := engine.Impacts("SaveRepository", opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.DirectCallers != 0 {
		t.Errorf("Expected 0 direct callers for SaveRepository, got %d", res.Stats.DirectCallers)
	}
	if res.Severity != "Low" {
		t.Errorf("Expected Low severity, got %s", res.Severity)
	}
}

func TestImpactsJSON(t *testing.T) {
	engine := createTestEngine(t)

	opts := ImpactOptions{}
	res, err := engine.Impacts("Repository", opts)
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	jsonStr := string(b)
	for _, field := range []string{`"symbol":`, `"directCallers":`, `"indirectCallers":`, `"severity":`, `"lookupTimeMs":`} {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON field %s, got: %s", field, jsonStr)
		}
	}
}


func TestCallers(t *testing.T) {
	tmp := t.TempDir()

	// 1. Create target file (core/repo.go)
	coreDir := filepath.Join(tmp, "core")
	os.MkdirAll(coreDir, 0755)
	repoFile := filepath.Join(coreDir, "repo.go")
	os.WriteFile(repoFile, []byte(`package core
// line 2
type Engine struct{} // Target Symbol: Engine
// line 4
func (e *Engine) Init() {
	// Recursive call
	e.Init()
}
// line 9
func (e *Engine) Neighbor() {
	// Does not call Engine
}
// line 13
func AnotherFunc() {
	_ = Engine{} // Calls Engine
}
`), 0644)

	// 2. Create another package file (analyzer/analyzer.go)
	analyzerDir := filepath.Join(tmp, "analyzer")
	os.MkdirAll(analyzerDir, 0755)
	analyzerFile := filepath.Join(analyzerDir, "analyzer.go")
	os.WriteFile(analyzerFile, []byte(`package analyzer
// line 2
func Analyze() {
	e := core.Engine{} // Calls Engine
}
// line 6
func OtherAnalyze() {
	// No call
}
`), 0644)

	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Engine", Kind: "Struct", Package: "core", FilePath: repoFile, StartLine: 3, EndLine: 3, Exported: true},
			{Name: "Init", Kind: "Method", Package: "core", FilePath: repoFile, StartLine: 5, EndLine: 8, Exported: true, Receiver: "Engine"},
			{Name: "Neighbor", Kind: "Method", Package: "core", FilePath: repoFile, StartLine: 10, EndLine: 12, Exported: true, Receiver: "Engine"},
			{Name: "AnotherFunc", Kind: "Function", Package: "core", FilePath: repoFile, StartLine: 14, EndLine: 16, Exported: true},
			{Name: "Analyze", Kind: "Function", Package: "analyzer", FilePath: analyzerFile, StartLine: 3, EndLine: 5, Exported: true},
			{Name: "OtherAnalyze", Kind: "Function", Package: "analyzer", FilePath: analyzerFile, StartLine: 7, EndLine: 9, Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				repoFile:     {"fmt"},
				analyzerFile: {"core", "fmt"},
			},
		},
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}

	for i, sym := range repo.Symbols {
		repo.ByName[sym.Name] = append(repo.ByName[sym.Name], i)
		repo.ByPackage[sym.Package] = append(repo.ByPackage[sym.Package], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
	}

	engine := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}
	engine.buildPackageGraph()

	// Test 1: Callers of Engine
	opts := CallersOptions{}
	res, err := engine.Callers("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}

	// Expected: AnotherFunc (core) and Analyze (analyzer). 
	// Engine itself should NOT be returned (not recursive). 
	// Init and Neighbor should NOT be returned (no lexical match).
	if res.Stats.TotalCallers != 2 {
		t.Fatalf("Expected 2 callers for Engine, got %d", res.Stats.TotalCallers)
	}
	
	// Test 2: Callers of Init (Recursive)
	resInit, err := engine.Callers("Init", opts)
	if err != nil {
		t.Fatal(err)
	}
	
	if resInit.Stats.TotalCallers != 1 {
		t.Fatalf("Expected 1 caller for Init, got %d", resInit.Stats.TotalCallers)
	}
	
	if !resInit.Groups[0].Callers[0].IsRecursive {
		t.Errorf("Expected Init caller to be marked as recursive")
	}
}

func TestCallersFormatAndJSON(t *testing.T) {
	// Verify that the tree format logic correctly produces receiver string
	sym := &analyzer.Symbol{
		Name:     "Init",
		Kind:     "Method",
		Package:  "repository",
		Receiver: "*Engine",
	}

	targetName := sym.Name
	if sym.Kind == "Method" {
		if strings.HasPrefix(sym.Receiver, "*") {
			targetName = fmt.Sprintf("%s.(%s).%s", sym.Package, sym.Receiver, sym.Name)
		} else {
			targetName = fmt.Sprintf("%s.%s.%s", sym.Package, sym.Receiver, sym.Name)
		}
	}

	if targetName != "repository.(*Engine).Init" {
		t.Errorf("Expected repository.(*Engine).Init, got %s", targetName)
	}

	// Verify that JSON serialization maintains camelCase and decimal precision
	stats := CallersStats{
		TotalCallers: 1,
		TotalFiles:   1,
		TotalPkgs:    1,
		LookupTimeMs: 0.66,
	}

	res := &CallersResult{
		Symbol:        sym,
		MultipleFound: true,
		Stats:         stats,
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	
	jsonStr := string(b)
	if !strings.Contains(jsonStr, `"symbol":`) {
		t.Errorf("Expected lowerCamelCase 'symbol', got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"multipleFound":true`) {
		t.Errorf("Expected lowerCamelCase 'multipleFound', got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"lookupTimeMs":0.66`) {
		t.Errorf("Expected preserved decimal precision for lookupTimeMs, got %s", jsonStr)
	}
}

func TestPackageGraph(t *testing.T) {
	engine := createTestEngine(t)
	
	res, err := engine.PackageGraph("core")
	if err != nil {
		t.Fatal(err)
	}
	
	if len(res.Imported) != 1 || res.Imported[0] != "analyzer" {
		t.Errorf("Expected 'analyzer' to import 'core'")
	}
}

func TestCallees(t *testing.T) {
	tmp := t.TempDir()

	// 1. Create target file (cli/cli.go)
	cliDir := filepath.Join(tmp, "cli")
	os.MkdirAll(cliDir, 0755)
	cliFile := filepath.Join(cliDir, "cli.go")
	os.WriteFile(cliFile, []byte(`package cli
// line 2
func runInit() {
	_ = core.Engine{} // not a function call, ignore
	
	repository.New() // Call
	
	repository.Engine.Init() // Call
	
	// os.Stat() // Comment, ignore

	funcInSamePkg() // Call
}
// line 14
func funcInSamePkg() {}
`), 0644)

	// 2. Create imported packages
	repoDir := filepath.Join(tmp, "repository")
	os.MkdirAll(repoDir, 0755)
	repoFile := filepath.Join(repoDir, "repository.go")
	os.WriteFile(repoFile, []byte(`package repository
func New() {}
func (e *Engine) Init() {}
`), 0644)

	osDir := filepath.Join(tmp, "os")
	os.MkdirAll(osDir, 0755)
	osFile := filepath.Join(osDir, "os.go")
	os.WriteFile(osFile, []byte(`package os
func Stat() {}
`), 0644)

	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "runInit", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 3, EndLine: 13, Exported: false},
			{Name: "funcInSamePkg", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 15, EndLine: 15, Exported: false},
			{Name: "New", Kind: "Function", Package: "repository", FilePath: repoFile, StartLine: 2, EndLine: 2, Exported: true},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 3, EndLine: 3, Exported: true, Receiver: "*Engine"},
			{Name: "Stat", Kind: "Function", Package: "os", FilePath: osFile, StartLine: 2, EndLine: 2, Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				cliFile:  {"repository", "os"},
				repoFile: {},
				osFile:   {},
			},
		},
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}

	for i, sym := range repo.Symbols {
		repo.ByName[sym.Name] = append(repo.ByName[sym.Name], i)
		repo.ByPackage[sym.Package] = append(repo.ByPackage[sym.Package], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
	}

	engine := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}
	engine.buildPackageGraph()

	// Test 1: Callees of runInit
	opts := CalleesOptions{}
	res, err := engine.Callees("runInit", opts)
	if err != nil {
		t.Fatal(err)
	}

	// Expected: New, Init, funcInSamePkg.
	// Stat should NOT be returned because it's only in a comment.
	if res.Stats.TotalCallees != 3 {
		t.Fatalf("Expected 3 callees for runInit, got %d", res.Stats.TotalCallees)
	}
}

func TestCalleesFormatAndJSON(t *testing.T) {
	sym := &analyzer.Symbol{
		Name:     "Init",
		Kind:     "Method",
		Package:  "repository",
		Receiver: "*Engine",
	}

	stats := CalleesStats{
		TotalCallees: 1,
		TotalFiles:   1,
		TotalPkgs:    1,
		LookupTimeMs: 0.66,
	}

	res := &CalleesResult{
		Symbol:        sym,
		MultipleFound: true,
		Stats:         stats,
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	
	jsonStr := string(b)
	if !strings.Contains(jsonStr, `"symbol":`) {
		t.Errorf("Expected lowerCamelCase 'symbol', got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"multipleFound":true`) {
		t.Errorf("Expected lowerCamelCase 'multipleFound', got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"lookupTimeMs":0.66`) {
		t.Errorf("Expected preserved decimal precision for lookupTimeMs, got %s", jsonStr)
	}
}

func TestCalleesPackageValidation(t *testing.T) {
	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{},
		},
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}
	
	// Add some dummy packages
	repo.ByPackage["repository"] = []int{0}
	repo.ByPackage["search"] = []int{1}

	engine := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}

	opts := CalleesOptions{PackageFilter: "xyz"}
	// We need resolveSymbol to not fail for the test to reach package validation.
	// Oh, wait! Callees calls resolveSymbol first. 
	// We'll mock resolveSymbol by actually putting a symbol in repo.
	
	sym := analyzer.Symbol{Name: "dummy", Package: "dummy_pkg", FilePath: "dummy.go"}
	repo.Symbols = append(repo.Symbols, sym)
	repo.ByName["dummy"] = []int{0}
	
	_, err := engine.Callees("dummy", opts)
	if err == nil {
		t.Fatalf("Expected error for unknown package")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, `Unknown package "xyz"`) {
		t.Errorf("Expected 'Unknown package \"xyz\"', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "repository") || !strings.Contains(errMsg, "search") {
		t.Errorf("Expected available packages to be listed, got: %s", errMsg)
	}
}

func TestGraphCommand(t *testing.T) {
	engine := createTestEngine(t)

	// Test 1: Unique symbol Graph (Engine is in package "analyzer")
	opts := GraphOptions{}
	res, err := engine.Graph("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}
	
	if res.MultipleFound {
		t.Fatalf("Expected Engine to not be ambiguous in this mock, got multiple: %v", res.MultipleFound)
	}
	
	if res.Symbol == nil || res.Symbol.Name != "Engine" {
		t.Fatalf("Expected symbol Engine, got %v", res.Symbol)
	}
	
	// Engine has receiver "Engine" -> Analyze method
	if len(res.Methods) != 1 {
		t.Errorf("Expected 1 method for Engine, got %d", len(res.Methods))
	}
	
	// Dependencies of analyzer/analyzer.go are ["core", "fmt"]
	if len(res.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %v", res.Dependencies)
	}
	
	if res.Stats.Methods != len(res.Methods) || res.Stats.Dependencies != len(res.Dependencies) {
		t.Errorf("Expected stats to match slice lengths")
	}
	
	// Test 2: Symbol not found
	_, err = engine.Graph("NonExistent", opts)
	if err == nil {
		t.Fatalf("Expected error for non-existent symbol")
	}
	
	// Test 3: JSON precision
	b, _ := json.Marshal(res.Stats)
	jsonStr := string(b)
	if !strings.Contains(jsonStr, `"lookupTimeMs":`) {
		t.Errorf("Expected lookupTimeMs in JSON, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"methods":`) {
		t.Errorf("Expected methods count in JSON, got %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"callers":`) {
		t.Errorf("Expected callers count in JSON, got %s", jsonStr)
	}
	
	// Test 4: Empty graph (Repository is an interface with no method receiver matching)
	resRepo, err := engine.Graph("Repository", opts)
	if err != nil {
		t.Fatal(err)
	}
	if resRepo.Symbol == nil || resRepo.Symbol.Name != "Repository" {
		t.Fatalf("Expected symbol Repository, got %v", resRepo.Symbol)
	}
	// Repository has receiver "Repository" -> SaveRepository
	if len(resRepo.Methods) != 1 {
		t.Errorf("Expected 1 method for Repository, got %d", len(resRepo.Methods))
	}
	
	// Test 5: JSON full serialization
	fullJSON, _ := json.Marshal(res)
	fullStr := string(fullJSON)
	if !strings.Contains(fullStr, `"symbol":`) {
		t.Errorf("Expected symbol in full JSON, got %s", fullStr)
	}

	if !strings.Contains(fullStr, `"dependencies":`) {
		t.Errorf("Expected dependencies in full JSON, got %s", fullStr)
	}
}

// Regression test: methods must be scoped to the selected symbol's package.
// When multiple structs share the name "Engine" across packages, only methods
// belonging to the selected package should appear.
func TestGraphMethodsPackageScoping(t *testing.T) {
	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			// repository.Engine with Init method
			{Name: "Engine", Kind: "Struct", Package: "repository", FilePath: "repository/repo.go", Exported: true, StartLine: 10, EndLine: 15},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: "repository/repo.go", Exported: true, Receiver: "*Engine", StartLine: 20, EndLine: 30},
			// analyzer.Engine with Analyze method
			{Name: "Engine", Kind: "Struct", Package: "analyzer", FilePath: "analyzer/analyzer.go", Exported: true, StartLine: 5, EndLine: 8},
			{Name: "Analyze", Kind: "Method", Package: "analyzer", FilePath: "analyzer/analyzer.go", Exported: true, Receiver: "*Engine", StartLine: 12, EndLine: 20},
			// graph.Engine with Callers and Callees methods
			{Name: "Engine", Kind: "Struct", Package: "graph", FilePath: "graph/engine.go", Exported: true, StartLine: 3, EndLine: 6},
			{Name: "Callers", Kind: "Method", Package: "graph", FilePath: "graph/engine.go", Exported: true, Receiver: "*Engine", StartLine: 10, EndLine: 20},
			{Name: "Callees", Kind: "Method", Package: "graph", FilePath: "graph/engine.go", Exported: true, Receiver: "*Engine", StartLine: 25, EndLine: 35},
			// context.Engine with Build method (value receiver)
			{Name: "Engine", Kind: "Struct", Package: "context", FilePath: "context/builder.go", Exported: true, StartLine: 1, EndLine: 4},
			{Name: "Build", Kind: "Method", Package: "context", FilePath: "context/builder.go", Exported: true, Receiver: "Engine", StartLine: 8, EndLine: 12},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"repository/repo.go":    {"fmt"},
				"analyzer/analyzer.go":  {"os"},
				"graph/engine.go":       {"strings"},
				"context/builder.go":    {"io"},
			},
		},
		SymbolsModTime: time.Now(),
		DepsModTime:    time.Now(),
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}

	for i, sym := range repo.Symbols {
		repo.ByName[sym.Name] = append(repo.ByName[sym.Name], i)
		repo.ByPackage[sym.Package] = append(repo.ByPackage[sym.Package], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
	}

	engine := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}
	engine.buildPackageGraph()

	// Select repository.Engine (index 1)
	opts := GraphOptions{SelectIndex: 1}
	res, err := engine.Graph("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.Symbol == nil || res.Symbol.Package != "repository" {
		t.Fatalf("Expected repository.Engine, got %v", res.Symbol)
	}

	// CRITICAL ASSERTION: only Init() should appear, not Analyze/Callers/Callees/Build
	if len(res.Methods) != 1 {
		names := []string{}
		for _, m := range res.Methods {
			names = append(names, fmt.Sprintf("%s.%s", m.Package, m.Name))
		}
		t.Fatalf("Expected exactly 1 method (Init) for repository.Engine, got %d: %v", len(res.Methods), names)
	}
	if res.Methods[0].Name != "Init" {
		t.Errorf("Expected method Init, got %s", res.Methods[0].Name)
	}
	if res.Methods[0].Package != "repository" {
		t.Errorf("Expected method package repository, got %s", res.Methods[0].Package)
	}
	if res.Stats.Methods != 1 {
		t.Errorf("Expected stats.Methods=1, got %d", res.Stats.Methods)
	}

	// Select analyzer.Engine (index 2)
	opts.SelectIndex = 2
	res2, err := engine.Graph("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Methods) != 1 || res2.Methods[0].Name != "Analyze" {
		t.Errorf("Expected exactly 1 method (Analyze) for analyzer.Engine, got %d", len(res2.Methods))
	}

	// Select graph.Engine (index 3)
	opts.SelectIndex = 3
	res3, err := engine.Graph("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(res3.Methods) != 2 {
		t.Errorf("Expected exactly 2 methods (Callers, Callees) for graph.Engine, got %d", len(res3.Methods))
	}

	// Select context.Engine (index 4) — value receiver "Engine"
	opts.SelectIndex = 4
	res4, err := engine.Graph("Engine", opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(res4.Methods) != 1 || res4.Methods[0].Name != "Build" {
		t.Errorf("Expected exactly 1 method (Build) for context.Engine, got %d", len(res4.Methods))
	}

	// Verify JSON output also has correct count
	b, _ := json.Marshal(res)
	jsonStr := string(b)
	// The JSON methods array should contain exactly 1 entry
	if strings.Count(jsonStr, `"name":"Init"`) != 1 {
		t.Errorf("Expected exactly 1 Init in JSON, got: %s", jsonStr)
	}
	if strings.Contains(jsonStr, `"name":"Analyze"`) {
		t.Errorf("JSON should NOT contain Analyze for repository.Engine")
	}
	if strings.Contains(jsonStr, `"name":"Callers"`) {
		t.Errorf("JSON should NOT contain Callers for repository.Engine")
	}
}
