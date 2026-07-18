package graph

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

func (e *Engine) Interface(symbol string, opts InterfaceOptions) (*InterfaceResult, error) {
	start := time.Now()

	candidates, selected, err := e.resolveSymbol(symbol, opts.SelectIndex)
	if err != nil {
		return nil, err
	}

	res := &InterfaceResult{
		Candidates: candidates,
	}

	if selected == nil {
		res.MultipleFound = true
		return res, nil
	}
	res.Symbol = selected

	if selected.Kind != "Struct" {
		return nil, fmt.Errorf("symbol '%s' is a %s, not a Struct", selected.Name, selected.Kind)
	}

	// Build the method set of the struct
	var valueMethods []string
	for _, idx := range e.repo.ByReceiver[selected.Name] {
		valueMethods = append(valueMethods, e.repo.Symbols[idx].Name)
	}
	var pointerMethods []string
	for _, idx := range e.repo.ByReceiver["*"+selected.Name] {
		pointerMethods = append(pointerMethods, e.repo.Symbols[idx].Name)
	}

	allMethods := make([]string, len(valueMethods))
	copy(allMethods, valueMethods)
	allMethods = append(allMethods, pointerMethods...)
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

	var allInterfaces []ImplementedInterface
	packageSet := make(map[string]bool)

	for i := range e.repo.Symbols {
		sym := e.repo.Symbols[i]
		if sym.Kind != "Interface" {
			continue
		}

		if opts.PackageFilter != "" && sym.Package != opts.PackageFilter {
			continue
		}

		requiredInterfaceMethods := e.extractInterfaceMethods(&sym, getFileLines)
		if len(requiredInterfaceMethods) == 0 {
			// Empty interfaces are satisfied by everything.
			// But for kyver, usually we don't list interface{} (it's not declared anyway).
			// If a declared interface is empty, technically it implements it. 
			// Let's include it for completeness unless it's just 'interface{}'.
			continue
		}

		var requiredSignatures []string
		var requiredNames []string
		for _, m := range requiredInterfaceMethods {
			requiredSignatures = append(requiredSignatures, m.Signature)
			requiredNames = append(requiredNames, m.Name)
		}

		implementsAsValue := hasAllMethods(valueMethods, requiredNames)
		implementsAsPointer := hasAllMethods(allMethods, requiredNames)

		if implementsAsValue || implementsAsPointer {
			allInterfaces = append(allInterfaces, ImplementedInterface{
				Name:               sym.Name,
				Package:            sym.Package,
				RequiredMethods:    requiredSignatures,
				ImplementedMethods: requiredNames,
				MissingMethods:     []string{},
				PointerReceiver:    !implementsAsValue,
			})
			packageSet[sym.Package] = true
		}
	}

	sortInterfaces(allInterfaces, opts.SortBy)

	if opts.Max > 0 && len(allInterfaces) > opts.Max {
		allInterfaces = allInterfaces[:opts.Max]
	}

	res.Interfaces = allInterfaces
	res.Stats.TotalInterfaces = len(allInterfaces)
	res.Stats.TotalPackages = len(packageSet)
	res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)

	return res, nil
}

func sortInterfaces(ifaces []ImplementedInterface, sortBy string) {
	sort.Slice(ifaces, func(i, j int) bool {
		if sortBy == "name" {
			if ifaces[i].Name != ifaces[j].Name {
				return ifaces[i].Name < ifaces[j].Name
			}
		}
		// Default: package then name
		if ifaces[i].Package != ifaces[j].Package {
			return ifaces[i].Package < ifaces[j].Package
		}
		return ifaces[i].Name < ifaces[j].Name
	})
}
