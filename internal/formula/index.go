package formula

import (
	"context"
	"fmt"

	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

// Index defines a formula index
type Index interface {
	Find(name string) *Formula                           // finds a formula
	List() v1.Index                                      // lists all formulae
	ListNames() []string                                 // lists all formula names
	SearchFunc(func(matchFunc *Formula) bool) []*Formula // searches for a list of hits defined by the matchFunc
}

// CachedIndex represents a cached formula index
type CachedIndex interface {
	Index                                                // index functions
	Loader                                               // index loading functions
	Source() string                                      // returns the index's source
	IsCached() bool                                      // report if cached
	ShouldAutoUpdate(opts *hopsv1.AutoUpdateConfig) bool // reports if the cached index should be auto updated
	Reset(opts *hopsv1.AutoUpdateConfig) error           // resets the cached index when auto-updating
}

// Loader defines an index's load functions
type Loader interface {
	Load(ctx context.Context) error
}

// AutoUpdate performs an auto update of a cached index
func AutoUpdate(ctx context.Context, index CachedIndex, opts *hopsv1.AutoUpdateConfig) error {
	// Check if an auto-update should be performed
	if index.ShouldAutoUpdate(opts) {
		if err := index.Reset(opts); err != nil {
			return fmt.Errorf("auto updating formula index: %w", err)
		}
	}
	return index.Load(ctx)
}

// Load loads an index without auto-updating
// The index should only be downloaded if it was not already downloaded
func Load(ctx context.Context, index CachedIndex) error {
	return index.Load(ctx)
}

// type MapIndex map[string]*v1.Info
