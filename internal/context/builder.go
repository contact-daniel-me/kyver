package context

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/contact-daniel-me/kyver/internal/graph"
	"github.com/contact-daniel-me/kyver/internal/navigation"
)

type Engine struct {
	nav   *navigation.Engine
	graph *graph.Engine
}

func NewEngine(rootDir string) (*Engine, error) {
	navEngine, err := navigation.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}
	graphEngine, err := graph.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}

	return &Engine{
		nav:   navEngine,
		graph: graphEngine,
	}, nil
}

func NewEngineWithEngines(nav *navigation.Engine, graph *graph.Engine) *Engine {
	return &Engine{
		nav:   nav,
		graph: graph,
	}
}

func (e *Engine) BuildSymbolContext(name string, opts ContextOptions) (*ContextResult, error) {
	start := time.Now()

	res := &ContextResult{}

	// Determine if a section should be included
	hasSection := func(sec string) bool {
		if opts.Full {
			return true
		}
		if len(opts.Sections) == 0 && opts.Summary {
			// Summary defaults
			switch sec {
			case "documentation", "package", "methods", "interfaces":
				return true
			}
			return false
		}
		if len(opts.Sections) == 0 {
			// Default full if nothing specified (unless summary was specified)
			return true
		}
		for _, s := range opts.Sections {
			if strings.EqualFold(s, sec) {
				return true
			}
		}
		return false
	}

	// 1. Resolve Symbol (Always required for context base)
	// We'll use graph.Engine's underlying symbol resolution by just calling Goto on nav engine
	gotoRes, err := e.nav.Goto(name, opts.SelectIndex)
	if err != nil {
		return nil, err
	}
	if gotoRes.Symbol == nil {
		if gotoRes.MultipleFound {
			res.MultipleFound = true
			res.Candidates = gotoRes.Candidates
			res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)
			return res, nil
		}
		return nil, fmt.Errorf("symbol not found")
	}

	sym := gotoRes.Symbol
	res.Symbol = sym

	if hasSection("documentation") {
		res.Documentation = sym.Documentation
	}

	if hasSection("package") {
		res.Package = &ContextPackage{
			Name: sym.Package,
		}
	}

	// 2. Execute requested sections
	if hasSection("methods") || opts.Summary {
		hierRes, _ := e.nav.Hierarchy(name, opts.SelectIndex)
		if hierRes != nil {
			if hasSection("methods") {
				res.Methods = hierRes.Methods
			}
			res.Statistics.Methods = len(hierRes.Methods)
		}
	}

	if hasSection("interfaces") || opts.Summary {
		ifaceRes, _ := e.graph.Interface(name, graph.InterfaceOptions{SelectIndex: opts.SelectIndex})
		if ifaceRes != nil {
			if hasSection("interfaces") {
				res.Interfaces = ifaceRes.Interfaces
			}
			res.Statistics.Interfaces = len(ifaceRes.Interfaces)
		}
	}

	if hasSection("implementations") || opts.Summary {
		implRes, _ := e.graph.Implements(name, graph.ImplementsOptions{SelectIndex: opts.SelectIndex})
		if implRes != nil {
			if hasSection("implementations") {
				res.Implementations = implRes.Implementations
			}
			// Statistics for implementations isn't explicitly required, but we can add it if needed.
		}
	}

	if hasSection("callers") || opts.Summary {
		callersRes, _ := e.graph.Callers(name, graph.CallersOptions{SelectIndex: opts.SelectIndex})
		if callersRes != nil {
			if hasSection("callers") {
				res.Callers = callersRes
			}
			res.Statistics.Callers = callersRes.Stats.TotalCallers
		}
	}

	if hasSection("callees") || opts.Summary {
		calleesRes, _ := e.graph.Callees(name, graph.CalleesOptions{SelectIndex: opts.SelectIndex})
		if calleesRes != nil {
			if hasSection("callees") {
				res.Callees = calleesRes
			}
			res.Statistics.Callees = calleesRes.Stats.TotalCallees
		}
	}

	if hasSection("references") || opts.Summary {
		refRes, _ := e.graph.References(name, graph.ReferenceOptions{SelectIndex: opts.SelectIndex})
		if refRes != nil {
			if hasSection("references") {
				res.References = refRes
			}
			res.Statistics.References = refRes.Stats.TotalReferences
		}
	}

	if hasSection("impacts") || opts.Summary {
		impRes, _ := e.graph.Impacts(name, graph.ImpactOptions{SelectIndex: opts.SelectIndex})
		if impRes != nil {
			if hasSection("impacts") {
				res.Impact = impRes
			}
		}
	}

	res.LookupTimeMs = math.Max(time.Since(start).Seconds()*1000, 0.001)

	return res, nil
}

func (e *Engine) BuildFileContext(file string, opts ContextOptions) (*ContextResult, error) {
	return &ContextResult{Type: "File", Target: file}, nil
}

func (e *Engine) BuildPackageContext(pkg string, opts ContextOptions) (*ContextResult, error) {
	return &ContextResult{Type: "Package", Target: pkg}, nil
}

func (e *Engine) BuildFunctionContext(function string, opts ContextOptions) (*ContextResult, error) {
	return &ContextResult{Type: "Function", Target: function}, nil
}
