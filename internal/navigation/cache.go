package navigation

import "github.com/contact-daniel-me/kyver/internal/search"

type Engine struct {
	repo *search.InMemoryRepository
}

func NewEngine(rootDir string) (*Engine, error) {
	repo, err := search.LoadRepository(rootDir)
	if err != nil {
		return nil, err
	}
	return &Engine{repo: repo}, nil
}

func NewEngineWithRepo(repo *search.InMemoryRepository) *Engine {
	return &Engine{repo: repo}
}
