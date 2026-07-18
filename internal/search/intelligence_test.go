package search

import (
	"testing"
	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func TestIntelligenceFeatures(t *testing.T) {
	repo := &InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Repository", Kind: "Interface", Package: "core", FilePath: "core/repo.go", Exported: true},
			{Name: "RepositoryManager", Kind: "Struct", Package: "core", FilePath: "core/repo.go", Exported: true},
			{Name: "SaveRepository", Kind: "Function", Package: "core", FilePath: "core/repo.go", Exported: true},
			{Name: "OtherThing", Kind: "Struct", Package: "models", FilePath: "models/other.go", Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"cmd/main.go": {"core"},
				"api/server.go": {"core", "models"},
				"utils/helper.go": {"models"},
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
	}

	engine := NewKeywordSearch(repo)

	// Test Definition
	defs, err := engine.Definition("Repository")
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) == 0 || defs[0].Name != "Repository" {
		t.Error("failed to find exact definition for Repository")
	}

	// Test References
	refs, err := engine.References("Repository")
	if err != nil {
		t.Fatal(err)
	}
	// "cmd/main.go" and "api/server.go" import "core"
	if len(refs) != 2 {
		t.Errorf("expected 2 references for Repository, got %d: %v", len(refs), refs)
	}

	// Test Related Symbols
	related, err := engine.Related("Repository")
	if err != nil {
		t.Fatal(err)
	}
	// "RepositoryManager" and "SaveRepository" contain "Repository"
	if len(related) != 2 {
		t.Errorf("expected 2 related symbols for Repository, got %d: %v", len(related), related)
	}
}
