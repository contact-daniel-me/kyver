package search

import "strings"

func (s *KeywordSearch) Definition(symbol string) ([]Result, error) {
	var results []Result
	target := strings.ToLower(symbol)

	indices := getIndicesIgnoreCase(s.repo.ByName, target)

	for _, idx := range indices {
		sym := s.repo.Symbols[idx]
		if strings.ToLower(sym.Name) == target {
			score := 100.0
			if sym.Exported {
				score += 50.0
			}
			results = append(results, NewResult(sym, 100.0, "Definition exact match"))
		}
	}
	
	RankResults(results)
	return results, nil
}
