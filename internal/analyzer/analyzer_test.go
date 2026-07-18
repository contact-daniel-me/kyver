package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

func TestEngine_Analyze(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create .kyver dir for artifacts
	os.MkdirAll(filepath.Join(tmpDir, ".kyver"), 0755)

	// Create test file
	filePath := filepath.Join(tmpDir, "main.go")
	code := `package main
import "fmt"
type MyStruct struct{}
func (m *MyStruct) Hello() {}
func main() {}
`
	os.WriteFile(filePath, []byte(code), 0644)

	idxData := &indexer.IndexData{
		Files: []indexer.FileMetadata{
			{Path: "main.go", Language: "Go", Size: int64(len(code))},
		},
	}

	engine := NewEngine(tmpDir, "v1.0.0")
	summary, err := engine.Analyze(idxData, "", false)
	
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if summary.FilesAnalyzed != 1 {
		t.Errorf("expected 1 file analyzed, got %d", summary.FilesAnalyzed)
	}
	if summary.TotalSymbols != 3 { // MyStruct, Hello, main
		t.Errorf("expected 3 symbols, got %d", summary.TotalSymbols)
	}
	if summary.ArtifactsSize == 0 {
		t.Errorf("expected generated artifacts size > 0")
	}
}

func TestGoAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.go")

	code := `package testpkg

import "fmt"

// MyStruct is a test struct
type MyStruct struct {
	Field int
}

// Hello prints hello
func (m *MyStruct) Hello() {
	fmt.Println("Hello")
}

func GlobalFunc() {}

const myConst = 1
var myVar = 2
type MyAlias = int
`
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	a := &GoAnalyzer{}
	meta := indexer.FileMetadata{Path: filePath, Language: "Go", Size: int64(len(code))}

	res, err := a.Analyze(filePath, meta)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if res.PackageName != "testpkg" {
		t.Errorf("expected package testpkg, got %s", res.PackageName)
	}

	if len(res.Imports) != 1 || res.Imports[0] != "fmt" {
		t.Errorf("expected import fmt, got %v", res.Imports)
	}

	if len(res.Symbols) != 6 {
		t.Fatalf("expected 6 symbols, got %d", len(res.Symbols))
	}

	foundStruct := false
	foundMethod := false
	for _, sym := range res.Symbols {
		if sym.Name == "MyStruct" && sym.Kind == "Struct" {
			foundStruct = true
			if sym.Documentation != "MyStruct is a test struct" {
				t.Errorf("wrong doc for struct: %s", sym.Documentation)
			}
		}
		if sym.Name == "Hello" && sym.Kind == "Method" {
			foundMethod = true
			if sym.Receiver != "*MyStruct" {
				t.Errorf("expected receiver *MyStruct, got %s", sym.Receiver)
			}
		}
	}

	if !foundStruct || !foundMethod {
		t.Error("failed to extract required symbols")
	}

	if res.Metrics.Functions != 1 || res.Metrics.Methods != 1 || res.Metrics.Structs != 1 {
		t.Errorf("incorrect metrics: %+v", res.Metrics)
	}
	
	if res.Metrics.Constants != 1 || res.Metrics.Variables != 1 || res.Metrics.Aliases != 1 {
		t.Errorf("incorrect variable metrics: %+v", res.Metrics)
	}
}

func TestArchitectureSummary(t *testing.T) {
	deps := NewDependencyGraph("v0.1.0")
	deps.AddPackageDependency("main", []string{"pkgA", "pkgB"})
	deps.AddPackageDependency("pkgA", []string{"pkgC"})
	deps.AddPackageDependency("pkgC", []string{"pkgA"})

	arch := GenerateArchitectureSummary("v0.1.0", deps)

	if len(arch.EntryPackages) != 1 || arch.EntryPackages[0] != "main" {
		t.Error("expected main as entry package")
	}

	if len(arch.PotentialCyclic) == 0 {
		t.Error("expected to detect cyclic dependency")
	}
}

func BenchmarkAnalyze(b *testing.B) {
	tmpDir := b.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".kyver"), 0755)

	filePath := filepath.Join(tmpDir, "main.go")
	code := `package main
func main() {}
`
	os.WriteFile(filePath, []byte(code), 0644)

	idxData := &indexer.IndexData{
		Files: []indexer.FileMetadata{
			{Path: "main.go", Language: "Go", Size: int64(len(code))},
		},
	}

	engine := NewEngine(tmpDir, "v1.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Analyze(idxData, "", false)
	}
}
