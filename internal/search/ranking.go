package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

// CalculateScore computes the relevance of a symbol to the query text
func CalculateScore(query string, sym analyzer.Symbol) (float64, string) {
	if query == "" {
		return 1.0, ""
	}

	q := strings.ToLower(query)
	name := strings.ToLower(sym.Name)
	pkg := strings.ToLower(sym.Package)
	doc := strings.ToLower(sym.Documentation)

	// Precedence 1: Exact symbol name match
	if name == q {
		return 100.0, "Exact symbol match"
	}
	
	// Precedence 2: Exact fully qualified symbol match (e.g., repository.Engine)
	if fmt.Sprintf("%s.%s", pkg, name) == q {
		return 95.0, "Exact fully qualified match"
	}
	
	// Precedence 3: Exact package match
	if pkg == q {
		return 90.0, "Exact package match"
	}
	
	// Precedence 4: Prefix match
	if strings.HasPrefix(name, q) {
		return 80.0, "Prefix match"
	}
	
	// Precedence 5: Whole-word match
	if IsWholeWordMatch(query, sym.Name) {
		return 70.0, "Whole-word match"
	}
	
	// Precedence 6: CamelCase abbreviation match
	if IsCamelCaseAbbr(query, sym.Name) {
		return 60.0, "CamelCase abbreviation match"
	}
	
	// Precedence 7: Subsequence match
	if IsSubsequence(query, sym.Name) {
		return 50.0, "Subsequence match"
	}
	
	// Precedence 8: Fuzzy match
	if IsFuzzyMatch(query, sym.Name) {
		return 40.0, "Fuzzy match"
	}
	
	// Precedence 9: Documentation match
	if strings.Contains(doc, q) {
		return 10.0, "Documentation match"
	}

	return 0.0, ""
}

// RankResults sorts the results by score descending
func RankResults(results []Result) {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}
