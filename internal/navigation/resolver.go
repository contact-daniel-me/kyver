package navigation

import (
	"fmt"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

// resolveSymbol finds exact case-insensitive matches for a symbol.
// If multiple are found, it checks selectIndex. If selectIndex <= 0, it indicates multiple choices exist.
func (e *Engine) resolveSymbol(symbol string, selectIndex int) ([]analyzer.Symbol, *analyzer.Symbol, error) {
	q := strings.ToLower(symbol)
	var candidates []analyzer.Symbol

	for name, indices := range e.repo.ByName {
		if strings.ToLower(name) == q {
			for _, idx := range indices {
				candidates = append(candidates, e.repo.Symbols[idx])
			}
		}
	}

	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("symbol '%s' not found", symbol)
	}

	if len(candidates) == 1 {
		return candidates, &candidates[0], nil
	}

	// Multiple candidates exist
	if selectIndex > 0 && selectIndex <= len(candidates) {
		return candidates, &candidates[selectIndex-1], nil
	}

	// Needs selection
	return candidates, nil, nil
}
