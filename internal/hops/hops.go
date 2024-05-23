package hopsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/sourcegraph/conc/iter"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/errdef"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/bottle"
	"github.com/act3-ai/hops/internal/hops/regbottle"
	hopsreg "github.com/act3-ai/hops/internal/hops/registry"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

// Client is an interface for combined metadata and bottle stores.
type Client interface {
	formula.ConcurrentPlatformFormulary
	bottle.ConcurrentRegistry
}

// NewClient creates a Hops formulary.
func NewClient(source hopsreg.Registry, cache *hopsreg.Local, alternateTags map[string]string, maxGoroutines int) Client {
	return &formulary{
		registry:      source,
		cache:         cache,
		tags:          alternateTags,
		resolved:      sync.Map{},
		maxGoroutines: maxGoroutines,
	}
}

// formulary is an OCI registry-backed formulary with caching and concurrency.
type formulary struct {
	registry      hopsreg.Registry
	cache         *hopsreg.Local
	tags          map[string]string // map names to special tags to use
	resolved      sync.Map
	maxGoroutines int
}

// FetchFormula implements formula.Formulary.
func (store *formulary) FetchFormula(ctx context.Context, name string) (formula.MultiPlatformFormula, error) {
	return store.fetch(ctx, name)
}

// FetchPlatformFormula implements formula.Formulary.
func (store *formulary) FetchPlatformFormula(ctx context.Context, name string, plat platform.Platform) (formula.PlatformFormula, error) {
	return store.fetchPlatform(ctx, name, plat)
}

// FetchFormulae implements formula.ConcurrentFormulary.
func (store *formulary) FetchFormulae(ctx context.Context, names []string) ([]formula.MultiPlatformFormula, error) {
	fetchers := iter.Mapper[string, formula.MultiPlatformFormula]{MaxGoroutines: store.maxGoroutines}
	return fetchers.MapErr(names, func(namep *string) (formula.MultiPlatformFormula, error) {
		f, err := store.fetch(ctx, *namep)
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", *namep, err)
		}
		return f, nil
	})
}

// FetchPlatformFormulae implements formula.ConcurrentFormulary.
func (store *formulary) FetchPlatformFormulae(ctx context.Context, names []string, plat platform.Platform) ([]formula.PlatformFormula, error) {
	fetchers := iter.Mapper[string, formula.PlatformFormula]{MaxGoroutines: store.maxGoroutines}
	return fetchers.MapErr(names, func(namep *string) (formula.PlatformFormula, error) {
		f, err := store.fetchPlatform(ctx, *namep, plat)
		if err != nil {
			return nil, fmt.Errorf("[%s] %w", *namep, err)
		}
		return f, nil
	})
}

// fetch fetches general metadata.
func (store *formulary) fetch(ctx context.Context, name string) (formula.MultiPlatformFormula, error) {
	source, err := store.registry.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	cache, err := store.cache.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	btl, err := store.resolve(ctx, name)
	if err != nil {
		return nil, err
	}

	err = regbottle.CopyGeneralMetadata(ctx, source, cache, btl)
	if err != nil {
		return nil, err
	}

	data, err := btl.GeneralMetadata(ctx, cache)
	if err != nil {
		return nil, err
	}

	return formula.FromV1(data), nil
}

// fetchPlatform fetches platform metadata.
func (store *formulary) fetchPlatform(ctx context.Context, name string, plat platform.Platform) (formula.PlatformFormula, error) {
	source, err := store.registry.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	cache, err := store.cache.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	btl, err := store.resolve(ctx, name)
	if err != nil {
		return nil, err
	}

	err = regbottle.CopyPlatformMetadata(ctx, source, cache, btl, plat)
	if err != nil {
		return nil, err
	}

	data, err := btl.PlatformMetadata(ctx, cache, plat)
	if err != nil {
		return nil, err
	}

	return formula.PlatformFromV1(plat, data), nil
}

func (store *formulary) resolve(ctx context.Context, name string) (*regbottle.BottleIndex, error) {
	if btl, ok := store.resolved.Load(name); ok && btl != nil {
		return btl.(*regbottle.BottleIndex), nil
	}

	source, err := store.registry.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	tag := store.tags[name]
	if tag == "" {
		tag = "latest"
	}

	btl, err := regbottle.ResolveVersion(ctx, source, tag)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, errors.Join(err, listAvailableTags(ctx, source, name))
		}
		return nil, err
	}
	store.resolved.Store(name, btl)

	return btl, nil
}

// FetchBottle implements formula.BottleRegistry.
func (store *formulary) FetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	return store.fetchBottle(ctx, f)
}

// FetchBottles implements formula.ConcurrentBottleRegistry.
func (store *formulary) FetchBottles(ctx context.Context, formulae []formula.PlatformFormula) ([]io.ReadCloser, error) {
	fetchers := iter.Mapper[formula.PlatformFormula, io.ReadCloser]{MaxGoroutines: store.maxGoroutines}
	return fetchers.MapErr(formulae, func(fp *formula.PlatformFormula) (io.ReadCloser, error) {
		return store.fetchBottle(ctx, *fp)
	})
}

// fetchBottle implements formula.BottleRegistry.
func (store *formulary) fetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	name := f.Name()

	source, err := store.registry.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	cache, err := store.cache.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	btl, err := store.resolve(ctx, name)
	if err != nil {
		return nil, err
	}

	// TODO: figure out why this was not copying the bottle blob
	// err = regbottle.CopyTargetPlatform(ctx, source, cache, btl, f.Platform())
	// if err != nil {
	// 	return nil, err
	// }

	btldesc, err := btl.ResolveBottle(ctx, cache, f.Platform())
	if err != nil {
		return nil, err
	}

	// Copy the bottle blob
	err = orasutil.CopyNode(ctx, source, cache, btldesc)
	if err != nil {
		return nil, err
	}

	// Fetch the bottle blob
	r, err := cache.Fetch(ctx, btldesc)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle from cache: %w", err)
	}

	return r, nil
}

func listAvailableTags(ctx context.Context, repo oras.ReadOnlyGraphTarget, name string) error {
	tags, err := hopsreg.ListTags(ctx, repo)
	if err != nil {
		o.Poo(fmt.Sprintf("[%s] Could not list available tags", name))
		return err
	}

	o.Hai(fmt.Sprintf("[%s] Found %d tags", name, len(tags)))
	if len(tags) > 0 {
		slices.Reverse(tags)
		if len(tags) > 10 {
			fmt.Println("Available tags:\n\t" + strings.Join(tags[:11], "\n\t") + "\n\t…")
			// fmt.Println("Available tags:\n\t" + strings.Join(tags[:11], "\n\t") + "\n\t...")
		} else {
			fmt.Println("Available tags:\n\t" + strings.Join(tags, "\n\t"))
		}
	}

	return nil
}
