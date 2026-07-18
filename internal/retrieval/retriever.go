package retrieval

import (
	"strings"

	"github.com/contact-daniel-me/kyver/internal/context"
	"github.com/contact-daniel-me/kyver/internal/search"
)

func (e *Engine) retrieveFromQuestion(text string, bundle *RetrievalBundle) error {
	keywords := parseKeywords(text)
	
	if len(keywords) == 0 {
		bundle.Confidence = 0.0
		return nil
	}

	var bestScore float64 = 0.0
	var bestContexts []*context.ContextResult

	for _, kw := range keywords {
		// Try to search for symbols
		q := search.Query{Text: kw}
		res, err := e.searchEngine.Search(q)
		if err == nil && len(res) > 0 {
			// Found symbols matching the keyword
			// Build context for the top hit
			topSym := res[0]
			ctxRes, err := e.ctxEngine.BuildSymbolContext(topSym.Name, context.ContextOptions{})
			if err == nil {
				// Score it
				score := calculateSymbolScore(kw, topSym.Name)
				if score > bestScore {
					bestScore = score
				}
				bestContexts = append(bestContexts, ctxRes)
			}
		}

		// Check if keyword is a package or directory
		pkgCtx, err := e.ctxEngine.BuildPackageContext(strings.ToLower(kw), context.ContextOptions{})
		if err == nil && len(pkgCtx.Dependencies) > 0 {
			score := calculatePackageScore(kw, strings.ToLower(kw))
			if score > bestScore {
				bestScore = score
			}
			bestContexts = append(bestContexts, pkgCtx)
		}
	}

	bundle.Confidence = bestScore
	// Trim bestContexts to top 3
	if len(bestContexts) > 3 {
		bundle.TopContexts = bestContexts[:3]
	} else {
		bundle.TopContexts = bestContexts
	}

	return nil
}
