package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)

func TestEngine_Interface(t *testing.T) {
	tmp := t.TempDir()
	structsFile := filepath.Join(tmp, "structs.go")
	ioFile := filepath.Join(tmp, "io.go")

	os.WriteFile(structsFile, []byte(`package testpkg
type TargetStruct struct{}
func (TargetStruct) Read() {}
func (*TargetStruct) Write() {}
type OtherStruct struct{}
`), 0644)
	
	os.WriteFile(ioFile, []byte(`package io
type ReadWriter interface {
	Read()
	Write()
}
type Reader interface {
	Read()
}
`), 0644)

	mockRepo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "TargetStruct", Kind: "Struct", Package: "testpkg", FilePath: structsFile, StartLine: 2, EndLine: 2},
			{Name: "OtherStruct", Kind: "Struct", Package: "otherpkg", FilePath: structsFile, StartLine: 5, EndLine: 5},
			{Name: "ReadWriter", Kind: "Interface", Package: "io", FilePath: ioFile, StartLine: 2, EndLine: 5},
			{Name: "Reader", Kind: "Interface", Package: "io", FilePath: ioFile, StartLine: 6, EndLine: 8},
			{Name: "Read", Kind: "Method", Package: "testpkg", FilePath: structsFile, StartLine: 3, EndLine: 3},
			{Name: "Write", Kind: "Method", Package: "testpkg", FilePath: structsFile, StartLine: 4, EndLine: 4},
		},
		ByReceiver: map[string][]int{
			"TargetStruct":  {4}, // Read
			"*TargetStruct": {5}, // Write
			"OtherStruct":   {},
		},
		ByPackage: map[string][]int{
			"testpkg":  {0, 4, 5},
			"otherpkg": {1},
			"io":       {2, 3},
		},
		ByName: map[string][]int{
			"TargetStruct": {0},
			"OtherStruct":  {1},
			"ReadWriter":   {2},
			"Reader":       {3},
			"Read":         {4},
			"Write":        {5},
		},
	}

	for i := range mockRepo.Symbols {
		mockRepo.Symbols[i].ID = "id" + string(rune(i))
	}

	engine := &Engine{
		repo: mockRepo,
	}

	// Because extractInterfaceMethods needs to read files, we'll mock it internally or via the getFileLines callback inside Engine.Interface.
	// Oh wait, getFileLines actually uses os.ReadFile. For the test, we don't have those files on disk.
	// Since extractInterfaceMethods reads the file to find methods, if we don't have the file, it will return 0 methods.
	// An empty interface is implemented by everyone.
	
	// Wait, to properly test extractInterfaceMethods, we should use a temporary directory.
	// But it's simpler to just let it return 0 methods, meaning every struct implements it.
	
	t.Run("Basic Functionality", func(t *testing.T) {
		res, err := engine.Interface("TargetStruct", InterfaceOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if res.MultipleFound {
			t.Fatalf("did not expect multiple found")
		}
		
		if res.Symbol.Name != "TargetStruct" {
			t.Errorf("Expected symbol TargetStruct, got %s", res.Symbol.Name)
		}
		
		// Because we didn't mock file contents, interfaces have 0 required methods.
		// Therefore TargetStruct implements ReadWriter and Reader.
		if len(res.Interfaces) != 2 {
			t.Errorf("Expected 2 interfaces, got %d", len(res.Interfaces))
		}
		
		// Wait, if it has 0 required methods, does it implement as value or pointer?
		// hasAllMethods([]string{"Read"}, []string{}) -> true for value!
		// So PointerReceiver should be false for both.
		for _, iface := range res.Interfaces {
			if iface.Name == "ReadWriter" && !iface.PointerReceiver {
				t.Errorf("Expected PointerReceiver to be true for ReadWriter")
			}
			if iface.Name == "Reader" && iface.PointerReceiver {
				t.Errorf("Expected PointerReceiver to be false for Reader")
			}
		}
		
		if res.Stats.TotalInterfaces != 2 {
			t.Errorf("Expected 2 total interfaces in stats, got %d", res.Stats.TotalInterfaces)
		}
		if res.Stats.TotalPackages != 1 {
			t.Errorf("Expected 1 package in stats, got %d", res.Stats.TotalPackages)
		}
		
		if res.LookupTimeMs <= 0 {
			t.Errorf("Expected LookupTimeMs > 0, got %v", res.LookupTimeMs)
		}
	})

	t.Run("Unknown Symbol", func(t *testing.T) {
		_, err := engine.Interface("NonExistent", InterfaceOptions{})
		if err == nil {
			t.Errorf("Expected error for non-existent symbol")
		}
	})

	t.Run("Not a Struct", func(t *testing.T) {
		_, err := engine.Interface("ReadWriter", InterfaceOptions{})
		if err == nil {
			t.Errorf("Expected error when targeting an Interface instead of Struct")
		} else if !strings.Contains(err.Error(), "not a Struct") {
			t.Errorf("Expected 'not a Struct' error, got %v", err)
		}
	})
	
	t.Run("Package Filter", func(t *testing.T) {
		res, err := engine.Interface("TargetStruct", InterfaceOptions{PackageFilter: "xyz"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if len(res.Interfaces) != 0 {
			t.Errorf("Expected 0 interfaces due to package filter, got %d", len(res.Interfaces))
		}
	})
	
	t.Run("Max Filter", func(t *testing.T) {
		res, err := engine.Interface("TargetStruct", InterfaceOptions{Max: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if len(res.Interfaces) != 1 {
			t.Errorf("Expected 1 interface due to max filter, got %d", len(res.Interfaces))
		}
	})

	t.Run("Sort By Name", func(t *testing.T) {
		res, err := engine.Interface("TargetStruct", InterfaceOptions{SortBy: "name"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if len(res.Interfaces) != 2 {
			t.Fatalf("Expected 2 interfaces, got %d", len(res.Interfaces))
		}
		if res.Interfaces[0].Name != "ReadWriter" {
			t.Errorf("Expected ReadWriter to be first when sorted by name, got %s", res.Interfaces[0].Name)
		}
	})
	
	t.Run("JSON Serialization", func(t *testing.T) {
		res, err := engine.Interface("TargetStruct", InterfaceOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Force non-zero sleep if it's too fast? engine sets it to math.Max(..., 0.001) anyway.
		time.Sleep(1 * time.Millisecond)
		
		b, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("JSON marshal error: %v", err)
		}
		
		jsonStr := string(b)
		if !strings.Contains(jsonStr, `"lookupTimeMs"`) {
			t.Errorf("Expected lookupTimeMs in JSON, got %s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"interfaces"`) {
			t.Errorf("Expected interfaces in JSON, got %s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"pointerReceiver"`) {
			t.Errorf("Expected pointerReceiver in JSON, got %s", jsonStr)
		}
		
		// Check that it's a float
		if !strings.Contains(jsonStr, ".") && !strings.Contains(jsonStr, "e-") {
			t.Errorf("Expected lookupTimeMs to be serialized as a floating-point value, got %s", jsonStr)
		}
	})
}
