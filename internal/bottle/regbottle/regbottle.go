package regbottle

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sourcegraph/conc/iter"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	hopsspec "github.com/act3-ai/hops/internal/apis/annotations.hops.io"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

var (
	// ErrTagNotFound is returned when a tag is not found
	ErrTagNotFound = fmt.Errorf("tag %w", errdef.ErrNotFound)

	// ErrNoMetadata is returned when metadata is not found
	ErrNoMetadata = fmt.Errorf("metadata %w", errdef.ErrNotFound)
)

// VersionedBottle represents a version of a Bottle
type VersionedBottle interface {
	Fetch(ctx context.Context, repo bottle.Repository, plat platform.Platform) (io.ReadCloser, error)
	Metadata(ctx context.Context, repo bottle.Repository) (*brewv1.Info, error)
	PlatformMetadata(ctx context.Context, repo bottle.Repository, plat platform.Platform) (*brewv1.Info, error)
}

// VersionMap maps bottle names to versions
// type VersionMap map[string]string

// type PlatformVersionedBottle interface {
// 	Fetch(ctx context.Context, repo bottle.Repository) (io.ReadCloser, error)
// 	Metadata(ctx context.Context, repo bottle.Repository) (*v1.PlatformInfo, error)
// }

// BottleIndex represents a versioned bottle
type BottleIndex struct {
	RepositoryName     string                                // name
	ocispec.Descriptor                                       // descriptor of the index
	index              *ocispec.Index                        // content of the bottle index
	metadata           *metadataManifest                     // descriptor of the metadata manifest, if found
	platforms          map[platform.Platform]*bottleManifest // maps platform to their respective bottle manifest
}

// bottleManifest represents a versioned bottle for a platform
type bottleManifest struct {
	ocispec.Descriptor                     // descriptor of the manifest
	manifest           *ocispec.Manifest   // content of the bottle manifest
	metadata           *metadataManifest   // descriptor of the metadata manifest, if found
	bottle             *ocispec.Descriptor // descriptor of the bottle layer
	// index              *BottleIndex        // backlink to the index containing this manifest
}

// metadataManifest represents a bottle metadata manifest
type metadataManifest struct {
	ocispec.Descriptor                   // descriptor of the manifest
	manifest           *ocispec.Manifest // content of the manifest
	config             *metadataConfig   // descriptor of the config
}

// metadataManifest represents a bottle metadata config
type metadataConfig struct {
	ocispec.Descriptor              // descriptor of the config
	config             *brewv1.Info // content of the config
}

// ResolveVersion resolves a bottle version
func ResolveVersion(ctx context.Context, repo bottle.Repository, version string) (*BottleIndex, error) {
	d, err := repo.Resolve(ctx, version)
	if errors.Is(err, errdef.ErrNotFound) {
		return nil, fmt.Errorf("[%s] resolving version %q: %w", repo.Name(), version, ErrTagNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("[%s] resolving version %q: %w", repo.Name(), version, err)
	}

	return &BottleIndex{
		RepositoryName: repo.Name(),
		Descriptor:     d,
		platforms:      map[platform.Platform]*bottleManifest{},
	}, nil
}

// resolvePlatform resolves the bottle manifest for a platform
func resolvePlatform(ctx context.Context, repo bottle.Repository, bottle *BottleIndex, plat platform.Platform) (*bottleManifest, error) {
	if p, ok := bottle.platforms[plat]; ok {
		return p, nil
	}

	if bottle.index == nil {
		index, err := orasutil.FetchDecode[ocispec.Index](ctx, repo, bottle.Descriptor)
		if err != nil {
			return nil, fmt.Errorf("[%s] fetching index: %w", repo.Name(), err)
		}
		bottle.index = index
	}

	sel := platform.SelectManifestIndex(bottle.index, plat)
	if sel < 0 {
		return nil, fmt.Errorf("[%s] selecting platform: no manifest for platform %s", repo.Name(), plat)
	}

	// Return selected manifest
	bottle.platforms[plat] = &bottleManifest{
		Descriptor: bottle.index.Manifests[sel],
		// index:      bottle,
	}
	return bottle.platforms[plat], nil
}

// ResolveBottle resolves the bottle artifact from a platform manifest
func (btl *BottleIndex) ResolveBottle(ctx context.Context, repo bottle.Repository, plat platform.Platform) (ocispec.Descriptor, error) {
	bottleManifest, err := resolvePlatform(ctx, repo, btl, plat)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	return resolveBottle(ctx, repo, bottleManifest)
}

// resolveBottle resolves the bottle artifact from a platform manifest
func resolveBottle(ctx context.Context, repo bottle.Repository, desc *bottleManifest) (ocispec.Descriptor, error) {
	if desc.bottle != nil {
		return *desc.bottle, nil
	}

	if desc.manifest == nil {
		manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, repo, desc.Descriptor)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("[%s] fetching manifest: %w", repo.Name(), err)
		}
		desc.manifest = manifest
	}

	if len(desc.manifest.Layers) == 0 {
		return ocispec.Descriptor{}, fmt.Errorf("[%s] %s: manifest has no layers", repo.Name(), desc.Descriptor.Digest.Encoded())
	}

	for _, l := range desc.manifest.Layers {
		// match the expected media type
		if l.MediaType == hopsspec.MediaTypeBottleArchiveLayer {
			desc.bottle = &l
			return *desc.bottle, nil
		}
	}

	return ocispec.Descriptor{}, fmt.Errorf("[%s] %s: manifest has no layers with mediaType %s", repo.Name(), desc.Descriptor.Digest.Encoded(), hopsspec.MediaTypeBottleArchiveLayer)
}

