package search

import (
	"testing"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func TestLevenshteinDistance(t *testing.T) {
	cases := []struct {
		s1, s2 string
		dist   int
	}{
		{"kitten", "sitting", 3},
		{"Authentcation", "Authentication", 1},
		{"", "abc", 3},
		{"abc", "", 3},
		{"IdxRepo", "IndexRepository", 8},
	}

	for _, c := range cases {
		d := LevenshteinDistance(c.s1, c.s2)
		if d != c.dist {
			t.Errorf("LevenshteinDistance(%q, %q) == %d, expected %d", c.s1, c.s2, d, c.dist)
		}
	}
}

func TestIsFuzzyMatch(t *testing.T) {
	if !IsFuzzyMatch("IdxRepo", "IndexRepository") {
		t.Error("expected IdxRepo to fuzzy match IndexRepository (via subsequence)")
	}
	if !IsFuzzyMatch("Authentcation", "Authentication") {
		t.Error("expected Authentcation to fuzzy match Authentication (via edit distance)")
	}
	if IsFuzzyMatch("xy", "IndexRepository") {
		t.Error("xy should not match IndexRepository")
	}
}

func TestFiltersAndRanking(t *testing.T) {
	repo := &InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "User", Kind: "Struct", Package: "models", FilePath: "models/user.go", Exported: true, Documentation: "Represents a user."},
			{Name: "UserRepository", Kind: "Struct", Package: "repository", FilePath: "repo/user.go", Exported: true},
			{Name: "FetchUser", Kind: "Function", Package: "models", FilePath: "models/user.go", Exported: true},
			{Name: "internalHelper", Kind: "Function", Package: "models", FilePath: "models/helper.go", Exported: false},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				"models/user.go": {"fmt", "time"},
				"repo/user.go":   {"database/sql"},
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
		repo.ByKind[sym.Kind] = append(repo.ByKind[sym.Kind], i)
		if sym.Exported {
			repo.Exported = append(repo.Exported, i)
		}
	}

	engine := NewKeywordSearch(repo)

	res, err := engine.Search(Query{Text: "User"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) < 3 {
		t.Fatalf("expected at least 3 results for User, got %d", len(res))
	}
	if res[0].Name != "User" {
		t.Errorf("expected exact match 'User' to be ranked first, got %s", res[0].Name)
	}

	res, err = engine.Search(Query{Text: "User", Kind: "Function"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Name != "FetchUser" {
		t.Errorf("expected FetchUser, got %+v", res)
	}

	res, err = engine.Search(Query{Text: "internal", Exported: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 0 {
		t.Errorf("expected 0 results, got %d", len(res))
	}

	res, err = engine.Search(Query{Text: "User", Imports: "time"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Errorf("expected 2 results from models/user.go importing time, got %d", len(res))
	}
}
