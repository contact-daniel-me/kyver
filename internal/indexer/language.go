package indexer

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	goPackageRegex   = regexp.MustCompile(`(?m)^package\s+([a-zA-Z0-9_]+)`)
	javaPackageRegex = regexp.MustCompile(`(?m)^package\s+([a-zA-Z0-9_.]+);`)
	phpPackageRegex  = regexp.MustCompile(`(?m)^namespace\s+([a-zA-Z0-9_\\]+);`)
)

// SupportedExtensions maps file extensions to their programming languages.
var SupportedExtensions = map[string]string{
	".go":   "Go",
	".java": "Java",
	".py":   "Python",
	".js":   "JavaScript",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".jsx":  "React",
	".rs":   "Rust",
	".c":    "C/C++",
	".cpp":  "C/C++",
	".hpp":  "C/C++",
	".h":    "C/C++",
	".cs":   "C#",
	".kt":   "Kotlin",
	".swift": "Swift",
	".php":  "PHP",
	".rb":   "Ruby",
	".html": "HTML",
	".css":  "CSS",
	".scss": "SCSS",
	".json": "JSON",
	".yaml": "YAML",
	".yml":  "YAML",
	".md":   "Markdown",
	".sql":  "SQL",
	".tf":   "Terraform",
}

// SpecialFiles maps specific file names to their languages.
var SpecialFiles = map[string]string{
	"Dockerfile":         "Docker",
	"docker-compose.yml": "Docker",
}

// GetLanguage returns the language string if supported, or empty if not supported.
func GetLanguage(filename string) string {
	if lang, ok := SpecialFiles[filename]; ok {
		return lang
	}
	
	ext := strings.ToLower(filepath.Ext(filename))
	if lang, ok := SupportedExtensions[ext]; ok {
		return lang
	}
	return ""
}

// ExtractModule attempts to find the package/module name from the file content.
func ExtractModule(content []byte, language string) string {
	switch language {
	case "Go":
		matches := goPackageRegex.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	case "Java":
		matches := javaPackageRegex.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	case "PHP":
		matches := phpPackageRegex.FindSubmatch(content)
		if len(matches) > 1 {
			return string(matches[1])
		}
	}
	return ""
}
