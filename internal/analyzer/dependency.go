package analyzer

import "time"

// DependencyGraph represents the relations between packages and files
type DependencyGraph struct {
	SchemaVersion string              `json:"schemaVersion"`
	KyverVersion  string              `json:"kyverVersion"`
	GeneratedAt   time.Time           `json:"generatedAt"`
	PackageDeps   map[string][]string `json:"packageDependencies"`
	FileDeps      map[string][]string `json:"fileDependencies"`
}

// NewDependencyGraph creates an empty dependency graph
func NewDependencyGraph(kyverVersion string) *DependencyGraph {
	return &DependencyGraph{
		SchemaVersion: "1.0",
		KyverVersion:  kyverVersion,
		GeneratedAt:   time.Now(),
		PackageDeps:   make(map[string][]string),
		FileDeps:      make(map[string][]string),
	}
}

// AddFileDependency adds imported packages for a specific file
func (d *DependencyGraph) AddFileDependency(filePath string, imports []string) {
	d.FileDeps[filePath] = imports
}

// AddPackageDependency aggregates dependencies at the package level
func (d *DependencyGraph) AddPackageDependency(pkgName string, imports []string) {
	existing := make(map[string]bool)
	for _, imp := range d.PackageDeps[pkgName] {
		existing[imp] = true
	}

	for _, imp := range imports {
		if !existing[imp] {
			d.PackageDeps[pkgName] = append(d.PackageDeps[pkgName], imp)
			existing[imp] = true
		}
	}
}
