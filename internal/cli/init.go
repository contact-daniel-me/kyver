package cli

import (
	"github.com/contact-daniel-me/kyver/internal/core/repository"
)

func runInit(args []string) error {
	repo := repository.New()
	return repo.Init()
}
