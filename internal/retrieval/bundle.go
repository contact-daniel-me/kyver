package retrieval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (e *Engine) SaveBundle(bundle *RetrievalBundle) error {
	dir := filepath.Join(e.rootDir, ".kyver", "retrieval")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create retrieval directory: %w", err)
	}

	filename := fmt.Sprintf("bundle_%d.json", time.Now().Unix())
	destPath := filepath.Join(dir, filename)

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bundle)
}
