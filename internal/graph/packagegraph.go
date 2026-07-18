package graph

import (
	"path/filepath"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func (e *Engine) PackageGraph(pkg string) (*PackageGraphResult, error) {
	var imports []string
	var importedBy []string

	for imp := range e.pkgImports[pkg] {
		imports = append(imports, imp)
	}
	for impBy := range e.pkgImportsBy[pkg] {
		importedBy = append(importedBy, impBy)
	}

	return &PackageGraphResult{
		Package:  pkg,
		Imports:  imports,
		Imported: importedBy,
	}, nil
}

func (e *Engine) FileGraph(file string) (*FileGraphResult, error) {
	var symbols []analyzer.Symbol
	var imports []string
	var incoming []string
	
	targetPath := filepath.ToSlash(file)
	var pkg string

	for _, sym := range e.repo.Symbols {
		if filepath.ToSlash(sym.FilePath) == targetPath {
			symbols = append(symbols, sym)
			pkg = sym.Package
		}
	}

	imports = e.repo.Deps.FileDeps[targetPath]

	if pkg != "" {
		for impBy := range e.pkgImportsBy[pkg] {
			incoming = append(incoming, impBy)
		}
	}

	return &FileGraphResult{
		File:     file,
		Imports:  imports,
		Symbols:  symbols,
		Incoming: incoming,
	}, nil
}
