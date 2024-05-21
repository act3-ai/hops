package brew

import (
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	brewreg "github.com/act3-ai/hops/internal/brew/registry"
	"github.com/act3-ai/hops/internal/formula"
)

// Client is the interface for Homebrew's metadata and bottle stores
type Client interface {
	formula.Client
	formula.Formulary
	formula.ConcurrentBottleRegistry
}

type preloadedClient struct {
	brewformulary.Preloaded
	brewreg.BottleStore
}

// NewHomebrewClient creates a new BrewClient
func NewHomebrewClient(formulary brewformulary.Preloaded, bottles brewreg.BottleStore) Client {
	return &preloadedClient{
		Preloaded:   formulary,
		BottleStore: bottles,
	}
}
