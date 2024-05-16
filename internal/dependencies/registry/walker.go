package regwalker

import (
	"context"
	"slices"

	"github.com/sourcegraph/conc/iter"

	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/bottle/regbottle"
	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/platform"
)

type registryStore struct {
	bottle.Registry
	cache    bottle.Registry // download all bottles to this cache
	plat     platform.Platform
	routines iter.Mapper[string, *regbottle.BottleIndex]
}

// New creates a dependency walker for an OCI registry
func New(reg bottle.Registry, cache bottle.Registry, plat platform.Platform, maxGoroutines int) dependencies.Store[*regbottle.BottleIndex] {
	return &registryStore{
		Registry: reg,
		cache:    cache,
		plat:     plat,
		routines: iter.Mapper[string, *regbottle.BottleIndex]{MaxGoroutines: maxGoroutines},
	}
}

// Key creates cache keys
func (store *registryStore) Key(btl *regbottle.BottleIndex) string {
	return btl.RepositoryName
}

// DirectDependencies evaluates the direct dependencies of a node
func (store *registryStore) DirectDependencies(ctx context.Context, btl *regbottle.BottleIndex, opts *dependencies.Options) ([]*regbottle.BottleIndex, error) {
	deps, err := store.directDependencyNames(ctx, btl, opts)
	if err != nil {
		return nil, err
	}

	// Download all dependency metadata
	return store.routines.MapErr(deps, func(s *string) (*regbottle.BottleIndex, error) {
		src, err := store.Repository(ctx, *s)
		if err != nil {
			return nil, err
		}

		cache, err := store.cache.Repository(ctx, *s)
		if err != nil {
			return nil, err
		}

		depbtl, err := regbottle.ResolveVersion(ctx, src, "latest")
		if err != nil {
			return nil, err
		}

		err = regbottle.CopyTargetPlatform(ctx, src, cache, depbtl, store.plat)
		if err != nil {
			return nil, err
		}

		return depbtl, nil
	})
}

// directDependencyNames evaluates the direct dependency names
func (store *registryStore) directDependencyNames(ctx context.Context, btl *regbottle.BottleIndex, opts *dependencies.Options) ([]string, error) {
	var deps []string

	repo, err := store.Repository(ctx, btl.RepositoryName)
	if err != nil {
		return nil, err
	}

	// Return all possible dependencies
	if store.plat == platform.All {
		info, err := btl.GeneralMetadata(ctx, repo)
		if err != nil {
			return nil, err
		}

		// Add all variations
		deps = dependencies.ForOptions(&info.PlatformInfo, opts)
		for _, variation := range info.Variations {
			deps = append(deps, dependencies.ForOptions(variation, opts)...)
		}

		slices.Sort(deps)                // sort dependencies
		return slices.Compact(deps), nil // remove duplicates
	}

	info, err := btl.PlatformMetadata(ctx, repo, store.plat)
	if err != nil {
		return nil, err
	}

	return dependencies.ForOptions(info, opts), nil
}
