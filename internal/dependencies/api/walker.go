package apiwalker

import (
	"context"
	"slices"

	"github.com/act3-ai/hops/internal/brew"
	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
)

// preloadedAPIStore is a dependency graph evaluator that uses the entire Homebrew API contents and is keyed by name
type preloadedAPIStore struct {
	index formula.Index
	plat  platform.Platform
}

// New creates a dependency graph evaluator that uses the entire Homebrew API contents and is keyed by name
func New(index formula.Index, plat platform.Platform) dependencies.Store[*formula.Formula] {
	return &preloadedAPIStore{
		index: index,
		plat:  plat,
	}
}

// Key creates cache keys
func (store *preloadedAPIStore) Key(f *formula.Formula) string {
	return f.Name
}

// DirectDependencies evaluates the direct dependencies of a node
func (store *preloadedAPIStore) DirectDependencies(_ context.Context, f *formula.Formula, opts *dependencies.Options) ([]*formula.Formula, error) {
	deps, err := directDependencies(f, store.plat, opts)
	if err != nil {
		return nil, err
	}

	result := make([]*formula.Formula, 0, len(deps))
	for _, dep := range deps {
		depinfo := store.index.Find(dep)
		if depinfo == nil {
			return nil, brew.NewErrFormulaNotFound(dep)
		}
		result = append(result, depinfo)
	}

	return result, nil
}

// directDependencies
func directDependencies(f *formula.Formula, plat platform.Platform, opts *dependencies.Options) ([]string, error) {
	var deps []string

	// Return all possible dependencies
	if plat == platform.All {
		// Add all variations
		deps = dependencies.ForOptions(&f.PlatformInfo, opts)
		for _, variation := range f.Variations {
			deps = append(deps, dependencies.ForOptions(variation, opts)...)
		}

		slices.Sort(deps)                // sort dependencies
		return slices.Compact(deps), nil // remove duplicates
	}

	pinfo, err := f.ForPlatform(plat)
	if err != nil {
		return nil, err
	}

	return dependencies.ForOptions(pinfo, opts), nil
}
