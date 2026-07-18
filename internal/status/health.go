package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

// CheckHealth returns a list of recommendations based on repository analysis
func CheckHealth(rootDir string, data *indexer.IndexData) []string {
	var recommendations []string

	if data == nil {
		recommendations = append(recommendations, "Index is missing or corrupt. Run 'kyver index'.")
		return recommendations
	}

	gitPath := filepath.Join(rootDir, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		recommendations = append(recommendations, "Git is not initialized in this repository.")
	}

	hasReadme := false
	hasLicense := false

	for _, file := range data.Files {
		nameLower := strings.ToLower(file.Name)
		if file.Path == "README.md" || file.Path == "readme.md" {
			hasReadme = true
		}
		if nameLower == "license" || nameLower == "license.md" || nameLower == "license.txt" {
			hasLicense = true
		}
	}

	if _, err := os.Stat(filepath.Join(rootDir, ".gitignore")); err == nil {
		// hasGitignore = true
	} else {
		recommendations = append(recommendations, "Consider adding a .gitignore")
	}

	if !hasReadme {
		recommendations = append(recommendations, "Consider adding a README.md")
	}
	if !hasLicense {
		recommendations = append(recommendations, "Consider adding a LICENSE")
	}

	emptyFiles := 0
	largeFiles := 0
	paths := make(map[string]bool)
	duplicates := 0

	for _, file := range data.Files {
		if file.Size == 0 {
			emptyFiles++
		}
		if file.Size > 10*1024*1024 {
			largeFiles++
		}
		if paths[file.Path] {
			duplicates++
		} else {
			paths[file.Path] = true
		}
	}

	if emptyFiles > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Repository contains %d empty files.", emptyFiles))
	}
	if largeFiles > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d file(s) exceed 10MB.", largeFiles))
	}
	if duplicates > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Index contains %d duplicate paths. Consider re-indexing.", duplicates))
	}

	return recommendations
}
