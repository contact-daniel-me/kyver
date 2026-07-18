package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	ErrPermissionDenied = fmt.Errorf("permission denied to write index")
	ErrCorruptedGit     = fmt.Errorf("repository appears corrupted or invalid")
	ErrIO               = fmt.Errorf("i/o error occurred during indexing")
)

// Indexer represents the engine for scanning and indexing repositories.
type Indexer struct {
	rootDir string
}

// Result holds the summary of an indexing operation.
type Result struct {
	FilesIndexed int
	IndexPath    string
	Duration     time.Duration
	IndexSize    int64
	MemoryUsed   uint64
	Stats        ScanStats
}

// New creates a new Indexer for the given root directory.
func New(rootDir string) *Indexer {
	return &Indexer{
		rootDir: rootDir,
	}
}

// Index performs the repository scan and saves the index data.
func (i *Indexer) Index(verbose bool) (Result, error) {
	start := time.Now()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// 1. Scan repository
	files, stats, err := Scan(i.rootDir, verbose)
	if err != nil {
		return Result{}, fmt.Errorf("failed to scan repository: %w", err)
	}

	// 2. Prepare IndexData
	repoName := filepath.Base(i.rootDir)
	data := IndexData{
		Repository: repoName,
		IndexedAt:  time.Now(),
		Files:      files,
	}

	// 3. Write index to disk
	indexPath := filepath.Join(i.rootDir, ".kyver", "index.json")
	if err := WriteIndex(indexPath, data); err != nil {
		if os.IsPermission(err) {
			return Result{}, fmt.Errorf("%w: %v", ErrPermissionDenied, err)
		}
		return Result{}, fmt.Errorf("%w: %v", ErrIO, err)
	}

	info, err := os.Stat(indexPath)
	var indexSize int64
	if err == nil {
		indexSize = info.Size()
	}

	runtime.ReadMemStats(&m2)
	duration := time.Since(start)

	return Result{
		FilesIndexed: len(files),
		IndexPath:    indexPath,
		Duration:     duration,
		IndexSize:    indexSize,
		MemoryUsed:   m2.Alloc - m1.Alloc,
		Stats:        stats,
	}, nil
}
