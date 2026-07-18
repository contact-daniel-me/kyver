package analyzer

import (
	"os"
	"strings"
	"time"
)

type ArchitectureSummary struct {
	SchemaVersion        string    `json:"schemaVersion"`
	KyverVersion         string    `json:"kyverVersion"`
	GeneratedAt          time.Time `json:"generatedAt"`
	Packages             []string  `json:"packages"`
	InternalDependencies []string  `json:"internalDependencies"`
	ExternalDependencies []string  `json:"externalDependencies"`
	EntryPackages        []string  `json:"entryPackages"`
	PotentialCyclic      []string  `json:"potentialCyclicDependencies"`
}

func GenerateArchitectureSummary(kyverVersion string, deps *DependencyGraph) *ArchitectureSummary {
	summary := &ArchitectureSummary{
		SchemaVersion:        "1.0",
		KyverVersion:         kyverVersion,
		GeneratedAt:          time.Now(),
		Packages:             []string{},
		InternalDependencies: []string{},
		ExternalDependencies: []string{},
		EntryPackages:        []string{},
		PotentialCyclic:      []string{},
	}

	for pkg := range deps.PackageDeps {
		summary.Packages = append(summary.Packages, pkg)
		if pkg == "main" {
			summary.EntryPackages = append(summary.EntryPackages, pkg)
		}
	}

	// Try to determine module name from go.mod
	modulePrefix := ""
	if data, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				modulePrefix = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}
	}

	isInternal := func(pkg string) bool {
		if modulePrefix != "" && strings.HasPrefix(pkg, modulePrefix) {
			return true
		}
		_, ok := deps.PackageDeps[pkg]
		if !ok {
			for known := range deps.PackageDeps {
				if strings.HasPrefix(pkg, known+"/") {
					return true
				}
			}
		}
		return ok
	}

	extMap := make(map[string]bool)
	intMap := make(map[string]bool)

	for _, imports := range deps.PackageDeps {
		for _, imp := range imports {
			if isInternal(imp) {
				intMap[imp] = true
			} else {
				extMap[imp] = true
			}
		}
	}

	for ext := range extMap {
		summary.ExternalDependencies = append(summary.ExternalDependencies, ext)
	}
	for intD := range intMap {
		summary.InternalDependencies = append(summary.InternalDependencies, intD)
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var checkCycle func(node string) bool
	checkCycle = func(node string) bool {
		if !visited[node] {
			visited[node] = true
			recStack[node] = true

			for _, neighbor := range deps.PackageDeps[node] {
				if !visited[neighbor] && checkCycle(neighbor) {
					return true
				} else if recStack[neighbor] {
					summary.PotentialCyclic = append(summary.PotentialCyclic, node+" -> "+neighbor)
					return true
				}
			}
		}
		recStack[node] = false
		return false
	}

	for pkg := range deps.PackageDeps {
		if !visited[pkg] {
			checkCycle(pkg)
		}
	}

	return summary
}
