package search

import (
	"sort"
	"strings"
)

func (s *KeywordSearch) Related(symbol string) ([]Result, error) {
	if symbol == "" {
		return nil, nil
	}

	target := strings.ToLower(symbol)
	var related []Result
	seen := make(map[string]bool)

	for name, indices := range s.repo.ByName {
		lowerName := strings.ToLower(name)
		if lowerName == target {
			continue
		}

		// Check for substring containment to identify related symbols
		if strings.Contains(lowerName, target) {
			if !seen[name] && len(indices) > 0 {
				idx := indices[0] // take first definition
				sym := s.repo.Symbols[idx]
				dist := LevenshteinDistance(target, lowerName)
				// We invert distance for score so higher is better, or just use -dist
				related = append(related, NewResult(sym, float64(-dist), "Related symbol"))
				seen[name] = true
			}
		}
	}

	// Sort by relevance (highest score first)
	sort.SliceStable(related, func(i, j int) bool {
		return related[i].Score > related[j].Score
	})

	if len(related) > 10 {
		related = related[:10]
	}

	return related, nil
}
