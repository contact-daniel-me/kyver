package graph

import (
	"fmt"
	"sort"
	"strings"
)

// validatePackageFilter checks that filter matches a known package (case-insensitive).
// On success it returns the canonical package name from the index.
func (e *Engine) validatePackageFilter(filter string) (string, error) {
	if filter == "" {
		return "", nil
	}

	var available []string
	found := false
	canonical := filter
	for pkg := range e.repo.ByPackage {
		available = append(available, pkg)
		if strings.EqualFold(pkg, filter) {
			found = true
			canonical = pkg
		}
	}
	if !found {
		sort.Strings(available)
		return "", fmt.Errorf("Unknown package %q\n\n%s", filter, strings.Join(available, "\n"))
	}
	return canonical, nil
}

func (e *Engine) listPackages() []string {
	pkgs := make([]string, 0, len(e.repo.ByPackage))
	for pkg := range e.repo.ByPackage {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}
