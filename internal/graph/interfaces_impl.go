package graph

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

type interfaceMethod struct {
	Name      string
	Signature string
}

func (e *Engine) Implements(symbol string, opts ImplementsOptions) (*ImplementsResult, error) {
	start := time.Now()

	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &ImplementsResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	if selected.Kind != "Interface" {
		return nil, fmt.Errorf("symbol '%s' is a %s, not an Interface", selected.Name, selected.Kind)
	}

	if e.fileCache == nil {
		e.fileCache = make(map[string][]string)
	}
	getFileLines := func(filePath string) []string {
		if lines, ok := e.fileCache[filePath]; ok {
			return lines
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(data), "\n")
		e.fileCache[filePath] = lines
		return lines
	}

	requiredInterfaceMethods := e.extractInterfaceMethods(selected, getFileLines)
	if len(requiredInterfaceMethods) == 0 {
		res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)
		return res, nil
	}

	var requiredSignatures []string
	var requiredNames []string
	for _, m := range requiredInterfaceMethods {
		requiredSignatures = append(requiredSignatures, m.Signature)
		requiredNames = append(requiredNames, m.Name)
	}
	res.RequiredMethods = requiredSignatures

	var allImpls []Implementation
	
	for i := range e.repo.Symbols {
		sym := e.repo.Symbols[i]
		if sym.Kind != "Struct" {
			continue
		}

		if opts.PackageFilter != "" && sym.Package != opts.PackageFilter {
			continue
		}

		var valueMethods []string
		for _, idx := range e.repo.ByReceiver[sym.Name] {
			valueMethods = append(valueMethods, e.repo.Symbols[idx].Name)
		}
		var pointerMethods []string
		for _, idx := range e.repo.ByReceiver["*"+sym.Name] {
			pointerMethods = append(pointerMethods, e.repo.Symbols[idx].Name)
		}

		implementsAsValue := hasAllMethods(valueMethods, requiredNames)
		
		allMethods := make([]string, len(valueMethods))
		copy(allMethods, valueMethods)
		allMethods = append(allMethods, pointerMethods...)
		
		implementsAsPointer := hasAllMethods(allMethods, requiredNames)

		if implementsAsValue || implementsAsPointer {
			allImpls = append(allImpls, Implementation{
				Struct:             sym.Name,
				Package:            sym.Package,
				PointerReceiver:    !implementsAsValue,
				ImplementedMethods: requiredNames,
				MissingMethods:     []string{},
			})
		}
	}

	sortImplementations(allImpls, opts.SortBy)

	if opts.Max > 0 && len(allImpls) > opts.Max {
		allImpls = allImpls[:opts.Max]
	}

	res.Implementations = allImpls
	res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)

	return res, nil
}

func sortImplementations(impls []Implementation, sortBy string) {
	sort.Slice(impls, func(i, j int) bool {
		if sortBy == "name" {
			if impls[i].Struct != impls[j].Struct {
				return impls[i].Struct < impls[j].Struct
			}
		}
		// Default: package then name
		if impls[i].Package != impls[j].Package {
			return impls[i].Package < impls[j].Package
		}
		return impls[i].Struct < impls[j].Struct
	})
}

func hasAllMethods(provided []string, required []string) bool {
	providedMap := make(map[string]bool)
	for _, m := range provided {
		providedMap[m] = true
	}
	for _, m := range required {
		if !providedMap[m] {
			return false
		}
	}
	return true
}

func (e *Engine) extractInterfaceMethods(sym *analyzer.Symbol, getFileLines func(string) []string) []interfaceMethod {
	lines := getFileLines(sym.FilePath)
	if len(lines) == 0 {
		return nil
	}
	
	var methods []interfaceMethod
	start := sym.StartLine - 1
	end := sym.EndLine
	if start < 0 { start = 0 }
	if end > len(lines) { end = len(lines) }

	for i := start; i < end; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "//") || line == "{" || line == "}" || strings.HasPrefix(line, "type ") {
			continue
		}
		
		if !strings.Contains(line, "(") {
			embeddedName := strings.Split(line, " ")[0]
			if strings.Contains(embeddedName, ".") {
				parts := strings.Split(embeddedName, ".")
				embeddedName = parts[1]
			}
			embeddedSyms := e.repo.ByName[embeddedName]
			for _, idx := range embeddedSyms {
				es := e.repo.Symbols[idx]
				if es.Kind == "Interface" {
					methods = append(methods, e.extractInterfaceMethods(&es, getFileLines)...)
					break
				}
			}
			continue
		}

		idx := strings.Index(line, "(")
		if idx > 0 {
			namePart := strings.TrimSpace(line[:idx])
			words := strings.Fields(namePart)
			if len(words) > 0 {
				methodName := words[len(words)-1]
				methods = append(methods, interfaceMethod{
					Name:      methodName,
					Signature: line,
				})
			}
		}
	}
	
	seen := make(map[string]bool)
	var unique []interfaceMethod
	for _, m := range methods {
		if !seen[m.Name] && m.Name != "" {
			seen[m.Name] = true
			unique = append(unique, m)
		}
	}
	return unique
}
