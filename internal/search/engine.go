package search

type KeywordSearch struct {
	repo *InMemoryRepository
}

func NewKeywordSearch(repo *InMemoryRepository) *KeywordSearch {
	return &KeywordSearch{repo: repo}
}

func (s *KeywordSearch) Search(q Query) ([]Result, error) {
	var results []Result

	// Use lookup indexes to rapidly narrow down search space
	candidateIndices := ApplyFilters(q, s.repo)

	// Evaluate remaining candidates
	for _, idx := range candidateIndices {
		sym := s.repo.Symbols[idx]
		
		score, reason := CalculateScore(q.Text, sym)
		
		if score > 0 {
			results = append(results, NewResult(sym, score, reason))
		}
	}

	RankResults(results)

	if q.Limit > 0 && len(results) > q.Limit {
		results = results[:q.Limit]
	}

	return results, nil
}
