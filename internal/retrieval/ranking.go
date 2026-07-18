package retrieval

import (
	"strings"
)

// populateBundleMetadata extracts distinct symbols, files, and packages from the assembled contexts
func populateBundleMetadata(bundle *RetrievalBundle) {
	symSet := make(map[string]bool)
	fileSet := make(map[string]bool)
	pkgSet := make(map[string]bool)

	for _, ctx := range bundle.TopContexts {
		if ctx.Symbol != nil {
			if !symSet[ctx.Symbol.Name] {
				bundle.TopSymbols = append(bundle.TopSymbols, *ctx.Symbol)
				symSet[ctx.Symbol.Name] = true
			}
			fileSet[ctx.Symbol.FilePath] = true
			pkgSet[ctx.Symbol.Package] = true
		}
		
		if ctx.Target != "" {
			if ctx.Type == "File" {
				fileSet[ctx.Target] = true
			} else if ctx.Type == "Package" {
				pkgSet[ctx.Target] = true
			}
		}

		for _, m := range ctx.Methods {
			if !symSet[m.Name] {
				bundle.TopSymbols = append(bundle.TopSymbols, m)
				symSet[m.Name] = true
			}
		}
	}

	for f := range fileSet {
		bundle.TopFiles = append(bundle.TopFiles, f)
	}
	for p := range pkgSet {
		bundle.TopPackages = append(bundle.TopPackages, p)
	}
}

func calculateSymbolScore(keyword string, symbol string) float64 {
	lowerK := strings.ToLower(keyword)
	lowerS := strings.ToLower(symbol)
	if lowerK == lowerS {
		return 0.95 // exact match is high confidence
	}
	if strings.Contains(lowerS, lowerK) {
		return 0.75
	}
	return 0.5
}

func calculatePackageScore(keyword string, pkg string) float64 {
	lowerK := strings.ToLower(keyword)
	lowerP := strings.ToLower(pkg)
	if lowerK == lowerP {
		return 0.90
	}
	if strings.Contains(lowerP, lowerK) {
		return 0.60
	}
	return 0.4
}
