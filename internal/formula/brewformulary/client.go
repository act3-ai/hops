package brewformulary

import "github.com/act3-ai/hops/internal/formula"

// BrewClient is the interface for Homebrew's metadata and bottle stores
type BrewClient interface {
	formula.Client
	formula.Formulary
	formula.ConcurrentBottleRegistry
}

type preloadedClient struct {
	Preloaded
	BottleStore
}

// NewHomebrewClient creates a new BrewClient
func NewHomebrewClient(formulary Preloaded, bottles BottleStore) BrewClient {
	return &preloadedClient{
		Preloaded:   formulary,
		BottleStore: bottles,
	}
}
