package status

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

// ReadIndex reads and parses the .kyver/index.json file
func ReadIndex(indexPath string) (*indexer.IndexData, error) {
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index not found at %s. run 'kyver index' first", indexPath)
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var indexData indexer.IndexData
	if err := json.Unmarshal(data, &indexData); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &indexData, nil
}
