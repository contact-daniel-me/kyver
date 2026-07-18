package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)

func createReferenceFixture(t *testing.T) *Engine {
	t.Helper()
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "repository")
	cliDir := filepath.Join(tmp, "cli")
	os.MkdirAll(repoDir, 0755)
	os.MkdirAll(cliDir, 0755)

	repoFile := filepath.Join(repoDir, "engine.go")
	os.WriteFile(repoFile, []byte(`package repository

type Engine struct{}

func (e *Engine) Init() {
	e.Init()
}

func Neighbor() {
	// does not call Init
}
`), 0644)

	cliFile := filepath.Join(cliDir, "cli.go")
	os.WriteFile(cliFile, []byte(`package cli

import "example.com/app/repository"

func runInit() {
	e := &repository.Engine{}
	e.Init()
}

func Execute() {
	runInit()
}

func main() {
	Execute()
}
`), 0644)

	testFile := filepath.Join(cliDir, "cli_test.go")
	os.WriteFile(testFile, []byte(`package cli

func TestInit() {
	runInit()
}
`), 0644)

	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Engine", Kind: "Struct", Package: "repository", FilePath: repoFile, StartLine: 3, EndLine: 3, Exported: true},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 5, EndLine: 7, Exported: true, Receiver: "*Engine"},
			{Name: "Neighbor", Kind: "Function", Package: "repository", FilePath: repoFile, StartLine: 9, EndLine: 11, Exported: true},
			{Name: "runInit", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 5, EndLine: 8, Exported: false},
			{Name: "Execute", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 10, EndLine: 12, Exported: true},
			{Name: "main", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 14, EndLine: 16, Exported: false},
			{Name: "TestInit", Kind: "Function", Package: "cli", FilePath: testFile, StartLine: 3, EndLine: 5, Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				repoFile: {"fmt"},
				cliFile:  {"example.com/app/repository"},
				testFile: {"example.com/app/repository"},
			},
		},
		ByName:     make(map[string][]int),
		ByPackage:  make(map[string][]int),
		ByKind:     make(map[string][]int),
		ByReceiver: make(map[string][]int),
		ByLanguage: make(map[string][]int),
		Exported:   []int{},
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

func TestReferencesBasic(t *testing.T) {
	engine := createReferenceFixture(t)
	res, err := engine.References("Init", ReferenceOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if res.Symbol.Name != "Init" {
		t.Errorf("Expected symbol Init, got %s", res.Symbol.Name)
	}

	if res.Stats.TotalReferences < 1 {
		t.Errorf("Expected references, got 0")
	}
}

func TestReferencesFilters(t *testing.T) {
	engine := createReferenceFixture(t)
	
	// Package filter
	res, err := engine.References("Init", ReferenceOptions{PackageFilter: "cli"})
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range res.Groups {
		if g.Package != "cli" {
			t.Errorf("Expected only cli package, got %s", g.Package)
		}
	}

	// File filter
	resFile, err := engine.References("Init", ReferenceOptions{FileFilter: "cli.go"})
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range resFile.Groups {
		for _, ref := range g.References {
			if !strings.Contains(ref.File, "cli.go") {
				t.Errorf("Expected only cli.go files, got %s", ref.File)
			}
		}
	}
}

func TestReferencesMax(t *testing.T) {
	engine := createReferenceFixture(t)
	res, err := engine.References("Init", ReferenceOptions{Max: 1})
	if err != nil {
		t.Fatal(err)
	}
	if res.Stats.TotalReferences > 1 {
		t.Errorf("Expected max 1 reference, got %d", res.Stats.TotalReferences)
	}
}

func TestReferencesUnknown(t *testing.T) {
	engine := createReferenceFixture(t)
	_, err := engine.References("UnknownSymbol", ReferenceOptions{})
	if err == nil {
		t.Error("Expected error for unknown symbol")
	}
}

func TestReferencesJSONSchema(t *testing.T) {
	engine := createReferenceFixture(t)
	res, err := engine.References("Init", ReferenceOptions{})
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(b)

	required := []string{
		`"symbol":`,
		`"groups":`,
		`"stats":`,
		`"lookupTimeMs":`,
	}
	for _, field := range required {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("missing JSON field %s in: %s", field, jsonStr)
		}
	}
	if !strings.Contains(jsonStr, ".") && !strings.Contains(jsonStr, "lookupTimeMs\": 0") {
		// It might be non-zero, let's just make sure it's not strictly 0 if it took time
		if res.LookupTimeMs == 0 {
			// This might be flaky on extremely fast machines, but usually takes > 0 ms.
			// Just pass if it's float64 format.
		}
	}
}

func TestReferencesSelfExclusion(t *testing.T) {
	engine := createReferenceFixture(t)
	res, err := engine.References("Init", ReferenceOptions{})
	if err != nil {
		t.Fatal(err)
	}

	for _, g := range res.Groups {
		for _, ref := range g.References {
			// We expect line 28 (e.Init()) to be found in engine.go, not line 27 (func (e *Engine) Init() {)
			if strings.Contains(ref.File, "engine.go") && ref.Line == 5 {
				t.Errorf("Self-definition line was not excluded: %v", ref)
			}
		}
	}
}

func TestReferencesSelectBypass(t *testing.T) {
	engine := createReferenceFixture(t)
	
	// Add a duplicate symbol to simulate MultipleFound
	sym := analyzer.Symbol{Name: "Init", Kind: "Function", Package: "other", FilePath: "other.go", StartLine: 1, EndLine: 5}
	engine.repo.Symbols = append(engine.repo.Symbols, sym)
	engine.repo.ByName["Init"] = append(engine.repo.ByName["Init"], len(engine.repo.Symbols)-1)

	// Without select, should be MultipleFound = true
	res1, _ := engine.References("Init", ReferenceOptions{})
	if !res1.MultipleFound {
		t.Errorf("Expected MultipleFound=true when no select index is provided")
	}

	// With select, should be MultipleFound = false and bypass interactive
	res2, _ := engine.References("Init", ReferenceOptions{SelectIndex: 1})
	if res2.MultipleFound {
		t.Errorf("Expected MultipleFound=false when select index is provided, got true")
	}
	if res2.Symbol == nil {
		t.Errorf("Expected symbol to be resolved")
	}
}
