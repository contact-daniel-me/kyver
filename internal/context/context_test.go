package context

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/graph"
	"github.com/contact-daniel-me/kyver/internal/navigation"
	"github.com/contact-daniel-me/kyver/internal/search"
)

func createTestEngine() *Engine {
	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Engine", Kind: "Struct", Package: "context", FilePath: "builder.go", StartLine: 1, EndLine: 100, Exported: true, Documentation: "ContextEngine is a semantic aggregation layer."},
			{Name: "BuildSymbolContext", Kind: "Method", Package: "context", FilePath: "builder.go", Exported: true, Receiver: "Engine"},
			{Name: "ContextBuilder", Kind: "Interface", Package: "context", FilePath: "interfaces.go", Exported: true},
			{Name: "BuildSymbolContext", Kind: "Method", Package: "context", FilePath: "interfaces.go", Exported: true, Receiver: "ContextBuilder"},
			{Name: "CallersHelper", Kind: "Function", Package: "testpkg", FilePath: "helper.go"},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"builder.go":    {"fmt", "time"},
				"interfaces.go": {"github.com/contact-daniel-me/kyver/internal/analyzer"},
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
		repo.ByKind[sym.Kind] = append(repo.ByKind[sym.Kind], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
		if sym.Exported {
			repo.Exported = append(repo.Exported, i)
		}
	}

	navEngine := navigation.NewEngineWithRepo(repo)
	graphEngine := graph.NewEngineWithRepo(repo)

	return NewEngineWithEngines(navEngine, graphEngine)
}

func TestContextSummary(t *testing.T) {
	e := createTestEngine()

	opts := ContextOptions{
		Summary: true,
	}

	res, err := e.BuildSymbolContext("Engine", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Symbol == nil || res.Symbol.Name != "Engine" {
		t.Fatalf("expected symbol Engine")
	}

	if res.Documentation == "" {
		t.Errorf("expected documentation to be included in summary")
	}

	if res.Statistics.Methods != 1 {
		t.Errorf("expected 1 method in statistics, got %d", res.Statistics.Methods)
	}
}

func TestContextFull(t *testing.T) {
	e := createTestEngine()

	opts := ContextOptions{
		Full: true,
	}

	res, err := e.BuildSymbolContext("Engine", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Methods) != 1 {
		t.Errorf("expected methods to be populated in full mode")
	}
}

func TestContextSections(t *testing.T) {
	e := createTestEngine()

	opts := ContextOptions{
		Sections: []string{"methods"},
	}

	res, err := e.BuildSymbolContext("Engine", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Methods) != 1 {
		t.Errorf("expected methods to be populated")
	}

	// other sections should be empty
	if res.Documentation != "" {
		t.Errorf("expected documentation to be omitted")
	}
}

func TestContextJSON(t *testing.T) {
	e := createTestEngine()

	opts := ContextOptions{
		Full: true,
	}

	res, _ := e.BuildSymbolContext("Engine", opts)

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Verify schema fields
	s := string(b)
	if !strings.Contains(s, `"methods"`) || !strings.Contains(s, `"statistics"`) || !strings.Contains(s, `"lookupTimeMs"`) {
		t.Errorf("JSON schema is missing required fields: %s", s)
	}
}
