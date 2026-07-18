package status

import (
	"testing"
	"time"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

func TestCalculateStats(t *testing.T) {
	data := &indexer.IndexData{
		Repository: "test-repo",
		IndexedAt:  time.Now(),
		Files: []indexer.FileMetadata{
			{Path: "main.go", Name: "main.go", Extension: ".go", Language: "Go", Size: 300, LastModified: time.Now()},
			{Path: "internal/app.go", Name: "app.go", Extension: ".go", Language: "Go", Size: 600, LastModified: time.Now().Add(-time.Hour)},
			{Path: "docs/README.md", Name: "README.md", Extension: ".md", Language: "Markdown", Size: 100, LastModified: time.Now()},
		},
	}

	stats := CalculateStats(data)

	if stats.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", stats.TotalFiles)
	}

	if stats.RepositorySize != 1000 {
		t.Errorf("expected 1000 bytes, got %d", stats.RepositorySize)
	}

	if stats.TotalLOC != 33 { // 1000 / 30
		t.Errorf("expected 33 LOC, got %d", stats.TotalLOC)
	}

	if len(stats.Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(stats.Languages))
	}

	if stats.Languages[0].Language != "Go" || stats.Languages[0].Count != 2 {
		t.Errorf("expected Go: 2, got %s: %d", stats.Languages[0].Language, stats.Languages[0].Count)
	}
}

func TestCheckHealth(t *testing.T) {
	data := &indexer.IndexData{
		Files: []indexer.FileMetadata{
			{Path: "main.go", Name: "main.go", Size: 500},
			{Path: "empty.go", Name: "empty.go", Size: 0},
			{Path: "big.bin", Name: "big.bin", Size: 15 * 1024 * 1024},
		},
	}

	recs := CheckHealth("/non-existent-dir", data)

	foundEmpty := false
	foundLarge := false
	foundReadme := false

	for _, r := range recs {
		if r == "Repository contains 1 empty files." {
			foundEmpty = true
		}
		if r == "1 file(s) exceed 10MB." {
			foundLarge = true
		}
		if r == "Consider adding a README.md" {
			foundReadme = true
		}
	}

	if !foundEmpty {
		t.Error("expected empty file recommendation")
	}
	if !foundLarge {
		t.Error("expected large file recommendation")
	}
	if !foundReadme {
		t.Error("expected missing README recommendation")
	}
}