// GeneralMetadata returns the full metadata for the bottle
func (btl *BottleIndex) GeneralMetadata(ctx context.Context, repo bottle.Repository) (*brewv1.Info, error) {
	mdman, err := resolveFullMetadata(ctx, repo, btl)
	if err != nil {
		return nil, err
	}

	mdconfig, err := resolveMetadataConfig(ctx, repo, mdman)
	if err != nil {
		return nil, err
	}

	return fetchMetadataConfig(ctx, repo, mdconfig)
}

// ResolvePlatformMetadata resolves the platform-specific metadata for a bottle
func (btl *BottleIndex) ResolvePlatformMetadata(ctx context.Context, repo bottle.Repository, plat platform.Platform) (ocispec.Descriptor, error) {
	pman, err := resolvePlatform(ctx, repo, btl, plat)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	mdman, err := resolvePlatformMetadata(ctx, repo, pman)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	return mdman.Descriptor, nil
}

// PlatformMetadata returns platform-specific metadata for a bottle
func (btl *BottleIndex) PlatformMetadata(ctx context.Context, repo bottle.Repository, plat platform.Platform) (*brewv1.PlatformInfo, error) {
	pman, err := resolvePlatform(ctx, repo, btl, plat)
	if err != nil {
		return nil, err
	}

	mdman, err := resolvePlatformMetadata(ctx, repo, pman)
	if err != nil {
		return nil, err
	}

	mdconfig, err := resolveMetadataConfig(ctx, repo, mdman)
	if err != nil {
		return nil, err
	}

	info, err := fetchMetadataConfig(ctx, repo, mdconfig)
	if err != nil {
		return nil, err
	}

	if len(info.Variations) > 0 {
		keys := make([]string, 0, len(info.Variations))
		for p := range info.Variations {
			keys = append(keys, p.String())
		}
		return nil, fmt.Errorf("[%s] platform metadata cannot contain variations: contains variations %v", repo.Name(), keys)
	}

	return &info.PlatformInfo, nil
}

// resolveFullMetadata resolves the bottle metadata manifest for a platform
func resolveFullMetadata(ctx context.Context, repo bottle.Repository, desc *BottleIndex) (*metadataManifest, error) {
	if desc.metadata != nil {
		return desc.metadata, nil
	}

	referrers, err := registry.Referrers(ctx, repo, desc.Descriptor, hopsspec.ArtifactTypeHopsMetadata)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle metadata: %w", err)
	}

	if len(referrers) == 0 {
		return nil, fmt.Errorf("[%s] fetching general metadata: %w", repo.Name(), ErrNoMetadata)
	}

	desc.metadata = &metadataManifest{Descriptor: referrers[0]}
	return desc.metadata, nil
}

// resolvePlatformMetadata resolves the bottle metadata manifest for a platform
func resolvePlatformMetadata(ctx context.Context, repo bottle.Repository, desc *bottleManifest) (*metadataManifest, error) {
	if desc.metadata != nil {
		return desc.metadata, nil
	}

	referrers, err := registry.Referrers(ctx, repo, desc.Descriptor, hopsspec.ArtifactTypeHopsMetadata)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle metadata: %w", err)
	}

	if len(referrers) == 0 {
		return nil, fmt.Errorf("[%s] fetching platform metadata: %w", repo.Name(), ErrNoMetadata)
	}

	desc.metadata = &metadataManifest{Descriptor: referrers[0]}
	return desc.metadata, nil
}

// resolveMetadataConfig resolves metadata config
func resolveMetadataConfig(ctx context.Context, repo bottle.Repository, desc *metadataManifest) (*metadataConfig, error) {
	if desc.manifest == nil {
		manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, repo, desc.Descriptor)
		if err != nil {
			return nil, fmt.Errorf("[%s] fetching manifest: %w", repo.Name(), err)
		}
		desc.manifest = manifest
	}

	desc.config = &metadataConfig{Descriptor: desc.manifest.Config}
	return desc.config, nil
}

