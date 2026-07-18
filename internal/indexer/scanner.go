package indexer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

var ignoredDirs = map[string]bool{
	".git":         true,
	".kyver":       true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"bin":          true,
	"obj":          true,
	".idea":        true,
	".vscode":      true,
}

// SkippedFile holds the path and reason for a file being skipped during indexing.
type SkippedFile struct {
	Path   string
	Reason string
}

// ScanStats holds granular statistics about the repository scan.
type ScanStats struct {
	Directories   int
	ExcludedDirs  []string
	Languages     map[string]int
	SkippedFiles  []SkippedFile
}

// Scan traverses the repository and collects metadata for all supported source files.
func Scan(rootDir string, verbose bool) ([]FileMetadata, ScanStats, error) {
	var metadataList []FileMetadata
	stats := ScanStats{
		Languages: make(map[string]int),
	}

	// Keep track of unique excluded dirs for summary
	excludedSet := make(map[string]bool)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "error reading file/dir"})
			}
			return nil
		}

		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				if _, exists := excludedSet[d.Name()]; !exists {
					excludedSet[d.Name()] = true
					stats.ExcludedDirs = append(stats.ExcludedDirs, d.Name())
				}
				return filepath.SkipDir
			}
			stats.Directories++
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "cannot get file info"})
			}
			return nil
		}

		// Skip symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "symbolic link"})
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "not a regular file"})
			}
			return nil
		}

		lang := GetLanguage(d.Name())
		if lang == "" {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "unsupported extension"})
			}
			return nil
		}

		// Handle invalid UTF-8 filenames
		if !utf8.ValidString(d.Name()) {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "invalid utf-8 filename"})
			}
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "cannot resolve relative path"})
			}
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		meta, err := processFile(path, relPath, d.Name(), info, lang)
		if err != nil {
			if verbose {
				stats.SkippedFiles = append(stats.SkippedFiles, SkippedFile{Path: path, Reason: "failed to process file"})
			}
			return nil
		}

		stats.Languages[lang]++
		metadataList = append(metadataList, meta)
		return nil
	})

	if err != nil {
		return nil, stats, fmt.Errorf("failed to scan repository: %w", err)
	}

	sort.Strings(stats.ExcludedDirs)
	return metadataList, stats, nil
}

func processFile(fullPath, relPath, name string, info fs.FileInfo, lang string) (FileMetadata, error) {
	hashStr, err := GenerateHash(fullPath)
	if err != nil {
		return FileMetadata{}, err
	}

	meta := FileMetadata{
		Path:         relPath,
		Name:         name,
		Extension:    strings.ToLower(filepath.Ext(name)),
		Language:     lang,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		Hash:         hashStr,
	}

	file, err := os.Open(fullPath)
	if err == nil {
		defer file.Close()
		buf := make([]byte, 4096)
		n, _ := file.Read(buf)
		if n > 0 {
			meta.Module = ExtractModule(buf[:n], lang)
		}
	}

	return meta, nil
}
