package graph

import (
	"sort"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)


func NewEngineWithRepo(repo *search.InMemoryRepository) *Engine {
	e := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
		fileCache:    make(map[string][]string),
	}
	e.buildPackageGraph()
	return e
}

func (e *Engine) Graph(symbol string, opts GraphOptions) (*SymbolGraphResult, error) {
	start := time.Now()
	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &SymbolGraphResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = len(candidates) > 1
		return res, nil
	}

	res.Symbol = selected

	// Methods — filter to only include methods whose package matches the selected symbol
	var methods []analyzer.Symbol
	for _, idx := range e.repo.ByReceiver[selected.Name] {
		m := e.repo.Symbols[idx]
		if m.Package == selected.Package {
			methods = append(methods, m)
		}
	}
	for _, idx := range e.repo.ByReceiver["*" + selected.Name] {
		m := e.repo.Symbols[idx]
		if m.Package == selected.Package {
			methods = append(methods, m)
		}
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Name < methods[j].Name })
	res.Methods = methods

	// Callers
	callersRes, _ := e.Callers(symbol, CallersOptions{SelectIndex: opts.SelectIndex})
	if callersRes != nil {
		for _, group := range callersRes.Groups {
			for _, caller := range group.Callers {
				res.Callers = append(res.Callers, caller.Symbol)
			}
		}
	}

	// Callees
	calleesRes, _ := e.Callees(symbol, CalleesOptions{SelectIndex: opts.SelectIndex})
	if calleesRes != nil {
		for _, group := range calleesRes.Groups {
			for _, callee := range group.Callees {
				res.Callees = append(res.Callees, callee.Symbol)
			}
		}
	}

	// Dependencies
	var deps []string
	if selected.FilePath != "" {
		fileDeps := e.repo.Deps.FileDeps[selected.FilePath]
		deps = append(deps, fileDeps...)
		sort.Strings(deps)
	}
	res.Dependencies = deps

	// References
	refsRes, _ := e.References(symbol, ReferenceOptions{SelectIndex: opts.SelectIndex})
	if refsRes != nil {
		for _, g := range refsRes.Groups {
			for _, ref := range g.References {
				res.References = append(res.References, ref)
			}
		}
	}

	// Stats
	res.Stats = GraphStats{
		Methods:      len(res.Methods),
		Callers:      len(res.Callers),
		Callees:      len(res.Callees),
		Dependencies: len(res.Dependencies),
		References:   len(res.References),
		LookupTimeMs: time.Since(start).Seconds() * 1000,
	}

	return res, nil
}