// fetchMetadataConfig fetches the metadata config
func fetchMetadataConfig(ctx context.Context, repo bottle.Repository, desc *metadataConfig) (*brewv1.Info, error) {
	if desc.config == nil {
		config, err := orasutil.FetchDecode[brewv1.Info](ctx, repo, desc.Descriptor)
		if err != nil {
			desc.config = &brewv1.Info{}
			return nil, fmt.Errorf("[%s] fetching metadata from config: %w", repo.Name(), err)
		}
		desc.config = config
	}

	return desc.config, nil
}

// CopyGeneralMetadata copies all bottle metadata artifacts
func CopyGeneralMetadata(ctx context.Context, src, dst bottle.Repository, btl *BottleIndex) error {
	opts := copyOptions
	opts.FindSuccessors = metadataSuccessors

	if err := oras.ExtendedCopyGraph(ctx, src, dst, btl.Descriptor, opts); err != nil {
		return fmt.Errorf("[%s] copying metadata: %w", src.Name(), err)
	}

	return nil
}

// CopyPlatformMetadata copies all bottle artifacts for a given platform
func CopyPlatformMetadata(ctx context.Context, src, dst bottle.Repository, btl *BottleIndex, plat platform.Platform) error {
	opts := copyOptions
	opts.FindSuccessors = metadataSuccessorsForPlatform(plat)

	manifest, err := resolvePlatform(ctx, src, btl, plat)
	if err != nil {
		return err
	}

	if err := oras.ExtendedCopyGraph(ctx, src, dst, manifest.Descriptor, opts); err != nil {
		return fmt.Errorf("[%s] copying metadata for platform %s: %w", src.Name(), plat, err)
	}

	return nil
}

// CopyTargetPlatform copies all bottle artifacts for a given platform
func CopyTargetPlatform(ctx context.Context, src, dst bottle.Repository, btl *BottleIndex, plat platform.Platform) error {
	opts := copyOptions
	opts.FindSuccessors = successorsForPlatform(plat)

	manifest, err := resolvePlatform(ctx, src, btl, plat)
	if err != nil {
		return err
	}

	if err := oras.ExtendedCopyGraph(ctx, src, dst, manifest.Descriptor, opts); err != nil {
		return fmt.Errorf("[%s] copying bottle for platform %s: %w", src.Name(), plat, err)
	}

	return nil
}

// Copy copies all bottle artifacts
func Copy(ctx context.Context, src, dst bottle.Repository, btl *BottleIndex) error {
	opts := copyOptions
	if err := oras.ExtendedCopyGraph(ctx, src, dst, btl.Descriptor, opts); err != nil {
		return fmt.Errorf("[%s] copying bottles: %w", src.Name(), err)
	}
	return nil
}

var copyOptions = oras.ExtendedCopyGraphOptions{
	// Add logging
	CopyGraphOptions: logutil.WithLogging(slog.Default(), slog.LevelDebug, &oras.DefaultCopyGraphOptions),
	// Filter predecessors to Hops metadata
	FindPredecessors: func(ctx context.Context, src content.ReadOnlyGraphStorage, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		return registry.Referrers(ctx, src, desc, hopsspec.ArtifactTypeHopsMetadata)
	},
}

// IterOptions configures iteration
type IterOptions struct {
	MaxGoroutines int
}

// List
func List(ctx context.Context, reg bottle.SearchableRegistry, opts *IterOptions) ([]*brewv1.Info, error) {
	repos, err := reg.Repositories(ctx)
	if err != nil {
		return nil, err
	}

	fetchers := iter.Mapper[string, *brewv1.Info]{
		MaxGoroutines: opts.MaxGoroutines,
	}

	return fetchers.MapErr(repos, func(repop *string) (*brewv1.Info, error) {
		repo, err := reg.Repository(ctx, *repop)
		if err != nil {
			return nil, err
		}

		btl, err := ResolveVersion(ctx, repo, "latest")
		if err != nil {
			return nil, err
		}

		info, err := btl.GeneralMetadata(ctx, repo)
		if err != nil {
			return nil, err
		}

		return info, nil
	})
}

// CopyAllMetadata copies all metadata artifacts from srcReg to dstReg
func CopyAllMetadata(ctx context.Context, srcReg, dstReg bottle.SearchableRegistry, opts *IterOptions) ([]*BottleIndex, error) {
	srcRepos, err := srcReg.Repositories(ctx)
	if err != nil {
		return nil, err
	}

	fetchers := iter.Mapper[string, *BottleIndex]{
		MaxGoroutines: opts.MaxGoroutines,
	}

	return fetchers.MapErr(srcRepos, func(repop *string) (*BottleIndex, error) {
		srcRepo, err := srcReg.Repository(ctx, *repop)
		if err != nil {
			return nil, err
		}

		dstRepo, err := dstReg.Repository(ctx, *repop)
		if err != nil {
			return nil, err
		}

		btl, err := ResolveVersion(ctx, srcRepo, "latest")
		if err != nil {
			return nil, err
		}

		err = CopyGeneralMetadata(ctx, srcRepo, dstRepo, btl)
		if err != nil {
			return nil, err
		}

		return btl, nil
	})
}
