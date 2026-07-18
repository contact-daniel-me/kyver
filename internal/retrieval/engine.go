package retrieval

import (
	"fmt"
	"time"

	"github.com/contact-daniel-me/kyver/internal/context"
	"github.com/contact-daniel-me/kyver/internal/graph"
	"github.com/contact-daniel-me/kyver/internal/navigation"
	"github.com/contact-daniel-me/kyver/internal/search"
)

type Engine struct {
	searchEngine *search.KeywordSearch
	navEngine    *navigation.Engine
	graphEngine  *graph.Engine
	ctxEngine    *context.Engine
	rootDir      string
}

func NewEngine(rootDir string) (*Engine, error) {
	repo, err := search.LoadRepository(rootDir)
	if err != nil {
		return nil, err
	}
	searchEngine := search.NewKeywordSearch(repo)
	
	navEngine, err := navigation.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}
	graphEngine, err := graph.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}
	ctxEngine, err := context.NewEngine(rootDir)
	if err != nil {
		return nil, err
	}

	return &Engine{
		searchEngine: searchEngine,
		navEngine:    navEngine,
		graphEngine:  graphEngine,
		ctxEngine:    ctxEngine,
		rootDir:      rootDir,
	}, nil
}

func (e *Engine) Retrieve(query Query) (*RetrievalBundle, error) {
	start := time.Now()

	bundle := &RetrievalBundle{
		Question: query.Text,
	}

	// Direct target
	if query.Symbol != "" {
		res, err := e.ctxEngine.BuildSymbolContext(query.Symbol, context.ContextOptions{})
		if err == nil {
			bundle.TopContexts = append(bundle.TopContexts, res)
			bundle.Confidence = 1.0
		}
	} else if query.Package != "" {
		res, err := e.ctxEngine.BuildPackageContext(query.Package, context.ContextOptions{})
		if err == nil {
			bundle.TopContexts = append(bundle.TopContexts, res)
			bundle.Confidence = 1.0
		}
	} else if query.File != "" {
		res, err := e.ctxEngine.BuildFileContext(query.File, context.ContextOptions{})
		if err == nil {
			bundle.TopContexts = append(bundle.TopContexts, res)
			bundle.Confidence = 1.0
		}
	} else if query.Text != "" {
		// Natural language retrieval
		err := e.retrieveFromQuestion(query.Text, bundle)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no valid query provided")
	}

	// Fill in TopSymbols, TopFiles, etc. from contexts if available
	populateBundleMetadata(bundle)

	bundle.RetrievalTime = time.Since(start).String()
	return bundle, nil
}
