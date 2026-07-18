package navigation

import (
	"bufio"
	"os"
	"strings"
)

func (e *Engine) Peek(symbol string, selectIndex int, linesLimit int) (*PeekResult, error) {
	candidates, selected, err := e.resolveSymbol(symbol, selectIndex)
	if err != nil {
		return nil, err
	}

	res := &PeekResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	// Extract context
	// 1. Methods
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

	// 2. Implements (Using Related logic for simplicity, focusing on Interfaces/Structs in same package)
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
						res.Implements = append(res.Implements, sym)
						seen[name] = true
						break
					}
				}
			}
		}
	}

	// 3. Dependencies
	if deps, ok := e.repo.Deps.FileDeps[selected.FilePath]; ok {
		res.Dependencies = deps
	}

	// 4. ReferencedBy
	var references []string
	targetPkg := selected.Package
	for file, imports := range e.repo.Deps.FileDeps {
		for _, imp := range imports {
			if imp == targetPkg || strings.HasSuffix(imp, "/"+targetPkg) {
				references = append(references, file)
				break
			}
		}
	}
	res.ReferencedBy = references

	// Extract Snippet
	f, err := os.Open(selected.FilePath)
	if err == nil {
		defer f.Close()

		targetStart := selected.StartLine
		targetEnd := targetStart + linesLimit - 1

		scanner := bufio.NewScanner(f)
		lineNum := 1
		var lines []string
		
		for scanner.Scan() {
			if lineNum >= targetStart && lineNum <= targetEnd {
				lines = append(lines, scanner.Text())
			}
			if lineNum > targetEnd {
				break
			}
			lineNum++
		}
		
		res.Snippet = strings.Join(lines, "\n")
		
		contextEnd := selected.EndLine
		for _, m := range res.Methods {
			if m.FilePath == selected.FilePath && m.EndLine > contextEnd {
				contextEnd = m.EndLine
			}
		}

		if contextEnd > targetEnd {
			res.Truncated = true
			res.TruncatedLines = contextEnd - targetEnd
		}
	} else {
		return nil, err
	}

	return res, nil
}
