package indexer

import (
	"time"
)

// FileMetadata represents the extracted information for a single indexed file.
type FileMetadata struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Language     string    `json:"language"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	Hash         string    `json:"hash"`
	Module       string    `json:"module,omitempty"`
}

// IndexData represents the root structure of the index file.
type IndexData struct {
	Repository string         `json:"repository"`
	IndexedAt  time.Time      `json:"indexedAt"`
	Files      []FileMetadata `json:"files"`
}
