package analyzer

import "time"

// FileMetrics represents the metrics extracted from a single file
type FileMetrics struct {
	Structs        int
	Interfaces     int
	Functions      int
	Methods        int
	Constants      int
	Variables      int
	Aliases        int
	Exported       int
	Imports        int
	TotalFuncLines int
	MaxFuncLines   int
	CommentsCount  int
	LOC            int
}

// Metrics represents the aggregated metrics for the repository
type Metrics struct {
	SchemaVersion      string    `json:"schemaVersion"`
	KyverVersion       string    `json:"kyverVersion"`
	GeneratedAt        time.Time `json:"generatedAt"`
	Packages           int       `json:"packages"`
	Structs            int       `json:"structs"`
	Interfaces         int       `json:"interfaces"`
	Functions          int       `json:"functions"`
	Methods            int       `json:"methods"`
	Constants          int       `json:"constants"`
	Variables          int       `json:"variables"`
	Aliases            int       `json:"aliases"`
	ExportedSymbols    int       `json:"exportedSymbols"`
	Imports            int       `json:"imports"`
	AverageFuncLength  int       `json:"averageFunctionLength"`
	LargestFunction    int       `json:"largestFunction"`
	CommentRatio       float64   `json:"commentRatio"`
	TotalFuncLines     int       `json:"-"`
	TotalFunctions     int       `json:"-"`
	TotalCommentsCount int       `json:"-"`
	TotalLOC           int       `json:"-"`
}

func NewMetrics(kyverVersion string) *Metrics {
	return &Metrics{
		SchemaVersion: "1.0",
		KyverVersion:  kyverVersion,
		GeneratedAt:   time.Now(),
	}
}

func (m *Metrics) AddFileMetrics(fm FileMetrics) {
	m.Structs += fm.Structs
	m.Interfaces += fm.Interfaces
	m.Functions += fm.Functions
	m.Methods += fm.Methods
	m.Constants += fm.Constants
	m.Variables += fm.Variables
	m.Aliases += fm.Aliases
	m.ExportedSymbols += fm.Exported
	m.Imports += fm.Imports
	m.TotalFuncLines += fm.TotalFuncLines
	m.TotalFunctions += fm.Functions + fm.Methods
	if fm.MaxFuncLines > m.LargestFunction {
		m.LargestFunction = fm.MaxFuncLines
	}
	m.TotalCommentsCount += fm.CommentsCount
	m.TotalLOC += fm.LOC
}

func (m *Metrics) Finalize(packageCount int) {
	m.Packages = packageCount
	if m.TotalFunctions > 0 {
		m.AverageFuncLength = m.TotalFuncLines / m.TotalFunctions
	}
	if m.TotalLOC > 0 {
		m.CommentRatio = float64(m.TotalCommentsCount) / float64(m.TotalLOC)
	}
}
