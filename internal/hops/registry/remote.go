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

// Remote defines a registry of bottles.
type Remote struct {
	registry *remote.Registry
	path     string
}

// NewRemote initializes a remote registry.
func NewRemote(prefix string, client remote.Client, plainHTTP bool) (*Remote, error) {
	var host, path string

	// Split registry host and path
	// from: https://pkg.go.dev/oras.land/oras-go/v2@v2.5.0/registry#ParseReference
	parts := strings.SplitN(prefix, "/", 2)
	if len(parts) == 1 {
		host, path = parts[0], ""
	} else {
		host, path = parts[0], parts[1]
	}

	reg, err := remote.NewRegistry(host)
	if err != nil {
		return nil, err
	}
	reg.Client = client
	reg.PlainHTTP = plainHTTP
	reg.SkipReferrersGC = true

	return &Remote{
		registry: reg,
		path:     strings.TrimSuffix(path, "/"), // remove trailing slash
	}, nil
}

// Ping checks whether or not the registry implement Docker Registry API V2 or OCI Distribution Specification. Ping can be used to check authentication when an auth client is configured.
func (r *Remote) Ping(ctx context.Context) error {
	err := r.registry.Ping(ctx)
	if err != nil {
		return fmt.Errorf("pinging bottle registry %s: %w", r.registry.Reference.String(), err)
	}
	return nil
}

// Repositories lists bottle repositories.
func (r *Remote) Repositories(ctx context.Context) ([]string, error) {
	return registry.Repositories(ctx, r.registry)
}

// Repository produces a bottle repository.
func (r *Remote) Repository(ctx context.Context, name string) (oras.GraphTarget, error) {
	return r.repository(ctx, name)
}

// repository produces a bottle repository.
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
	return repoi.(*remote.Repository), nil //revive:disable:unchecked-type-assertion
}
