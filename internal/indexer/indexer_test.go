package indexer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexer(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "kyver-indexer-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go file
	err = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create an ignored directory
	ignoredDir := filepath.Join(tempDir, "node_modules")
	err = os.Mkdir(ignoredDir, 0755)
	if err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}
	err = os.WriteFile(filepath.Join(ignoredDir, "index.js"), []byte("console.log('hi');"), 0644)
	if err != nil {
		t.Fatalf("failed to write index.js: %v", err)
	}

	// Create an unindexed file (e.g., .txt)
	err = os.WriteFile(filepath.Join(tempDir, "notes.txt"), []byte("some notes"), 0644)
	if err != nil {
		t.Fatalf("failed to write notes.txt: %v", err)
	}

	// Run Indexer
	idx := New(tempDir)
	res, err := idx.Index(false)
	if err != nil {
		t.Fatalf("indexer failed: %v", err)
	}

	if res.FilesIndexed != 1 {
		t.Errorf("expected 1 file indexed, got %d", res.FilesIndexed)
	}

	// Verify the output file
	indexPath := res.IndexPath
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index.json: %v", err)
	}

	var indexData IndexData
	if err := json.Unmarshal(data, &indexData); err != nil {
		t.Fatalf("failed to parse index.json: %v", err)
	}

	if len(indexData.Files) != 1 {
		t.Fatalf("expected 1 file in JSON, got %d", len(indexData.Files))
	}

	meta := indexData.Files[0]
	if meta.Name != "main.go" {
		t.Errorf("expected file name main.go, got %s", meta.Name)
	}
	if meta.Language != "Go" {
		t.Errorf("expected language Go, got %s", meta.Language)
	}
	if meta.Module != "main" {
		t.Errorf("expected module main, got %s", meta.Module)
	}
}

func TestGetLanguage(t *testing.T) {
	if lang := GetLanguage("server.go"); lang != "Go" {
		t.Errorf("expected Go, got %s", lang)
	}
	if lang := GetLanguage("App.tsx"); lang != "TypeScript" {
		t.Errorf("expected TypeScript, got %s", lang)
	}
	if lang := GetLanguage("Dockerfile"); lang != "Docker" {
		t.Errorf("expected Docker, got %s", lang)
	}
	if lang := GetLanguage("unknown.xyz"); lang != "" {
		t.Errorf("expected empty string, got %s", lang)
	}
}

func TestIndexer_CorruptedDir(t *testing.T) {
	tmpDir := t.TempDir()
	kyverDir := filepath.Join(tmpDir, ".kyver")
	
	// Create a file named .kyver instead of a directory to simulate permission/IO error
	os.WriteFile(kyverDir, []byte("not a dir"), 0644)

	idx := New(tmpDir)
	_, err := idx.Index(false)
	if err == nil {
		t.Fatalf("expected error when writing index to corrupted .kyver directory")
	}
}
