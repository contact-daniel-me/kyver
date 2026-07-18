package navigation

import (
	"strings"
)

func (e *Engine) Hierarchy(symbol string, selectIndex int) (*HierarchyResult, error) {
	candidates, selected, err := e.resolveSymbol(symbol, selectIndex)
	if err != nil {
		return nil, err
	}

	res := &HierarchyResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	// Find Methods (where Receiver == selected.Name or *selected.Name)
	for _, idx := range e.repo.ByReceiver[selected.Name] {
		sym := e.repo.Symbols[idx]
		if sym.Package == selected.Package {
			res.Methods = append(res.Methods, sym)
		}
	}
	for _, idx := range e.repo.ByReceiver["*"+selected.Name] {
		sym := e.repo.Symbols[idx]
		if sym.Package == selected.Package {
			res.Methods = append(res.Methods, sym)
		}
	}

	// Find Related Symbols (interfaces/structs in same package with substring)
	target := strings.ToLower(selected.Name)
	seen := make(map[string]bool)
	seen[selected.Name] = true

	for name, indices := range e.repo.ByName {
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, target) || strings.Contains(target, lowerName) {
			if !seen[name] {
				for _, idx := range indices {
					sym := e.repo.Symbols[idx]
					if sym.Package == selected.Package && (sym.Kind == "Struct" || sym.Kind == "Interface") {
						res.Related = append(res.Related, sym)
						seen[name] = true
						break
					}
				}
			}
		}
	}

	// Find Dependencies
	res.Dependencies = e.repo.Deps.FileDeps[selected.FilePath]

	return res, nil
}
