package search

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

type InMemoryRepository struct {
	Symbols        []analyzer.Symbol
	Deps           *analyzer.DependencyGraph
	SymbolsModTime time.Time
	DepsModTime    time.Time

	// Lookup indexes
	ByName     map[string][]int
	ByPackage  map[string][]int
	ByKind     map[string][]int
	ByReceiver map[string][]int
	ByLanguage map[string][]int
	Exported   []int
}

var cache *InMemoryRepository

func LoadRepository(rootDir string) (*InMemoryRepository, error) {
	symPath := filepath.Join(rootDir, ".kyver", "symbols.json")
	depPath := filepath.Join(rootDir, ".kyver", "dependency_graph.json")

	symStat, err := os.Stat(symPath)
	if err != nil {
		return nil, err
	}
	depStat, err := os.Stat(depPath)
	if err != nil {
		return nil, err
	}

	if cache != nil {
		if cache.SymbolsModTime.Equal(symStat.ModTime()) && cache.DepsModTime.Equal(depStat.ModTime()) {
			return cache, nil
		}
	}

	symFile, err := os.Open(symPath)
	if err != nil {
		return nil, err
	}
	defer symFile.Close()

	var symIndex analyzer.SymbolIndex
	if err := json.NewDecoder(symFile).Decode(&symIndex); err != nil {
		return nil, err
	}

	depFile, err := os.Open(depPath)
	if err != nil {
		return nil, err
	}
	defer depFile.Close()

	var deps analyzer.DependencyGraph
	if err := json.NewDecoder(depFile).Decode(&deps); err != nil {
		return nil, err
	}

	cache = &InMemoryRepository{
		Symbols:        symIndex.Symbols,
		Deps:           &deps,
		SymbolsModTime: symStat.ModTime(),
		DepsModTime:    depStat.ModTime(),
		ByName:         make(map[string][]int),
		ByPackage:      make(map[string][]int),
		ByKind:         make(map[string][]int),
		ByReceiver:     make(map[string][]int),
		ByLanguage:     make(map[string][]int),
		Exported:       []int{},
	}

	for i, sym := range cache.Symbols {
		cache.ByName[sym.Name] = append(cache.ByName[sym.Name], i)
		cache.ByPackage[sym.Package] = append(cache.ByPackage[sym.Package], i)
		cache.ByKind[sym.Kind] = append(cache.ByKind[sym.Kind], i)
		if sym.Receiver != "" {
			cache.ByReceiver[sym.Receiver] = append(cache.ByReceiver[sym.Receiver], i)
		}
		cache.ByLanguage[sym.Language] = append(cache.ByLanguage[sym.Language], i)
		if sym.Exported {
			cache.Exported = append(cache.Exported, i)
		}
	}

	return cache, nil
}
