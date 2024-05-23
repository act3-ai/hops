package hopsreg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"

	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
)

// Local defines a local bottle store.
// The local directory stores oci-layout bottle repository dirs.
type Local struct {
	Dir string
}

// NewLocal initializes a local bottle store.
func NewLocal(dir string) *Local {
	return &Local{
		Dir: dir,
	}
}

// Repository produces a local bottle repository.
func (r *Local) Repository(ctx context.Context, name string) (oras.GraphTarget, error) {
	return r.repository(ctx, name)
}

// repository produces a local bottle repository.
func (r *Local) repository(ctx context.Context, name string) (*oci.Store, error) {
	dir := filepath.Join(r.Dir, brewfmt.Repo(name))
	s, err := oci.NewWithContext(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("initializing local storage for %s: %w", name, err)
	}

	return s, nil
}

// Repositories lists bottle repositories.
func (r *Local) Repositories(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(r.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("listing repositories: %w", err)
	}

	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}
