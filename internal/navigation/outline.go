package navigation

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func normalizeFilter(filter string) (string, error) {
	if filter == "" {
		return "", nil
	}
	f := strings.ToLower(filter)
	switch f {
	case "struct", "structs":
		return "Struct", nil
	case "interface", "interfaces":
		return "Interface", nil
	case "function", "functions":
		return "Function", nil
	case "method", "methods":
		return "Method", nil
	case "variable", "variables":
		return "Variable", nil
	case "constant", "constants":
		return "Constant", nil
	case "alias", "aliases":
		return "TypeAlias", nil
	default:
		return "", fmt.Errorf("Unknown filter %q\n\nSupported filters:\n\nstruct\ninterface\nfunction\nmethod\nvariable\nconstant\nalias", filter)
	}
}

func (e *Engine) Outline(file string, filter string, sortBy string) (*OutlineResult, error) {
	normalizedFilter, err := normalizeFilter(filter)
	if err != nil {
		return nil, err
	}

	var rawSymbols []analyzer.Symbol
	targetPath := filepath.ToSlash(file)

	for _, sym := range e.repo.Symbols {
		if strings.HasSuffix(filepath.ToSlash(sym.FilePath), targetPath) {
			if normalizedFilter != "" {
				kind := sym.Kind
				if kind == "Type" {
					kind = "TypeAlias"
				}
				if !strings.EqualFold(kind, normalizedFilter) {
					continue
				}
			}
			rawSymbols = append(rawSymbols, sym)
		}
	}

	if len(rawSymbols) == 0 {
		return nil, fmt.Errorf("no symbols found in file: %s", file)
	}

	groups := []string{"Structs", "Interfaces", "Type Aliases", "Constants", "Variables", "Functions", "Methods"}
	groupedMap := make(map[string][]analyzer.Symbol)
	
	stats := OutlineStats{
		TotalSymbols: len(rawSymbols),
		Counts:       make(map[string]int),
	}
	
	for _, sym := range rawSymbols {
		bucket := sym.Kind
		switch sym.Kind {
		case "Struct": bucket = "Structs"
		case "Interface": bucket = "Interfaces"
		case "TypeAlias", "Type": bucket = "Type Aliases"
		case "Constant": bucket = "Constants"
		case "Variable": bucket = "Variables"
		case "Function": bucket = "Functions"
		case "Method": bucket = "Methods"
		default: bucket = sym.Kind + "s"
		}
		groupedMap[bucket] = append(groupedMap[bucket], sym)
		stats.Counts[bucket]++
	}
	
	var resultGroups []SymbolGroup
	for _, gName := range groups {
		if syms, ok := groupedMap[gName]; ok && len(syms) > 0 {
			sort.Slice(syms, func(i, j int) bool {
				if sortBy == "name" {
					return strings.ToLower(syms[i].Name) < strings.ToLower(syms[j].Name)
				}
				return syms[i].StartLine < syms[j].StartLine
			})
			resultGroups = append(resultGroups, SymbolGroup{
				Kind:    gName,
				Symbols: syms,
			})
		}
	}
	
	for k, syms := range groupedMap {
		found := false
		for _, gName := range groups {
			if k == gName {
				found = true
				break
			}
		}
		if !found && len(syms) > 0 {
			sort.Slice(syms, func(i, j int) bool {
				if sortBy == "name" {
					return strings.ToLower(syms[i].Name) < strings.ToLower(syms[j].Name)
				}
				return syms[i].StartLine < syms[j].StartLine
			})
			resultGroups = append(resultGroups, SymbolGroup{
				Kind:    k,
				Symbols: syms,
			})
		}
	}

	var imports []string
	for depFile, depImports := range e.repo.Deps.FileDeps {
		if strings.HasSuffix(filepath.ToSlash(depFile), targetPath) {
			imports = depImports
			break
		}
	}
	stats.TotalImports = len(imports)

	lang := rawSymbols[0].Language
	if lang == "" {
		lang = "Go"
	}

	return &OutlineResult{
		File:     file,
		Package:  rawSymbols[0].Package,
		Language: lang,
		Imports:  imports,
		Groups:   resultGroups,
		Stats:    stats,
	}, nil
}
