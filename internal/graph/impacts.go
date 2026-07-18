package graph

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

// Impacts performs semantic impact analysis by walking the reverse call graph.
// Only symbols reachable through actual call edges are reported — never every
// function in an importing package.
//
// Severity scoring algorithm (documented + tested):
//
//	Inputs: direct callers (D), transitive/indirect callers (T),
//	        affected packages (P), public/exported APIs among callers (A).
//
//	Low:      D == 0
//	Medium:   D >= 1 AND D < 3 AND P <= 1 AND T < 5
//	High:     (D >= 3 OR P >= 2 OR T >= 5) AND not Critical
//	Critical: D >= 5 AND P >= 3 AND A >= 1
//	          OR T >= 10 AND P >= 3 AND A >= 1
//
// Performance: reuses the indexed symbol/dependency graph and the shared
// lexical caller scanner (findCallersOf). File contents are cached and each
// unique symbol’s callers are computed at most once — O(N) over the impact cone.
func (e *Engine) Impacts(symbol string, opts ImpactOptions) (*ImpactResult, error) {
	start := time.Now()
	if opts.SortBy == "" {
		opts.SortBy = "line"
	}

	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &ImpactResult{
		Candidates:       candidates,
		DirectCallers:    []ImpactCaller{},
		IndirectCallers:  []ImpactCaller{},
		AffectedFiles:    []string{},
		AffectedPackages: []string{},
		Tests:            []ImpactCaller{},
	}

	if selected == nil {
		res.MultipleFound = len(candidates) > 1
		return res, nil
	}
	res.Symbol = selected

	if opts.PackageFilter != "" {
		canonical, err := e.validatePackageFilter(opts.PackageFilter)
		if err != nil {
			return nil, err
		}
		opts.PackageFilter = canonical
	}

	memo := make(map[string][]CallerNode)

	callersCached := func(sym *analyzer.Symbol) []CallerNode {
		key := symbolKey(sym.Package, sym.Name, sym.Receiver)
		if cached, ok := memo[key]; ok {
			return cached
		}
		found := e.findCallersOf(sym, e.fileCache)
		memo[key] = found
		return found
	}

	targetLabel := formatSymbolLabel(selected)
	targetKey := symbolKey(selected.Package, selected.Name, selected.Receiver)

	var directCallers []ImpactCaller
	visited := map[string]bool{targetKey: true}

	for _, caller := range callersCached(selected) {
		if caller.IsRecursive {
			continue
		}
		key := symbolKey(caller.Package, caller.Name, caller.Receiver)
		visited[key] = true
		directCallers = append(directCallers, toImpactCaller(caller.Symbol, "directly calls "+targetLabel, targetLabel))
	}
	res.DirectCallers = directCallers

	type bfsEntry struct {
		sym      *analyzer.Symbol
		viaLabel string
	}

	var queue []bfsEntry
	for _, dc := range directCallers {
		sym := e.lookupExactSymbol(dc.Package, dc.Name, dc.Receiver)
		if sym == nil {
			continue
		}
		queue = append(queue, bfsEntry{
			sym:      sym,
			viaLabel: formatCallerLabel(dc),
		})
	}

	var indirectCallers []ImpactCaller
	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		for _, caller := range callersCached(entry.sym) {
			if caller.IsRecursive {
				continue
			}
			key := symbolKey(caller.Package, caller.Name, caller.Receiver)
			if visited[key] {
				continue
			}
			visited[key] = true

			why := "calls " + entry.viaLabel
			ic := toImpactCaller(caller.Symbol, why, entry.viaLabel)
			indirectCallers = append(indirectCallers, ic)

			upstream := e.lookupExactSymbol(caller.Package, caller.Name, caller.Receiver)
			if upstream == nil {
				continue
			}
			queue = append(queue, bfsEntry{
				sym:      upstream,
				viaLabel: formatCallerLabel(ic),
			})
		}
	}
	res.IndirectCallers = indirectCallers

	elapsedMs := time.Since(start).Seconds() * 1000
	finalizeImpactResult(res, opts, elapsedMs)

	return res, nil
}

// computeImpactSeverity scores blast radius from call-graph metrics.
// See Impacts doc comment for the full algorithm.
func computeImpactSeverity(direct, indirect, pkgs, publicAPIs int) (string, string) {
	var reason []string
	
	if direct == 1 {
		reason = append(reason, "• 1 direct caller")
	} else {
		reason = append(reason, fmt.Sprintf("• %d direct callers", direct))
	}
	
	if pkgs == 1 {
		reason = append(reason, "• 1 affected package")
	} else {
		reason = append(reason, fmt.Sprintf("• %d affected packages", pkgs))
	}
	
	if indirect == 1 {
		reason = append(reason, "• 1 indirect caller")
	} else {
		reason = append(reason, fmt.Sprintf("• %d indirect callers", indirect))
	}
	
	if publicAPIs > 0 {
		reason = append(reason, "• Touches exported API")
	} else {
		reason = append(reason, "• Does not affect exported public APIs")
	}
	
	rationale := strings.Join(reason, "\n")

	if direct == 0 {
		return "Low", rationale
	}

	if (direct >= 5 && pkgs >= 3 && publicAPIs >= 1) ||
		(indirect >= 10 && pkgs >= 3 && publicAPIs >= 1) {
		return "Critical", rationale
	}

	if direct >= 3 || pkgs >= 2 || indirect >= 5 {
		return "High", rationale
	}

	return "Medium", rationale
}

func toImpactCaller(sym analyzer.Symbol, why, calls string) ImpactCaller {
	return ImpactCaller{
		Symbol: sym,
		Calls:  calls,
		Chain:  []string{why},
	}
}

func formatCallerLabel(c ImpactCaller) string {
	if c.Kind == "Method" && c.Receiver != "" {
		if strings.HasPrefix(c.Receiver, "*") {
			return fmt.Sprintf("%s.(%s).%s()", c.Package, c.Receiver, c.Name)
		}
		return fmt.Sprintf("%s.%s.%s()", c.Package, c.Receiver, c.Name)
	}
	if c.Package != "" {
		return fmt.Sprintf("%s.%s()", c.Package, c.Name)
	}
	return c.Name + "()"
}

func formatSymbolLabel(sym *analyzer.Symbol) string {
	if sym.Kind == "Method" {
		if strings.HasPrefix(sym.Receiver, "*") {
			return fmt.Sprintf("%s.(%s).%s()", sym.Package, sym.Receiver, sym.Name)
		}
		return fmt.Sprintf("%s.%s.%s()", sym.Package, sym.Receiver, sym.Name)
	}
	return fmt.Sprintf("%s.%s()", sym.Package, sym.Name)
}

func isTestCaller(c ImpactCaller) bool {
	if strings.HasSuffix(c.FilePath, "_test.go") {
		return true
	}
	return strings.HasPrefix(c.Name, "Test")
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// callerKey is retained for tests/helpers that key by package+name only.
func callerKey(pkg, name string) string {
	return pkg + "." + name
}
