package status

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

type LanguageStat struct {
	Language string `json:"language"`
	Count    int    `json:"count"`
}

type Statistics struct {
	TotalFiles      int            `json:"totalFiles"`
	Directories     int            `json:"directories"`
	RepositorySize  int64          `json:"repositorySize"`
	TotalLOC        int64          `json:"totalLOC"`
	Languages       []LanguageStat `json:"languages"`
	LargestFiles    []string       `json:"largestFiles"`
	RecentFiles     []string       `json:"recentFiles"`
}

func CalculateStats(data *indexer.IndexData) *Statistics {
	stats := &Statistics{}

	dirs := make(map[string]bool)
	langCounts := make(map[string]int)

	for _, file := range data.Files {
		stats.TotalFiles++
		stats.RepositorySize += file.Size
		
		dir := filepath.ToSlash(filepath.Dir(file.Path))
		if dir != "." && dir != "" {
			parts := strings.Split(dir, "/")
			currentPath := ""
			for _, part := range parts {
				if currentPath == "" {
					currentPath = part
				} else {
					currentPath = currentPath + "/" + part
				}
				dirs[currentPath] = true
			}
		}

		if file.Language != "" {
			langCounts[file.Language]++
		}
	}

	stats.Directories = len(dirs)
	stats.TotalLOC = stats.RepositorySize / 30

	for lang, count := range langCounts {
		stats.Languages = append(stats.Languages, LanguageStat{Language: lang, Count: count})
	}
	sort.Slice(stats.Languages, func(i, j int) bool {
		if stats.Languages[i].Count == stats.Languages[j].Count {
			return stats.Languages[i].Language < stats.Languages[j].Language
		}
		return stats.Languages[i].Count > stats.Languages[j].Count
	})

	filesBySize := make([]indexer.FileMetadata, len(data.Files))
	copy(filesBySize, data.Files)
	sort.Slice(filesBySize, func(i, j int) bool {
		return filesBySize[i].Size > filesBySize[j].Size
	})
	for i := 0; i < len(filesBySize) && i < 3; i++ {
		stats.LargestFiles = append(stats.LargestFiles, filesBySize[i].Path)
	}

	filesByTime := make([]indexer.FileMetadata, len(data.Files))
	copy(filesByTime, data.Files)
	sort.Slice(filesByTime, func(i, j int) bool {
		return filesByTime[i].LastModified.After(filesByTime[j].LastModified)
	})
	for i := 0; i < len(filesByTime) && i < 3; i++ {
		stats.RecentFiles = append(stats.RecentFiles, filesByTime[i].Path)
	}

	return stats
}
