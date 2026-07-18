package search

import "strings"

func (s *KeywordSearch) References(symbol string) ([]string, error) {
	defs, _ := s.Definition(symbol)
	if len(defs) == 0 {
		return nil, nil
	}

	// Assume top result is the primary definition
	targetPkg := defs[0].Package

	var references []string
	for file, imports := range s.repo.Deps.FileDeps {
		for _, imp := range imports {
			// Check if import path matches package name directly or as a suffix
			if imp == targetPkg || strings.HasSuffix(imp, "/"+targetPkg) {
				references = append(references, file)
				break
			}
		}
	}
	
	return references, nil
}
