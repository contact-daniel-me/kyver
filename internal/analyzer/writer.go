package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SymbolWriter struct {
	file    *os.File
	encoder *json.Encoder
	first   bool
}

func NewSymbolWriter(filePath string, kyverVersion string) (*SymbolWriter, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	header := fmt.Sprintf("{\n  \"schemaVersion\": \"1.0\",\n  \"kyverVersion\": \"%s\",\n  \"generatedAt\": \"%s\",\n  \"symbols\": [\n", kyverVersion, time.Now().Format(time.RFC3339))
	if _, err := f.WriteString(header); err != nil {
		f.Close()
		return nil, err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("    ", "  ")

	return &SymbolWriter{
		file:    f,
		encoder: encoder,
		first:   true,
	}, nil
}

func (w *SymbolWriter) Write(sym Symbol) error {
	if !w.first {
		if _, err := w.file.WriteString(",\n"); err != nil {
			return err
		}
	}
	w.first = false

	// We want to write the encoded JSON without trailing newline to keep the comma formatting clean if needed,
	// but encoder.Encode adds a newline anyway, so we accept it.
	return w.encoder.Encode(sym)
}

func (w *SymbolWriter) Close() error {
	if _, err := w.file.WriteString("  ]\n}\n"); err != nil {
		w.file.Close()
		return err
	}
	return w.file.Close()
}

func WriteJSON(filePath string, data interface{}) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
