package navigation

import (
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
			{Name: "Repository", Kind: "Interface", Package: "core", FilePath: "core/repo.go", Exported: true, StartLine: 10, EndLine: 15},
			{Name: "SaveRepository", Kind: "Struct", Package: "core", FilePath: "core/repo.go", Exported: true, StartLine: 20, EndLine: 25},
			{Name: "Duplicate", Kind: "Struct", Package: "models", FilePath: "models/one.go", Exported: true},
			{Name: "Duplicate", Kind: "Struct", Package: "api", FilePath: "api/two.go", Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"core/repo.go": {"fmt", "strings"},
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
	}

	return &Engine{repo: repo}
}

func TestGoto(t *testing.T) {
	engine := createTestEngine(t)

	res, err := engine.Goto("Repository", 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.MultipleFound || res.Symbol == nil {
		t.Errorf("Expected exact match, got multiple=%v", res.MultipleFound)
	}
	if res.Symbol.Name != "Repository" {
		t.Errorf("Expected Repository, got %s", res.Symbol.Name)
	}

	resMulti, _ := engine.Goto("Duplicate", 0)
	if !resMulti.MultipleFound {
		t.Errorf("Expected multiple found for Duplicate")
	}

	resSelect, _ := engine.Goto("Duplicate", 1)
	if resSelect.MultipleFound || resSelect.Symbol.Package != "models" {
		t.Errorf("Expected selection of models.Duplicate")
	}
}

func TestOutline(t *testing.T) {
	engine := createTestEngine(t)
	
	// Add another Interface to test sorting
	engine.repo.Symbols = append(engine.repo.Symbols, analyzer.Symbol{
		Name: "ARepository", Kind: "Interface", Package: "core", FilePath: "core/repo.go", Exported: true, StartLine: 30, EndLine: 35,
	})
	engine.repo.ByName["ARepository"] = []int{len(engine.repo.Symbols) - 1}
	
	// 1. Basic Stats and Groups
	res, err := engine.Outline(filepath.Join("core", "repo.go"), "", "line")
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.TotalSymbols != 3 {
		t.Fatalf("Expected 3 total symbols, got %d", res.Stats.TotalSymbols)
	}
	
	if res.Stats.TotalImports != 2 {
		t.Fatalf("Expected 2 total imports, got %d", res.Stats.TotalImports)
	}

	if len(res.Groups) != 2 {
		t.Fatalf("Expected 2 groups (Interfaces, Structs), got %d", len(res.Groups))
	}

	// 2. Singular, Plural, Case-insensitive Filter
	// "interface" should match 2 interfaces
	resFilter, err := engine.Outline(filepath.Join("core", "repo.go"), "interface", "line")
	if err != nil {
		t.Fatal(err)
	}
	if resFilter.Stats.TotalSymbols != 2 {
		t.Fatalf("Expected 2 filtered symbol, got %d", resFilter.Stats.TotalSymbols)
	}
	
	// "STRUCTS" should match 1 struct
	resFilterPlural, err := engine.Outline(filepath.Join("core", "repo.go"), "STRUCTS", "line")
	if err != nil {
		t.Fatal(err)
	}
	if resFilterPlural.Stats.TotalSymbols != 1 {
		t.Fatalf("Expected 1 filtered symbol, got %d", resFilterPlural.Stats.TotalSymbols)
	}
	if resFilterPlural.Groups[0].Symbols[0].Name != "SaveRepository" {
		t.Errorf("Expected SaveRepository")
	}

	// 2.5 Invalid Filter
	_, errInvalid := engine.Outline(filepath.Join("core", "repo.go"), "xyz", "line")
	if errInvalid == nil {
		t.Fatal("Expected error for invalid filter")
	}
	if !strings.Contains(errInvalid.Error(), "Unknown filter \"xyz\"") {
		t.Fatalf("Expected 'Unknown filter' error, got %v", errInvalid)
	}

	// 3. Sort Name
	resSortName, err := engine.Outline(filepath.Join("core", "repo.go"), "", "name")
	if err != nil {
		t.Fatal(err)
	}
	
	// Find the Interfaces group
	var interfacesGroup *SymbolGroup
	for _, g := range resSortName.Groups {
		if g.Kind == "Interfaces" {
			// copy is necessary to take address in loop safely? no just take copy
			gCopy := g
			interfacesGroup = &gCopy
			break
		}
	}
	
	if interfacesGroup == nil {
		t.Fatalf("Expected Interfaces group")
	}
	
	if len(interfacesGroup.Symbols) != 2 {
		t.Fatalf("Expected 2 interfaces")
	}
	
	// Sort name: ARepository should come before Repository
	if interfacesGroup.Symbols[0].Name != "ARepository" {
		t.Errorf("Expected ARepository first, got %s", interfacesGroup.Symbols[0].Name)
	}

	// 4. Sort Line
	resSortLine, err := engine.Outline(filepath.Join("core", "repo.go"), "", "line")
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range resSortLine.Groups {
		if g.Kind == "Interfaces" {
			if g.Symbols[0].Name != "Repository" { // StartLine 10 vs 30
				t.Errorf("Expected Repository first by line, got %s", g.Symbols[0].Name)
			}
			break
		}
	}
}

func TestHierarchy(t *testing.T) {
	engine := createTestEngine(t)

	// Add Engine struct in two different packages
	engine.repo.Symbols = append(engine.repo.Symbols, analyzer.Symbol{
		Name: "Engine", Kind: "Struct", Package: "repository", FilePath: "repo.go",
	}, analyzer.Symbol{
		Name: "Engine", Kind: "Struct", Package: "graph", FilePath: "graph.go",
	})
	idx1 := len(engine.repo.Symbols) - 2
	idx2 := len(engine.repo.Symbols) - 1
	engine.repo.ByName["Engine"] = append(engine.repo.ByName["Engine"], idx1, idx2)
	engine.repo.ByPackage["repository"] = append(engine.repo.ByPackage["repository"], idx1)
	engine.repo.ByPackage["graph"] = append(engine.repo.ByPackage["graph"], idx2)

	// Add Methods for Engine
	engine.repo.Symbols = append(engine.repo.Symbols, analyzer.Symbol{
		Name: "Init", Kind: "Method", Package: "repository", FilePath: "repo.go",
	}, analyzer.Symbol{
		Name: "Analyze", Kind: "Method", Package: "graph", FilePath: "graph.go",
	})
	mIdx1 := len(engine.repo.Symbols) - 2
	mIdx2 := len(engine.repo.Symbols) - 1
	engine.repo.ByReceiver["Engine"] = append(engine.repo.ByReceiver["Engine"], mIdx1, mIdx2)

	// Select repository.Engine
	// Wait, Engine is just added, but what if there's multiple? We should find the exact index.
	// Actually, Goto/Hierarchy returns MultipleFound if we don't provide the right selectIndex.
	
	resMultiple, _ := engine.Hierarchy("Engine", 0)
	if !resMultiple.MultipleFound {
		t.Fatalf("Expected MultipleFound for Engine")
	}
	
	// Select repository.Engine which is index 1 of the candidates (0-based in selection usually? actually 1-based in CLI, but the method uses 1-based selection natively)
	// Let's pass 1
	resRepo, err := engine.Hierarchy("Engine", 1)
	if err != nil {
		t.Fatal(err)
	}

	if resRepo.Symbol.Package != "repository" {
		t.Fatalf("Expected package repository, got %s", resRepo.Symbol.Package)
	}

	if len(resRepo.Methods) != 1 {
		t.Fatalf("Expected 1 method for repository.Engine, got %d", len(resRepo.Methods))
	}
	
	if resRepo.Methods[0].Name != "Init" {
		t.Errorf("Expected Init method, got %s", resRepo.Methods[0].Name)
	}

	// Select graph.Engine
	resGraph, err := engine.Hierarchy("Engine", 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(resGraph.Methods) != 1 {
		t.Fatalf("Expected 1 method for graph.Engine, got %d", len(resGraph.Methods))
	}
	
	if resGraph.Methods[0].Name != "Analyze" {
		t.Errorf("Expected Analyze method, got %s", resGraph.Methods[0].Name)
	}
}

func TestPeek(t *testing.T) {
	engine := createTestEngine(t)
	// Create a dummy file for peek
	tmpFile, err := os.CreateTemp("", "repo.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	for i := 1; i <= 200; i++ {
		tmpFile.WriteString("line\n")
	}
	tmpFile.Close()

	engine.repo.Symbols[0].StartLine = 1
	engine.repo.Symbols[0].EndLine = 100
	engine.repo.Symbols[0].FilePath = tmpFile.Name()

	engine.repo.Symbols = append(engine.repo.Symbols, analyzer.Symbol{
		Name: "Init", Kind: "Method", Package: "core", FilePath: tmpFile.Name(), StartLine: 105, EndLine: 120,
	})
	// Link the method to the receiver "Repository"
	engine.repo.ByReceiver["Repository"] = []int{len(engine.repo.Symbols) - 1}

	tests := []struct {
		limit     int
		wantLines int
		truncated bool
	}{
		{10, 10, true},
		{20, 20, true},
		{40, 40, true},
		{110, 110, true},   // targetEnd = 110, contextEnd = 120 -> truncated
		{150, 150, false},  // targetEnd = 150, contextEnd = 120 -> fully fits
	}

	for _, tt := range tests {
		res, err := engine.Peek("Repository", 0, tt.limit)
		if err != nil {
			t.Fatalf("limit %d: %v", tt.limit, err)
		}
		
		lines := len(strings.Split(res.Snippet, "\n"))
		if lines != tt.wantLines {
			t.Errorf("limit %d: got %d lines, want %d", tt.limit, lines, tt.wantLines)
		}
		if res.Truncated != tt.truncated {
			t.Errorf("limit %d: got truncated %v, want %v (contextEnd: %v)", tt.limit, res.Truncated, tt.truncated, 120)
		}
	}
}

func BenchmarkPeek(b *testing.B) {
	// We'll benchmark on the engine initialized in tests
	engine := createTestEngine(&testing.T{})
	tmpFile, err := os.CreateTemp("", "repo.go")
	if err == nil {
		for i := 1; i <= 200; i++ {
			tmpFile.WriteString("line\n")
		}
		tmpFile.Close()
		engine.repo.Symbols[0].FilePath = tmpFile.Name()
		defer os.Remove(tmpFile.Name())
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Peek("Repository", 0, 40)
	}
}
