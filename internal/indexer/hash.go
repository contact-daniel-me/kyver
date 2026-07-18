package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// GenerateHash computes the SHA256 hash of a file without loading it all into memory.
func GenerateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
