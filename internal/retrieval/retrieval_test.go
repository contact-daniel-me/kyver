package retrieval

import (
	"testing"
)

func TestParseKeywords(t *testing.T) {
	query := "How does indexing work?"
	keywords := parseKeywords(query)

	if len(keywords) != 1 {
		t.Fatalf("Expected 1 keyword, got %d", len(keywords))
	}

	if keywords[0] != "indexing" {
		t.Errorf("Expected 'indexing', got '%s'", keywords[0])
	}

	query2 := "What is the Repository interface?"
	keywords2 := parseKeywords(query2)

	// 'What', 'is', 'the' should be dropped. 'Repository', 'interface' should be kept.
	if len(keywords2) != 2 {
		t.Fatalf("Expected 2 keywords, got %d", len(keywords2))
	}
}

func TestCalculateScores(t *testing.T) {
	exactSym := calculateSymbolScore("Index", "Index")
	if exactSym != 0.95 {
		t.Errorf("Expected 0.95, got %f", exactSym)
	}

	partialSym := calculateSymbolScore("index", "indexer")
	if partialSym != 0.75 {
		t.Errorf("Expected 0.75, got %f", partialSym)
	}

	noMatchSym := calculateSymbolScore("foo", "bar")
	if noMatchSym != 0.5 {
		t.Errorf("Expected 0.5, got %f", noMatchSym)
	}

	exactPkg := calculatePackageScore("parser", "parser")
	if exactPkg != 0.90 {
		t.Errorf("Expected 0.90, got %f", exactPkg)
	}
}
