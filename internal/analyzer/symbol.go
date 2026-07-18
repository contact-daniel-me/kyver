package analyzer

import "time"

// Symbol represents an extracted code structure (function, struct, etc.)
type Symbol struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Language      string `json:"language"`
	Package       string `json:"package"`
	Receiver      string `json:"receiver,omitempty"`
	Signature     string `json:"signature,omitempty"`
	FilePath      string `json:"file"`
	StartLine     int    `json:"startLine"`
	EndLine       int    `json:"endLine"`
	Exported      bool   `json:"exported"`
	Documentation string `json:"documentation,omitempty"`
}

// SymbolIndex is the root structure for symbols.json
type SymbolIndex struct {
	SchemaVersion string    `json:"schemaVersion"`
	KyverVersion  string    `json:"kyverVersion"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Symbols       []Symbol  `json:"symbols"`
}
