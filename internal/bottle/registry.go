package bottle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/platform"
)

const (
	// AnnotationMetadataVersion is the annotation key used to describe the metadata version
	AnnotationMetadataVersion = "formulae.brew.sh/version"

	// MetadataVersionV1 is the value of the "formulae.brew.sh/version" annotation for the v1 API
	MetadataVersionV1 = "v1"

	// MetadataVersionV2 is the value of the "formulae.brew.sh/version" annotation for the v2 API
	MetadataVersionV2 = "v2"

	// MetadataVersionV3 is the value of the "formulae.brew.sh/version" annotation for the v3 API
	MetadataVersionV3 = "v3"
)

// Registry is a source of Bottles
type Registry interface {
	Repository(ctx context.Context, name string) (Repository, error)
}

// SearchableRegistry is a source supporting search
type SearchableRegistry interface {
	Registry
	RepositoryLister
}

// MetadataRegistry is a source of metadata
type MetadataRegistry interface {
	Metadata(ctx context.Context, name string) (*v1.Info, error)
	PlatformMetadata(ctx context.Context, name string, plat platform.Platform) (*v1.Info, error)
}

// RepositoryLister lists Bottle repositories
type RepositoryLister interface {
	Repositories(ctx context.Context) ([]string, error)
}

// Fetcher is unused
type Fetcher interface {
	Fetch(ctx context.Context, name string, plat platform.Platform) (io.ReadCloser, error)
	Metadata(ctx context.Context, name string) (*v1.Info, error)
	PlatformMetadata(ctx context.Context, name string, plat platform.Platform) (*v1.Info, error)
}

// Repository represents the minimum interface for a bottle repository
type Repository interface {
	oras.GraphTarget
	Name() string
}

// Remote defines a registry of bottles
type Remote struct {
	registry *remote.Registry
	path     string
}

// OCILayoutStore implements Repository
type remoteRepository struct {
	*remote.Repository
	name string
}

// NewRegistry initializes
func NewRegistry(prefix string, client remote.Client, plainHTTP bool) (*Remote, error) {
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

	return &Remote{
		registry: reg,
		path:     strings.TrimSuffix(path, "/"), // remove trailing slash
	}, nil
}

// repository produces a bottle repository
func (r *Remote) repository(ctx context.Context, name string) (*remoteRepository, error) {
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
	repo := repoi.(*remote.Repository)

	return &remoteRepository{
		Repository: repo,
		name:       name,
	}, nil
}

// Repositories lists bottle repositories
func (r *Remote) Repositories(ctx context.Context) ([]string, error) {
	return registry.Repositories(ctx, r.registry)
}

// Repository produces a bottle repository
func (r *Remote) Repository(ctx context.Context, name string) (Repository, error) {
	return r.repository(ctx, name)
}

// Name produces the name of the bottle stored
func (store *remoteRepository) Name() string {
	return store.name
}

// Local represents a collection of OCI-layout bottle dirs
type Local struct {
	Dir string
}

// NewLocal initializes
func NewLocal(dir string) *Local {
	return &Local{
		Dir: dir,
	}
}

// repository produces a bottle repository
func (r *Local) repository(ctx context.Context, name string) (*OCILayoutStore, error) {
	s, err := oci.NewWithContext(ctx, filepath.Join(r.Dir, brewfmt.Repo(name)))
	if err != nil {
		return nil, fmt.Errorf("initializing local storage for %s: %w", name, err)
	}

	return &OCILayoutStore{
		Store: s,
		name:  name,
	}, nil
}

// Repository produces a bottle repository
func (r *Local) Repository(ctx context.Context, name string) (Repository, error) {
	return r.repository(ctx, name)
}

// Repositories lists bottle repositories
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

// OCILayoutStore implements Repository
type OCILayoutStore struct {
	*oci.Store
	name string
}

// Name produces the name of the bottle stored
func (store *OCILayoutStore) Name() string {
	return store.name
}

// ListTags lists the tags available in a repository, only if the repository supports listing tags
func ListTags(ctx context.Context, repo Repository) ([]string, error) {
	lister, ok := repo.(registry.TagLister)
	if !ok {
		return nil, nil
	}

	tags, err := registry.Tags(ctx, lister)
	if err != nil {
		return nil, fmt.Errorf("listing bottle tags: %w", err)
	}

	return tags, nil
}
