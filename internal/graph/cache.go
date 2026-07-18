package graph

import (
	"fmt"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)

type Engine struct {
	repo         *search.InMemoryRepository
	pkgImports   map[string]map[string]bool
	pkgImportsBy map[string]map[string]bool
	fileCache    map[string][]string
}

func NewEngine(rootDir string) (*Engine, error) {
	repo, err := search.LoadRepository(rootDir)
	if err != nil {
		return nil, err
	}

	e := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
		fileCache:    make(map[string][]string),
	}
	e.buildPackageGraph()

	return e, nil
}

func (e *Engine) buildPackageGraph() {
	// Build mapping from filepath to package name
	fileToPkg := make(map[string]string)
	for _, sym := range e.repo.Symbols {
		fileToPkg[sym.FilePath] = sym.Package
	}

	for file, deps := range e.repo.Deps.FileDeps {
		pkg := fileToPkg[file]
		if pkg == "" {
			continue // Not a recognized file with symbols
		}

		if e.pkgImports[pkg] == nil {
			e.pkgImports[pkg] = make(map[string]bool)
		}

		for _, dep := range deps {
			// Extract last part of import path as a naive package name
			parts := strings.Split(dep, "/")
			depPkg := parts[len(parts)-1]
			
			e.pkgImports[pkg][depPkg] = true

			if e.pkgImportsBy[depPkg] == nil {
				e.pkgImportsBy[depPkg] = make(map[string]bool)
			}
			e.pkgImportsBy[depPkg][pkg] = true
		}
	}
}

// resolveSymbol finds exact case-insensitive matches for a symbol.
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

	if selectIndex > 0 {
		if selectIndex <= len(candidates) {
			return candidates, &candidates[selectIndex-1], nil
		}
		return nil, nil, fmt.Errorf("invalid selection: %d (max %d)", selectIndex, len(candidates))
	}

	return candidates, nil, nil
}
