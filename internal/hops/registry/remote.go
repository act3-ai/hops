package hopsreg

import (
	"context"
	"fmt"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
)

// Remote defines a registry of bottles
type Remote struct {
	registry *remote.Registry
	path     string
}

// // OCILayoutStore implements Repository
// type remoteRepository struct {
// 	*remote.Repository
// 	name string
// }

// NewRegistry initializes
func NewRegistry(ctx context.Context, prefix string, client remote.Client, plainHTTP bool) (*Remote, error) {
	var registry, path string

	// Split registry and path
	// from: https://pkg.go.dev/oras.land/oras-go/v2@v2.5.0/registry#ParseReference
	parts := strings.SplitN(prefix, "/", 2)
	if len(parts) == 1 {
		registry, path = parts[0], ""
	} else {
		registry, path = parts[0], parts[1]
	}

	reg, err := remote.NewRegistry(registry)
	if err != nil {
		return nil, err
	}
	reg.Client = client
	reg.PlainHTTP = plainHTTP

	err = reg.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("pinging bottle registry %s: %w", prefix, err)
	}

	return &Remote{
		registry: reg,
		path:     strings.TrimSuffix(path, "/"), // remove trailing slash
	}, nil
}

// Repositories lists bottle repositories
func (r *Remote) Repositories(ctx context.Context) ([]string, error) {
	return registry.Repositories(ctx, r.registry)
}

// Repository produces a bottle repository
func (r *Remote) Repository(ctx context.Context, name string) (oras.GraphTarget, error) {
	return r.repository(ctx, name)
}

// repository produces a bottle repository
func (r *Remote) repository(ctx context.Context, name string) (*remote.Repository, error) {
	var ref string

	if r.path == "" {
		ref = brewfmt.Repo(name)
	} else {
		ref = r.path + "/" + brewfmt.Repo(name)
	}

	repoi, err := r.registry.Repository(ctx, ref)
	if err != nil {
		return nil, err
	}
	return repoi.(*remote.Repository), nil
}
