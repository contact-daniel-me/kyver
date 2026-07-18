package indexer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteIndex saves the IndexData to the specified path in pretty JSON format.
func WriteIndex(indexPath string, data IndexData) error {
	// Ensure directory exists
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
