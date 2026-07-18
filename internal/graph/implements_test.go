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

func createImplementsFixture(t *testing.T) *Engine {
	t.Helper()
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "repository")
	os.MkdirAll(repoDir, 0755)

	repoFile := filepath.Join(repoDir, "repository.go")
	os.WriteFile(repoFile, []byte(`package repository

type Engine struct{}

// Partial implementation
func (e *Engine) Init() error { return nil }

type MemoryRepository struct{}

// Value receiver complete implementation
func (m MemoryRepository) Init() error { return nil }
func (m MemoryRepository) Save() {}
func (m MemoryRepository) Load() {}

type Repository interface {
	Init() error
	Save()
	Load()
}

type Storage interface {
	Repository
	Delete()
}

type MockRepository struct{}

// Pointer receiver complete implementation for Storage
func (m *MockRepository) Init() error { return nil }
func (m *MockRepository) Save() {}
func (m *MockRepository) Load() {}
func (m *MockRepository) Delete() {}
`), 0644)

	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Engine", Kind: "Struct", Package: "repository", FilePath: repoFile, StartLine: 3, EndLine: 3},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 6, EndLine: 6, Receiver: "*Engine"},
			
			{Name: "MemoryRepository", Kind: "Struct", Package: "repository", FilePath: repoFile, StartLine: 8, EndLine: 8},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 11, EndLine: 11, Receiver: "MemoryRepository"},
			{Name: "Save", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 12, EndLine: 12, Receiver: "MemoryRepository"},
			{Name: "Load", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 13, EndLine: 13, Receiver: "MemoryRepository"},
			
			{Name: "Repository", Kind: "Interface", Package: "repository", FilePath: repoFile, StartLine: 15, EndLine: 19},
			{Name: "Storage", Kind: "Interface", Package: "repository", FilePath: repoFile, StartLine: 21, EndLine: 24},
			
			{Name: "MockRepository", Kind: "Struct", Package: "repository", FilePath: repoFile, StartLine: 26, EndLine: 26},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 29, EndLine: 29, Receiver: "*MockRepository"},
			{Name: "Save", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 30, EndLine: 30, Receiver: "*MockRepository"},
			{Name: "Load", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 31, EndLine: 31, Receiver: "*MockRepository"},
			{Name: "Delete", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 32, EndLine: 32, Receiver: "*MockRepository"},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{},
		},
		ByName:     make(map[string][]int),
		ByPackage:  make(map[string][]int),
		ByReceiver: make(map[string][]int),
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

	return engine
}

func TestImplementsBasic(t *testing.T) {
	engine := createImplementsFixture(t)
	res, err := engine.Implements("Repository", ImplementsOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Implementations) != 2 {
		t.Fatalf("Expected 2 implementations, got %d", len(res.Implementations))
	}
	
	foundMemory := false
	foundMock := false
	
	for _, impl := range res.Implementations {
		if impl.Struct == "Engine" {
			t.Errorf("Engine should NOT be an implementation because it is partial")
		}
		if impl.Struct == "MemoryRepository" {
			foundMemory = true
			if impl.PointerReceiver {
				t.Errorf("MemoryRepository uses value receivers, should be pointerReceiver=false")
			}
		}
		if impl.Struct == "MockRepository" {
			foundMock = true
			if !impl.PointerReceiver {
				t.Errorf("MockRepository uses pointer receivers, should be pointerReceiver=true")
			}
		}
	}
	
	if !foundMemory || !foundMock {
		t.Errorf("Missing expected implementations")
	}
}

func TestImplementsEmbedded(t *testing.T) {
	engine := createImplementsFixture(t)
	res, err := engine.Implements("Storage", ImplementsOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Implementations) != 1 {
		t.Fatalf("Expected 1 implementation, got %d", len(res.Implementations))
	}
	
	impl := res.Implementations[0]
	if impl.Struct != "MockRepository" {
		t.Errorf("Expected MockRepository to implement Storage, got %s", impl.Struct)
	}
	if !impl.PointerReceiver {
		t.Errorf("MockRepository uses pointer receivers, should be pointerReceiver=true")
	}
	
	if len(impl.ImplementedMethods) != 4 {
		t.Errorf("Expected 4 methods, got %d", len(impl.ImplementedMethods))
	}
}

func TestImplementsPackageFilter(t *testing.T) {
	engine := createImplementsFixture(t)
	res, err := engine.Implements("Repository", ImplementsOptions{PackageFilter: "repository"})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Implementations) != 2 {
		t.Errorf("Expected 2 implementations in package repository, got %d", len(res.Implementations))
	}
}

func TestImplementsMax(t *testing.T) {
	engine := createImplementsFixture(t)
	res, err := engine.Implements("Repository", ImplementsOptions{Max: 1})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Implementations) != 1 {
		t.Errorf("Expected max 1 implementation, got %d", len(res.Implementations))
	}
}

func TestImplementsJSONSchema(t *testing.T) {
	engine := createImplementsFixture(t)
	res, err := engine.Implements("Repository", ImplementsOptions{})
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
		`"requiredMethods":`,
		`"implementations":`,
		`"struct":`,
		`"package":`,
		`"pointerReceiver":`,
		`"implementedMethods":`,
		`"lookupTimeMs":`,
	}
	for _, field := range required {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("missing JSON field %s in: %s", field, jsonStr)
		}
	}
	
	if res.LookupTimeMs <= 0 {
		t.Errorf("Expected LookupTimeMs to be > 0, got %f", res.LookupTimeMs)
	}

	if !strings.Contains(jsonStr, ".") && !strings.Contains(jsonStr, "e-") {
		t.Errorf("Expected lookupTimeMs to be serialized as a floating-point value, got: %s", jsonStr)
	}
}

func TestImplementsSelectBypass(t *testing.T) {
	engine := createImplementsFixture(t)
	
	sym := analyzer.Symbol{Name: "Repository", Kind: "Interface", Package: "other", FilePath: "other.go", StartLine: 1, EndLine: 5}
	engine.repo.Symbols = append(engine.repo.Symbols, sym)
	engine.repo.ByName["Repository"] = append(engine.repo.ByName["Repository"], len(engine.repo.Symbols)-1)

	res1, _ := engine.Implements("Repository", ImplementsOptions{})
	if !res1.MultipleFound {
		t.Errorf("Expected MultipleFound=true when no select index is provided")
	}

	res2, _ := engine.Implements("Repository", ImplementsOptions{SelectIndex: 1})
	if res2.MultipleFound {
		t.Errorf("Expected MultipleFound=false when select index is provided, got true")
	}
	if res2.Symbol == nil {
		t.Errorf("Expected symbol to be resolved")
	}
}

func TestImplementsUnknown(t *testing.T) {
	engine := createImplementsFixture(t)
	_, err := engine.Implements("xyz", ImplementsOptions{})
	if err == nil {
		t.Errorf("Expected error for unknown interface")
	}
}
