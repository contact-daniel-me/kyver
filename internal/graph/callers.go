package graph

import (
	"math"
	"sort"
	"strings"
	"time"
)

func (e *Engine) Callers(symbol string, opts CallersOptions) (*CallersResult, error) {
	start := time.Now()
	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &CallersResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	found := e.findCallersOf(selected, e.fileCache)

	if opts.PackageFilter != "" {
		var filtered []CallerNode
		for _, c := range found {
			if strings.EqualFold(c.Package, opts.PackageFilter) {
				filtered = append(filtered, c)
			}
		}
		found = filtered
	}

	sort.Slice(found, func(i, j int) bool {
		if opts.SortBy == "name" {
			return strings.ToLower(found[i].Name) < strings.ToLower(found[j].Name)
		}
		return found[i].StartLine < found[j].StartLine
	})

	if opts.Max > 0 && len(found) > opts.Max {
		found = found[:opts.Max]
	}

	grouped := make(map[string][]CallerNode)
	for _, sym := range found {
		grouped[sym.Package] = append(grouped[sym.Package], sym)
	}

	var groups []CallerGroup
	for pkg, syms := range grouped {
		groups = append(groups, CallerGroup{
			Package: pkg,
			Callers: syms,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Package) < strings.ToLower(groups[j].Package)
	})

	res.Groups = groups

	filesInvolved := make(map[string]bool)
	for _, sym := range found {
		filesInvolved[sym.FilePath] = true
	}

	res.Stats = CallersStats{
		TotalCallers: len(found),
		TotalPkgs:    len(groups),
		TotalFiles:   len(filesInvolved),
		LookupTimeMs: math.Round((time.Since(start).Seconds()*1000)*100) / 100,
	}

	return res, nil
}
