package graph

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func (e *Engine) Callees(symbol string, opts CalleesOptions) (*CalleesResult, error) {
	start := time.Now()
	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &CalleesResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	var found []CalleeNode
	seen := make(map[string]bool)

	// Read the file of the selected symbol (target)
	var body string
	
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

	lines := getFileLines(selected.FilePath)
	if len(lines) > 0 {
		startLine := selected.StartLine - 1
		if startLine < 0 {
			startLine = 0
		}
		endLine := selected.EndLine
		if endLine > len(lines) {
			endLine = len(lines)
		}
		if startLine < endLine {
			body = strings.Join(lines[startLine:endLine], "\n")
			
			// Skip the signature
			braceIdx := strings.Index(body, "{")
			if braceIdx != -1 {
				body = body[braceIdx:]
			}

			// Strip out single-line comments
			var cleanBody strings.Builder
			inLineComment := false
			for i := 0; i < len(body); i++ {
				if inLineComment {
					if body[i] == '\n' {
						inLineComment = false
						cleanBody.WriteByte(body[i])
					}
					continue
				}
				if i+1 < len(body) && body[i] == '/' && body[i+1] == '/' {
					inLineComment = true
					i++
					continue
				}
				cleanBody.WriteByte(body[i])
			}
			body = cleanBody.String()
		}
	}

	if opts.PackageFilter != "" {
		canonical, err := e.validatePackageFilter(opts.PackageFilter)
		if err != nil {
			return nil, err
		}
		opts.PackageFilter = canonical
	}

	if body != "" {
		// Identify potential callees from imported packages + same package
		deps := e.repo.Deps.FileDeps[selected.FilePath]
		targetPkg := selected.Package
		
		var potentialCallees []analyzer.Symbol
		
		for _, dep := range deps {
			parts := strings.Split(dep, "/")
			depPkg := parts[len(parts)-1]
			
			for _, idx := range e.repo.ByPackage[depPkg] {
				sym := e.repo.Symbols[idx]
				if sym.Exported && (sym.Kind == "Function" || sym.Kind == "Method") {
					potentialCallees = append(potentialCallees, sym)
				}
			}
		}
		
		// Add functions in the same package
		for _, idx := range e.repo.ByPackage[targetPkg] {
			sym := e.repo.Symbols[idx]
			if sym.Kind == "Function" || sym.Kind == "Method" {
				// skip self unless recursive (handled later maybe, or skip exact self)
				if sym.Name == selected.Name && sym.Receiver == selected.Receiver {
					continue
				}
				potentialCallees = append(potentialCallees, sym)
			}
		}

		// Perform lexical matching
		for _, sym := range potentialCallees {
			if !seen[sym.Name] {
				if opts.PackageFilter == "" || strings.EqualFold(sym.Package, opts.PackageFilter) {
					targetToken := sym.Name
					idx := strings.Index(body, targetToken)
					hasInvocation := false

					for idx != -1 {
						preOK, postOK := true, true
						
						if idx > 0 {
							preChar := body[idx-1]
							if (preChar >= 'a' && preChar <= 'z') || (preChar >= 'A' && preChar <= 'Z') || (preChar >= '0' && preChar <= '9') || preChar == '_' {
								preOK = false
							}
						}
						
						endIdx := idx + len(targetToken)
						if endIdx < len(body) {
							postChar := body[endIdx]
							if (postChar >= 'a' && postChar <= 'z') || (postChar >= 'A' && postChar <= 'Z') || (postChar >= '0' && postChar <= '9') || postChar == '_' {
								postOK = false
							}
						}
						
						if preOK && postOK {
							hasInvocation = true
							break
						}
						
						nextIdx := strings.Index(body[idx+1:], targetToken)
						if nextIdx == -1 {
							break
						}
						idx += 1 + nextIdx
					}

					if hasInvocation {
						found = append(found, CalleeNode{
							Symbol: sym,
						})
						seen[sym.Name] = true
					}
				}
			}
		}
	}

	// Sort globally first
	sort.Slice(found, func(i, j int) bool {
		if opts.SortBy == "name" {
			return strings.ToLower(found[i].Name) < strings.ToLower(found[j].Name)
		}
		return found[i].StartLine < found[j].StartLine
	})

	// Truncate by max
	if opts.Max > 0 && len(found) > opts.Max {
		found = found[:opts.Max]
	}

	// Group by package
	grouped := make(map[string][]CalleeNode)
	for _, sym := range found {
		grouped[sym.Package] = append(grouped[sym.Package], sym)
	}

	var groups []CalleeGroup
	for pkg, syms := range grouped {
		groups = append(groups, CalleeGroup{
			Package: pkg,
			Callees: syms,
		})
	}

	// Sort groups by package name
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Package) < strings.ToLower(groups[j].Package)
	})

	res.Groups = groups

	// Stats
	filesInvolved := make(map[string]bool)
	for _, sym := range found {
		filesInvolved[sym.FilePath] = true
	}
	
	res.Stats = CalleesStats{
		TotalCallees: len(found),
		TotalPkgs:    len(groups),
		TotalFiles:   len(filesInvolved),
		LookupTimeMs: time.Since(start).Seconds() * 1000,
	}

	return res, nil
}
