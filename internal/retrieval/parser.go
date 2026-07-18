package retrieval

import (
	"regexp"
	"strings"
)

var stopWords = map[string]bool{
	"how": true, "does": true, "work": true, "what": true, "is": true,
	"the": true, "a": true, "an": true, "in": true, "of": true,
	"and": true, "or": true, "to": true, "for": true, "with": true,
	"on": true, "at": true, "by": true, "from": true, "about": true,
	"can": true, "you": true, "explain": true, "show": true, "me": true,
	"where": true, "why": true, "who": true, "which": true,
}

var wordRegex = regexp.MustCompile(`[a-zA-Z0-9_]+(\.[a-zA-Z0-9]+)?`)

// parseKeywords extracts relevant search keywords from a natural language query
func parseKeywords(text string) []string {
	var keywords []string
	words := wordRegex.FindAllString(text, -1)
	
	for _, word := range words {
		lower := strings.ToLower(word)
		if !stopWords[lower] && len(lower) > 2 {
			// Basic heuristic: preserve original casing for search, but check stopword against lowercase
			keywords = append(keywords, word)
		}
	}

	return keywords
}
