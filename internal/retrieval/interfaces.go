package retrieval

import (
	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/context"
)

type RetrievalEngine interface {
	Retrieve(query Query) (*RetrievalBundle, error)
}

type Query struct {
	Text     string
	Symbol   string
	Package  string
	File     string
	JSON     bool
	Verbose  bool
}

type RetrievalBundle struct {
	Question       string                  `json:"question"`
	Confidence     float64                 `json:"confidence"`
	RetrievalTime  string                  `json:"retrieval_time"`
	TopSymbols     []analyzer.Symbol       `json:"symbols"`
	TopFiles       []string                `json:"files"`
	TopPackages    []string                `json:"packages"`
	TopContexts    []*context.ContextResult `json:"context"`
	Dependencies   []string                `json:"dependencies"`
}
