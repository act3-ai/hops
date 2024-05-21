package formula

import (
	"context"

	"github.com/act3-ai/hops/internal/platform"
)

// Fetch fetches a Formula from the Formulary
func Fetch(ctx context.Context, src Formulary, name string) (MultiPlatformFormula, error) {
	return src.Fetch(ctx, name)
}

// FetchAll fetches Formulae from the Formulary
func FetchAll(ctx context.Context, src Formulary, names []string) ([]MultiPlatformFormula, error) {
	switch src := src.(type) {
	// Call FetchAll for Formulary with support for concurrency
	case ConcurrentFormulary:
		return src.FetchAll(ctx, names)
	// Call Fetch for all other Formulary kinds
	default:
		formulae := make([]MultiPlatformFormula, 0, len(names))
		for _, name := range names {
			f, err := src.Fetch(ctx, name)
			if err != nil {
				return nil, err
			}
			formulae = append(formulae, f)
		}
		return formulae, nil
	}
}

// FetchPlatform fetches a PlatformFormula from the Formulary
func FetchPlatform(ctx context.Context, src Formulary, name string, plat platform.Platform) (PlatformFormula, error) {
	switch src := src.(type) {
	case PlatformFormulary:
		return src.FetchPlatform(ctx, name, plat)
	default:
		f, err := src.Fetch(ctx, name)
		if err != nil {
			return nil, err
		}
		return f.ForPlatform(plat)
	}
}

// FetchAllPlatform fetches PlatformFormulae from the Formulary
func FetchAllPlatform(ctx context.Context, src Formulary, names []string, plat platform.Platform) ([]PlatformFormula, error) {
	switch src := src.(type) {
	case ConcurrentPlatformFormulary:
		return src.FetchAllPlatform(ctx, names, plat)
	case PlatformFormulary:
		formulae := make([]PlatformFormula, 0, len(names))
		for _, name := range names {
			f, err := src.FetchPlatform(ctx, name, plat)
			if err != nil {
				return nil, err
			}
			formulae = append(formulae, f)
		}
		return formulae, nil
	// Call FetchAll for Formulary with support for concurrency
	case ConcurrentFormulary:
		formulae, err := src.FetchAll(ctx, names)
		if err != nil {
			return nil, err
		}

		platformulae := make([]PlatformFormula, 0, len(names))
		for _, f := range formulae {
			platf, err := f.ForPlatform(plat)
			if err != nil {
				return nil, err
			}
			platformulae = append(platformulae, platf)
		}

		return platformulae, nil
	// Call Fetch for all other Formulary kinds
	default:
		platformulae := make([]PlatformFormula, 0, len(names))
		for _, name := range names {
			f, err := src.Fetch(ctx, name)
			if err != nil {
				return nil, err
			}
			platf, err := f.ForPlatform(plat)
			if err != nil {
				return nil, err
			}
			platformulae = append(platformulae, platf)
		}
		return platformulae, nil
	}
}
