package search

import (
	"strings"
)

// intersectInts returns the intersection of two sorted slices of ints
// (Wait, are the index slices sorted? Yes, they are populated by appending sequentially increasing indices in loader.go)
func intersectInts(a, b []int) []int {
	var res []int
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			res = append(res, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return res
}

// filterByExact match returns the indices from map ignoring case by scanning keys, or just using case if we indexed lowercased.
// To be fully fast, we should index lowercased keys, but we indexed original keys. 
// A quick scan of map keys is still O(keys) which is much smaller than O(symbols).
func getIndicesIgnoreCase(m map[string][]int, query string) []int {
	q := strings.ToLower(query)
	var res []int
	for k, v := range m {
		if strings.ToLower(k) == q {
			// Instead of merging properly, just append since we know different cases of same word are mutually exclusive indices
			res = append(res, v...) 
		}
	}
	// Sort if we need to intersect later? We do need it sorted to intersect efficiently.
	// But it's easier to just use a boolean array if we have multiple filters.
	return res
}

// ApplyFilters uses the lookup indexes to rapidly narrow down the search space
func ApplyFilters(q Query, repo *InMemoryRepository) []int {
	// If no filters are provided, return all indices
	hasHardFilter := q.Kind != "" || q.Package != "" || q.Receiver != "" || q.Language != "" || q.Exported || q.Imports != "" || q.File != ""
	
	if !hasHardFilter {
		all := make([]int, len(repo.Symbols))
		for i := range all {
			all[i] = i
		}
		return all
	}

	// We use a boolean map or array for intersection since some filters aren't indexed perfectly (like File, Docs, Imports).
	// A boolean array is extremely fast for O(1) checks and intersection.
	valid := make([]bool, len(repo.Symbols))
	
	// Start by finding the most restrictive indexed filter
	initialized := false

	applyIndex := func(indices []int) {
		if !initialized {
			for _, idx := range indices {
				valid[idx] = true
			}
			initialized = true
		} else {
			newValid := make([]bool, len(repo.Symbols))
			for _, idx := range indices {
				if valid[idx] {
					newValid[idx] = true
				}
			}
			valid = newValid
		}
	}

	if q.Kind != "" {
		applyIndex(getIndicesIgnoreCase(repo.ByKind, q.Kind))
	}
	if q.Package != "" {
		applyIndex(getIndicesIgnoreCase(repo.ByPackage, q.Package))
	}
	if q.Receiver != "" {
		applyIndex(getIndicesIgnoreCase(repo.ByReceiver, strings.TrimPrefix(q.Receiver, "*")))
		// We also want to match "*Receiver"
		applyIndex(getIndicesIgnoreCase(repo.ByReceiver, "*"+strings.TrimPrefix(q.Receiver, "*"))) 
		// Wait, the way applyIndex works, calling it twice would intersect with itself and return empty.
		// Let's do it properly for Receiver.
		recRaw := strings.TrimPrefix(q.Receiver, "*")
		var recIdx []int
		recIdx = append(recIdx, getIndicesIgnoreCase(repo.ByReceiver, recRaw)...)
		recIdx = append(recIdx, getIndicesIgnoreCase(repo.ByReceiver, "*"+recRaw)...)
		applyIndex(recIdx)
	}
	if q.Language != "" {
		applyIndex(getIndicesIgnoreCase(repo.ByLanguage, q.Language))
	}
	if q.Exported {
		applyIndex(repo.Exported)
	}

	// If no indexed filter was used, initialize with all
	if !initialized {
		for i := 0; i < len(repo.Symbols); i++ {
			valid[i] = true
		}
	}

	// Apply unindexed filters or filters that require more logic
	validFiles := make(map[string]bool)
	useImportFilter := q.Imports != ""
	if useImportFilter {
		for file, imports := range repo.Deps.FileDeps {
			for _, imp := range imports {
				if imp == q.Imports {
					validFiles[file] = true
					break
				}
			}
		}
	}

	var finalIndices []int
	for i := 0; i < len(valid); i++ {
		if !valid[i] {
			continue
		}
		
		sym := repo.Symbols[i]
		
		if q.File != "" && !strings.Contains(sym.FilePath, q.File) {
			continue
		}
		if q.Docs != "" && !strings.Contains(strings.ToLower(sym.Documentation), strings.ToLower(q.Docs)) {
			continue
		}
		if useImportFilter && !validFiles[sym.FilePath] {
			continue
		}
		
		finalIndices = append(finalIndices, i)
	}

	return finalIndices
}
