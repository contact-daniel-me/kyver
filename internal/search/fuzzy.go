package search

import "strings"

// LevenshteinDistance calculates the minimum number of single-character edits
// (insertions, deletions, or substitutions) required to change s into t.
func LevenshteinDistance(s, t string) int {
	s, t = strings.ToLower(s), strings.ToLower(t)
	if len(s) == 0 {
		return len(t)
	}
	if len(t) == 0 {
		return len(s)
	}

	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j] + 1
				if d[i][j-1]+1 < min {
					min = d[i][j-1] + 1
				}
				if d[i-1][j-1]+1 < min {
					min = d[i-1][j-1] + 1
				}
				d[i][j] = min
			}
		}
	}
	return d[len(s)][len(t)]
}

// IsSubsequence checks if query characters appear in target in the same order
// e.g. "IdxRepo" in "IndexRepository"
func IsSubsequence(query, target string) bool {
	q := strings.ToLower(query)
	t := strings.ToLower(target)
	
	i, j := 0, 0
	for i < len(q) && j < len(t) {
		if q[i] == t[j] {
			i++
		}
		j++
	}
	return i == len(q)
}

// IsCamelCaseAbbr checks if the query matches the uppercase letters of a camelCase/PascalCase string
// e.g. "RR" matches "RepositoryRoot"
func IsCamelCaseAbbr(query, target string) bool {
	if query == "" {
		return false
	}
	var caps string
	for _, ch := range target {
		if ch >= 'A' && ch <= 'Z' {
			caps += string(ch)
		}
	}
	if len(caps) == 0 {
		return false
	}
	return strings.ToLower(query) == strings.ToLower(caps)
}

// IsWholeWordMatch checks if the query is a distinct word boundary match inside the target
func IsWholeWordMatch(query, target string) bool {
	if query == "" {
		return false
	}
	q := strings.ToLower(query)
	t := strings.ToLower(target)
	
	idx := strings.Index(t, q)
	if idx == -1 {
		return false
	}
	
	// Check left boundary
	if idx > 0 {
		left := t[idx-1]
		if (left >= 'a' && left <= 'z') || (left >= '0' && left <= '9') {
			return false
		}
	}
	
	// Check right boundary
	rightIdx := idx + len(q)
	if rightIdx < len(t) {
		right := t[rightIdx]
		if (right >= 'a' && right <= 'z') || (right >= '0' && right <= '9') {
			return false
		}
	}
	
	return true
}

// IsFuzzyMatch returns true if the query is a fuzzy match for target
func IsFuzzyMatch(query, target string) bool {
	if query == "" {
		return false
	}
	
	if len(query) >= 4 && IsSubsequence(query, target) {
		return true
	}
	
	// Check edit distance for small typos (e.g. Authentcation -> Authentication)
	// Allow 1 typo per 4 characters of query length, max 3
	allowedTypos := len(query) / 4
	if allowedTypos < 1 {
		allowedTypos = 1
	}
	if allowedTypos > 3 {
		allowedTypos = 3
	}
	
	dist := LevenshteinDistance(query, target)
	return dist <= allowedTypos
}
