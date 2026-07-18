package graph

import (
	"math"
	"sort"
	"strings"
)

func sortImpactCallers(callers []ImpactCaller, sortBy string) {
	sort.Slice(callers, func(i, j int) bool {
		return impactCallerLess(callers[i], callers[j], sortBy)
	})
}

func impactCallerLess(a, b ImpactCaller, sortBy string) bool {
	if sortBy == "name" {
		pa, pb := strings.ToLower(a.Package), strings.ToLower(b.Package)
		if pa != pb {
			return pa < pb
		}
		na, nb := strings.ToLower(a.Name), strings.ToLower(b.Name)
		if na != nb {
			return na < nb
		}
		return strings.ToLower(a.Receiver) < strings.ToLower(b.Receiver)
	}
	pa, pb := strings.ToLower(a.Package), strings.ToLower(b.Package)
	if pa != pb {
		return pa < pb
	}
	if a.StartLine != b.StartLine {
		return a.StartLine < b.StartLine
	}
	return strings.ToLower(a.Name) < strings.ToLower(b.Name)
}

func dedupeImpactCallers(callers []ImpactCaller) []ImpactCaller {
	seen := make(map[string]bool)
	out := make([]ImpactCaller, 0, len(callers))
	for _, c := range callers {
		key := symbolKey(c.Package, c.Name, c.Receiver)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

func filterImpactCallersByPackage(callers []ImpactCaller, pkg string) []ImpactCaller {
	if pkg == "" {
		return callers
	}
	out := make([]ImpactCaller, 0, len(callers))
	for _, c := range callers {
		if strings.EqualFold(c.Package, pkg) {
			out = append(out, c)
		}
	}
	return out
}

func applyImpactMax(direct, indirect []ImpactCaller, max int) ([]ImpactCaller, []ImpactCaller) {
	if max <= 0 {
		return direct, indirect
	}
	if len(direct) >= max {
		return direct[:max], nil
	}
	remaining := max - len(direct)
	if len(indirect) > remaining {
		return direct, indirect[:remaining]
	}
	return direct, indirect
}

func collectImpactDerived(direct, indirect []ImpactCaller) (files, pkgs []string, tests []ImpactCaller) {
	fileSet := make(map[string]bool)
	pkgSet := make(map[string]bool)

	collect := func(c ImpactCaller) {
		if c.FilePath != "" {
			fileSet[c.FilePath] = true
		}
		if c.Package != "" {
			pkgSet[c.Package] = true
		}
		if isTestCaller(c) {
			tests = append(tests, c)
		}
	}

	for _, c := range direct {
		collect(c)
	}
	for _, c := range indirect {
		collect(c)
	}

	return sortedKeys(fileSet), sortedKeys(pkgSet), tests
}

func countPublicImpactCallers(direct, indirect []ImpactCaller) int {
	n := 0
	for _, c := range direct {
		if c.Exported {
			n++
		}
	}
	for _, c := range indirect {
		if c.Exported {
			n++
		}
	}
	return n
}

func roundLookupMs(ms float64) float64 {
	return math.Round(ms*100) / 100
}

func finalizeImpactResult(res *ImpactResult, opts ImpactOptions, startMs float64) {
	res.DirectCallers = dedupeImpactCallers(res.DirectCallers)
	res.IndirectCallers = dedupeImpactCallers(res.IndirectCallers)

	sortImpactCallers(res.DirectCallers, opts.SortBy)
	sortImpactCallers(res.IndirectCallers, opts.SortBy)

	if opts.PackageFilter != "" {
		res.DirectCallers = filterImpactCallersByPackage(res.DirectCallers, opts.PackageFilter)
		res.IndirectCallers = filterImpactCallersByPackage(res.IndirectCallers, opts.PackageFilter)
	}

	res.DirectCallers, res.IndirectCallers = applyImpactMax(res.DirectCallers, res.IndirectCallers, opts.Max)

	files, pkgs, tests := collectImpactDerived(res.DirectCallers, res.IndirectCallers)
	res.AffectedFiles = files
	res.AffectedPackages = pkgs
	sortImpactCallers(tests, opts.SortBy)
	res.Tests = tests

	publicAPIs := countPublicImpactCallers(res.DirectCallers, res.IndirectCallers)
	nDirect := len(res.DirectCallers)
	nIndirect := len(res.IndirectCallers)
	nPkgs := len(pkgs)

	severity, rationale := computeImpactSeverity(nDirect, nIndirect, nPkgs, publicAPIs)
	res.Severity = severity
	res.SeverityRationale = rationale

	res.Stats = ImpactStats{
		DirectCallers:   nDirect,
		IndirectCallers: nIndirect,
		AffectedFiles:   len(files),
		AffectedPkgs:    nPkgs,
		PublicAPIs:      publicAPIs,
	}
	res.LookupTimeMs = startMs // Do not round to zero; millisecond precision
}
