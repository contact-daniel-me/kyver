package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func (e *Engine) References(symbol string, opts ReferenceOptions) (*ReferenceResult, error) {
	start := time.Now()

	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &ReferenceResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	targetPkg := selected.Package

	if e.fileCache == nil {
		e.fileCache = make(map[string][]string)
	}
	getFileLines := func(filePath string) []string {
		if lines, ok := e.fileCache[filePath]; ok {
			return lines
		}
		var lines []string
		if content, err := os.ReadFile(filePath); err == nil {
			lines = strings.Split(string(content), "\n")
		}
		e.fileCache[filePath] = lines
		return lines
	}

	var allRefs []Reference
	seenLine := make(map[string]bool)

	// We look for files that import the target package, or are in the target package itself.
	for file, deps := range e.repo.Deps.FileDeps {
		importsTarget := false
		for _, dep := range deps {
			if strings.HasSuffix(dep, targetPkg) {
				importsTarget = true
				break
			}
		}

		var filePkg string
		for _, sym := range e.repo.Symbols {
			if filepath.ToSlash(sym.FilePath) == filepath.ToSlash(file) {
				filePkg = sym.Package
				break
			}
		}

		if filePkg == targetPkg {
			importsTarget = true
		}

		if !importsTarget {
			continue
		}

		// Apply file filter
		if opts.FileFilter != "" {
			if !strings.Contains(filepath.ToSlash(file), opts.FileFilter) && !strings.Contains(filepath.Base(file), opts.FileFilter) {
				continue
			}
		}

		// Apply package filter
		if opts.PackageFilter != "" && filePkg != opts.PackageFilter {
			continue
		}

		lines := getFileLines(file)
		for i, line := range lines {
			if !containsToken(line, selected.Name) {
				continue
			}
			
			// Quick filtering of import statements for exact package match if it's the package name itself
			// However, since we search for the specific symbol, an import line is usually not a reference to the symbol unless the symbol is the package name.
			// Let's strip comments to be safe
			cleanLine := stripLineComments(line)
			if !containsToken(cleanLine, selected.Name) {
				continue
			}

			// Skip the symbol's own definition line
			if filepath.ToSlash(file) == filepath.ToSlash(selected.FilePath) {
				if i+1 == selected.StartLine {
					continue
				}
				// Robust check for stale caches
				if (selected.Kind == "Struct" || selected.Kind == "Interface") && strings.HasPrefix(strings.TrimSpace(cleanLine), "type "+selected.Name) {
					continue
				}
				if selected.Kind == "Function" && strings.HasPrefix(strings.TrimSpace(cleanLine), "func "+selected.Name) {
					continue
				}
				if selected.Kind == "Method" && strings.HasPrefix(strings.TrimSpace(cleanLine), "func ") && strings.Contains(cleanLine, selected.Name+"(") {
					continue
				}
			}

			// We found a reference. What is the symbol context of this reference?
			// We can find the enclosing function/struct to set as "Name" and "Kind".
			var enclosing analyzer.Symbol
			for _, sym := range e.repo.Symbols {
				if filepath.ToSlash(sym.FilePath) == filepath.ToSlash(file) {
					if i+1 >= sym.StartLine && i+1 <= sym.EndLine {
						enclosing = sym
						// We prefer the most specific symbol if nested, but in Go top-level is fine.
					}
				}
			}

			key := fmt.Sprintf("%s:%d", file, i+1)
			if seenLine[key] {
				continue
			}
			seenLine[key] = true
			
			rel := "reference"
			if enclosing.Name != "" {
				rel = enclosing.Name + "()"
			} else {
				enclosing.Name = filepath.Base(file)
				enclosing.Kind = "File"
				enclosing.Package = filePkg
			}

			ref := Reference{
				Name:         enclosing.Name,
				Kind:         enclosing.Kind,
				Package:      filePkg,
				File:         file,
				Line:         i + 1,
				Context:      strings.TrimSpace(line),
				Receiver:     enclosing.Receiver,
				Relationship: rel,
			}
			allRefs = append(allRefs, ref)
		}
	}

	sortReferences(allRefs, opts.SortBy)
	
	if opts.Max > 0 && len(allRefs) > opts.Max {
		allRefs = allRefs[:opts.Max]
	}

	// Group by package
	pkgGroups := make(map[string][]Reference)
	for _, ref := range allRefs {
		pkgGroups[ref.Package] = append(pkgGroups[ref.Package], ref)
	}

	var groups []ReferenceGroup
	for pkg, refs := range pkgGroups {
		groups = append(groups, ReferenceGroup{
			Package:    pkg,
			References: refs,
		})
	}

	// Sort groups alphabetically
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Package < groups[j].Package
	})

	res.Groups = groups

	uniqueFiles := make(map[string]bool)
	uniquePkgs := make(map[string]bool)
	for _, ref := range allRefs {
		uniqueFiles[ref.File] = true
		uniquePkgs[ref.Package] = true
	}

	res.Stats = ReferenceStats{
		TotalReferences: len(allRefs),
		TotalFiles:      len(uniqueFiles),
		TotalPkgs:       len(uniquePkgs),
	}
	res.LookupTimeMs = time.Since(start).Seconds() * 1000

	return res, nil
}

func sortReferences(refs []Reference, sortBy string) {
	sort.Slice(refs, func(i, j int) bool {
		if sortBy == "name" {
			if refs[i].Name != refs[j].Name {
				return refs[i].Name < refs[j].Name
			}
		}
		// Default: file then line
		if refs[i].File != refs[j].File {
			return refs[i].File < refs[j].File
		}
		return refs[i].Line < refs[j].Line
	})
}
