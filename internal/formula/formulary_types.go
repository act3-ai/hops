package formula

import (
	"context"

	"github.com/act3-ai/hops/internal/platform"
)

// Client is the minimal interface for metadata and bottles
type Client interface {
	Formulary
	BottleRegistry
}

// Formulary types
type (
	// Formulary is a source of Formulae
	Formulary interface {
		Fetch(ctx context.Context, name string) (MultiPlatformFormula, error)
	}

	// PlatformFormulary is a source of Formulae that supports platform metadata
	PlatformFormulary interface {
		Formulary
		FetchPlatform(ctx context.Context, name string, plat platform.Platform) (PlatformFormula, error)
	}

	// ConcurrentFormulary is a source of Formulae that supports concurrent fetching
	ConcurrentFormulary interface {
		Formulary
		FetchAll(ctx context.Context, names []string) ([]MultiPlatformFormula, error)
	}

	// ConcurrentFormulary is a source of Formulae that supports platform metadata and concurrent fetching
	ConcurrentPlatformFormulary interface {
		PlatformFormulary
		ConcurrentFormulary
		FetchAllPlatform(ctx context.Context, names []string, plat platform.Platform) ([]PlatformFormula, error)
	}

	// NameLister is implemented by searchable Formularies
	NameLister interface {
		ListNames(ctx context.Context) ([]string, error)
	}

	// SearchableFormulary is a Formulary that supports search
	SearchableFormulary interface {
		Formulary
		NameLister
	}

	// SearchableFormulary is a Formulary that supports search and concurrent fetching
	ConcurrentSearchableFormulary interface {
		ConcurrentFormulary
		NameLister
	}
)
