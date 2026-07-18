package analyzer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

var (
	ErrInvalidIndex    = errors.New("invalid or corrupted repository index")
	ErrMissingRepo     = errors.New("repository not found")
	ErrUnsupportedLang = errors.New("unsupported or unknown language")
)

// AnalysisSummary encapsulates all extracted statistics.
type AnalysisSummary struct {
	TotalSymbols       int
	ExportedSymbols    int
	Packages           int
	Interfaces         int
	Structs            int
	Methods            int
	Functions          int
	Constants          int
	Variables          int
	Aliases            int
	DependencyEdges    int
	InternalImports    int
	ExternalImports    int
	CyclesDetected     int
	ArchitectureLayers []string
	WorkersCount       int
	Duration           time.Duration
	ArtifactsSize      int64
	FilesAnalyzed      int
	TotalIndexedFiles  int
	Languages          map[string]int
}

type Engine struct {
	rootDir      string
	analyzers    map[string]Analyzer
	kyverVersion string
}

func NewEngine(rootDir, kyverVersion string) *Engine {
	e := &Engine{
		rootDir:      rootDir,
		analyzers:    make(map[string]Analyzer),
		kyverVersion: kyverVersion,
	}

	e.RegisterAnalyzer(&GoAnalyzer{})
	return e
}

func (e *Engine) RegisterAnalyzer(a Analyzer) {
	e.analyzers[a.Language()] = a
}

func (e *Engine) Analyze(indexData *indexer.IndexData, targetLang string, verbose bool) (AnalysisSummary, error) {
	start := time.Now()
	var summary AnalysisSummary
	
	if indexData == nil || len(indexData.Files) == 0 {
		return summary, ErrInvalidIndex
	}
	
	summary.TotalIndexedFiles = len(indexData.Files)

	deps := NewDependencyGraph(e.kyverVersion)
	metrics := NewMetrics(e.kyverVersion)

	symPath := filepath.Join(e.rootDir, ".kyver", "symbols.json")
	symWriter, err := NewSymbolWriter(symPath, e.kyverVersion)
	if err != nil {
		return summary, err
	}
	defer symWriter.Close()

	numWorkers := runtime.NumCPU()
	summary.WorkersCount = numWorkers
	summary.Languages = make(map[string]int)

	jobs := make(chan indexer.FileMetadata, len(indexData.Files))
	results := make(chan *AnalysisResult, len(indexData.Files))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for file := range jobs {
				analyzer, ok := e.analyzers[file.Language]
				if !ok {
					results <- nil
					continue
				}

				if targetLang != "" && file.Language != targetLang {
					results <- nil
					continue
				}

				if verbose {
					fmt.Printf("[Worker %d] Analyzing: %s\n", workerID, file.Path)
				}

				fullPath := filepath.Join(e.rootDir, file.Path)
				res, err := analyzer.Analyze(fullPath, file)
				if err != nil {
					results <- nil
					continue
				}

				res.Metrics.LOC = int(file.Size / 30)
				results <- res
			}
		}(i)
	}

	for _, file := range indexData.Files {
		jobs <- file
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	packagesSet := make(map[string]bool)

	for res := range results {
		if res == nil {
			continue
		}

		metrics.AddFileMetrics(res.Metrics)
		if res.PackageName != "" {
			packagesSet[res.PackageName] = true
			deps.AddPackageDependency(res.PackageName, res.Imports)
		}

		deps.AddFileDependency(res.FilePath, res.Imports)

		for _, sym := range res.Symbols {
			if err := symWriter.Write(sym); err != nil {
				return summary, err
			}
			summary.TotalSymbols++
		}
		summary.FilesAnalyzed++
	}

	for _, file := range indexData.Files {
		summary.Languages[file.Language]++
	}

	metrics.Finalize(len(packagesSet))
	arch := GenerateArchitectureSummary(e.kyverVersion, deps)

	summary.ExportedSymbols = metrics.ExportedSymbols
	summary.Packages = metrics.Packages
	summary.Interfaces = metrics.Interfaces
	summary.Structs = metrics.Structs
	summary.Methods = metrics.Methods
	summary.Functions = metrics.Functions
	summary.Constants = metrics.Constants
	summary.Variables = metrics.Variables
	summary.Aliases = metrics.Aliases
	
	edges := 0
	for _, imp := range deps.FileDeps {
		edges += len(imp)
	}
	summary.DependencyEdges = edges
	summary.InternalImports = len(arch.InternalDependencies)
	summary.ExternalImports = len(arch.ExternalDependencies)
	summary.CyclesDetected = len(arch.PotentialCyclic)
	summary.ArchitectureLayers = arch.Packages

	depPath := filepath.Join(e.rootDir, ".kyver", "dependency_graph.json")
	if err := WriteJSON(depPath, deps); err != nil {
		return summary, err
	}
	
	metPath := filepath.Join(e.rootDir, ".kyver", "metrics.json")
	if err := WriteJSON(metPath, metrics); err != nil {
		return summary, err
	}
	
	archPath := filepath.Join(e.rootDir, ".kyver", "architecture.json")
	if err := WriteJSON(archPath, arch); err != nil {
		return summary, err
	}

	var artifactsSize int64
	for _, path := range []string{symPath, depPath, metPath, archPath} {
		if info, err := os.Stat(path); err == nil {
			artifactsSize += info.Size()
		}
	}
	summary.ArtifactsSize = artifactsSize
	summary.Duration = time.Since(start)

	return summary, nil
}
