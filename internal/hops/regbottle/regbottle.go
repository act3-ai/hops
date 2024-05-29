package regbottle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	hopsspec "github.com/act3-ai/hops/internal/apis/annotations.hops.io"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

var (
	// ErrTagNotFound is returned when a tag is not found.
	ErrTagNotFound = fmt.Errorf("tag %w", errdef.ErrNotFound)

	// ErrNoMetadata is returned when metadata is not found.
	ErrNoMetadata = fmt.Errorf("metadata %w", errdef.ErrNotFound)
)

// BottleIndex represents a versioned bottle.
type BottleIndex struct {
	ocispec.Descriptor                                       // descriptor of the index
	index              *ocispec.Index                        // content of the bottle index
	metadata           *metadataManifest                     // descriptor of the metadata manifest, if found
	platforms          map[platform.Platform]*bottleManifest // maps platform to their respective bottle manifest
}

// bottleManifest represents a versioned bottle for a platform.
type bottleManifest struct {
	ocispec.Descriptor                     // descriptor of the manifest
	manifest           *ocispec.Manifest   // content of the bottle manifest
	metadata           *metadataManifest   // descriptor of the metadata manifest, if found
	bottle             *ocispec.Descriptor // descriptor of the bottle layer
	// index              *BottleIndex        // backlink to the index containing this manifest
}

// metadataManifest represents a bottle metadata manifest.
type metadataManifest struct {
	ocispec.Descriptor                   // descriptor of the manifest
	manifest           *ocispec.Manifest // content of the manifest
	config             *metadataConfig   // descriptor of the config
}

// metadataManifest represents a bottle metadata config.
type metadataConfig struct {
	ocispec.Descriptor              // descriptor of the config
	config             *brewv1.Info // content of the config
}

// ResolveVersion resolves a bottle version.
func ResolveVersion(ctx context.Context, repo oras.ReadOnlyGraphTarget, version string) (*BottleIndex, error) {
	d, err := repo.Resolve(ctx, version)
	if errors.Is(err, errdef.ErrNotFound) {
		return nil, fmt.Errorf("resolving version %q: %w", version, ErrTagNotFound)
	} else if err != nil {
		return nil, fmt.Errorf("resolving version %q: %w", version, err)
	}

	return &BottleIndex{
		Descriptor: d,
		platforms:  map[platform.Platform]*bottleManifest{},
	}, nil
}

// resolvePlatform resolves the bottle manifest for a platform.
func resolvePlatform(ctx context.Context, repo oras.ReadOnlyGraphTarget, bottle *BottleIndex, plat platform.Platform) (*bottleManifest, error) {
	if p, ok := bottle.platforms[plat]; ok {
		return p, nil
	}

	if bottle.index == nil {
		index, err := orasutil.FetchDecode[ocispec.Index](ctx, repo, bottle.Descriptor)
		if err != nil {
			return nil, fmt.Errorf("fetching index: %w", err)
		}
		bottle.index = index
	}

	sel, err := platform.SelectManifest(bottle.index, plat)
	if err != nil {
		return nil, err
	}

	// Return selected manifest
	bottle.platforms[plat] = &bottleManifest{
		Descriptor: sel,
		// index:      bottle,
	}
	return bottle.platforms[plat], nil
}

// ResolveBottle resolves the bottle artifact from a platform manifest.
func (btl *BottleIndex) ResolveBottle(ctx context.Context, repo oras.ReadOnlyGraphTarget, plat platform.Platform) (ocispec.Descriptor, error) {
	bottleManifest, err := resolvePlatform(ctx, repo, btl, plat)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	return resolveBottle(ctx, repo, bottleManifest)
}

// resolveBottle resolves the bottle artifact from a platform manifest.
func resolveBottle(ctx context.Context, repo oras.ReadOnlyGraphTarget, desc *bottleManifest) (ocispec.Descriptor, error) {
	if desc.bottle != nil {
		return *desc.bottle, nil
	}

	if desc.manifest == nil {
		manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, repo, desc.Descriptor)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("fetching manifest: %w", err)
		}
		desc.manifest = manifest
	}

	if len(desc.manifest.Layers) == 0 {
		return ocispec.Descriptor{}, fmt.Errorf("%s: manifest has no layers", desc.Descriptor.Digest.Encoded())
	}

	for _, l := range desc.manifest.Layers {
		// match the expected media type
		if l.MediaType == hopsspec.MediaTypeBottleArchiveLayer {
			desc.bottle = &l
			return *desc.bottle, nil
		}
	}

	return ocispec.Descriptor{}, fmt.Errorf("%s: manifest has no layers with mediaType %s", desc.Descriptor.Digest.Encoded(), hopsspec.MediaTypeBottleArchiveLayer)
}

// GeneralMetadata returns the full metadata for the bottle.
func (btl *BottleIndex) GeneralMetadata(ctx context.Context, repo oras.ReadOnlyGraphTarget) (*brewv1.Info, error) {
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

// ResolvePlatformMetadata resolves the platform-specific metadata for a bottle.
func (btl *BottleIndex) ResolvePlatformMetadata(ctx context.Context, repo oras.ReadOnlyGraphTarget, plat platform.Platform) (ocispec.Descriptor, error) {
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

// PlatformMetadata returns platform-specific metadata for a bottle.
func (btl *BottleIndex) PlatformMetadata(ctx context.Context, repo oras.ReadOnlyGraphTarget, plat platform.Platform) (*brewv1.PlatformInfo, error) {
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
		return nil, fmt.Errorf("platform metadata cannot contain variations: contains variations %v", keys)
	}

	return &info.PlatformInfo, nil
}

// resolveFullMetadata resolves the bottle metadata manifest for a platform.
func resolveFullMetadata(ctx context.Context, repo oras.ReadOnlyGraphTarget, desc *BottleIndex) (*metadataManifest, error) {
	if desc.metadata != nil {
		return desc.metadata, nil
	}

	referrers, err := registry.Referrers(ctx, repo, desc.Descriptor, hopsspec.ArtifactTypeHopsMetadata)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle metadata: %w", err)
	}

	if len(referrers) == 0 {
		return nil, fmt.Errorf("fetching general metadata: %w", ErrNoMetadata)
	}

	desc.metadata = &metadataManifest{Descriptor: referrers[0]}
	return desc.metadata, nil
}

// resolvePlatformMetadata resolves the bottle metadata manifest for a platform.
func resolvePlatformMetadata(ctx context.Context, repo oras.ReadOnlyGraphTarget, desc *bottleManifest) (*metadataManifest, error) {
	if desc.metadata != nil {
		return desc.metadata, nil
	}

	referrers, err := registry.Referrers(ctx, repo, desc.Descriptor, hopsspec.ArtifactTypeHopsMetadata)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle metadata: %w", err)
	}

	if len(referrers) == 0 {
		return nil, fmt.Errorf("fetching platform metadata: %w", ErrNoMetadata)
	}

	desc.metadata = &metadataManifest{Descriptor: referrers[0]}
	return desc.metadata, nil
}

// resolveMetadataConfig resolves metadata config.
func resolveMetadataConfig(ctx context.Context, repo oras.ReadOnlyGraphTarget, desc *metadataManifest) (*metadataConfig, error) {
	if desc.manifest == nil {
		manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, repo, desc.Descriptor)
		if err != nil {
			return nil, fmt.Errorf("fetching manifest: %w", err)
		}
		desc.manifest = manifest
	}

	desc.config = &metadataConfig{Descriptor: desc.manifest.Config}
	return desc.config, nil
}

// fetchMetadataConfig fetches the metadata config.
func fetchMetadataConfig(ctx context.Context, repo oras.ReadOnlyGraphTarget, desc *metadataConfig) (*brewv1.Info, error) {
	if desc.config == nil {
		config, err := orasutil.FetchDecode[brewv1.Info](ctx, repo, desc.Descriptor)
		if err != nil {
			desc.config = &brewv1.Info{}
			return nil, fmt.Errorf("fetching metadata from config: %w", err)
		}
		desc.config = config
	}

	return desc.config, nil
}

// CopyGeneralMetadata copies all bottle metadata artifacts.
func CopyGeneralMetadata(ctx context.Context, src oras.ReadOnlyGraphTarget, dst oras.GraphTarget, btl *BottleIndex) error {
	opts := copyOptions()
	opts.FindSuccessors = metadataSuccessors
	opts.CopyGraphOptions.PreCopy = ensureAllSuccessors(dst)

	// Add logging last
	opts.CopyGraphOptions = logutil.WithLogging(slog.Default(), slog.LevelDebug, &opts.CopyGraphOptions)

	if err := oras.ExtendedCopyGraph(ctx, src, dst, btl.Descriptor, opts); err != nil {
		return fmt.Errorf("copying metadata: %w", err)
	}

	return nil
}

// CopyPlatformMetadata copies all bottle artifacts for a given platform.
func CopyPlatformMetadata(ctx context.Context, src oras.ReadOnlyGraphTarget, dst oras.GraphTarget, btl *BottleIndex, plat platform.Platform) error {
	manifest, err := resolvePlatform(ctx, src, btl, plat)
	if err != nil {
		return err
	}

	opts := copyOptions()
	opts.FindSuccessors = metadataSuccessorsForPlatform(plat)
	opts.CopyGraphOptions.PreCopy = ensurePlatformSuccessors(dst, plat)

	// Add logging last
	opts.CopyGraphOptions = logutil.WithLogging(slog.Default(), slog.LevelDebug, &opts.CopyGraphOptions)

	if err := oras.ExtendedCopyGraph(ctx, src, dst, manifest.Descriptor, opts); err != nil {
		return fmt.Errorf("copying metadata for platform %s: %w", plat, err)
	}

	return nil
}

// CopyTargetPlatform copies all bottle artifacts for a given platform.
func CopyTargetPlatform(ctx context.Context, src oras.ReadOnlyGraphTarget, dst oras.GraphTarget, btl *BottleIndex, plat platform.Platform) error {
	manifest, err := resolvePlatform(ctx, src, btl, plat)
	if err != nil {
		return err
	}

	opts := oras.ExtendedCopyGraphOptions{}
	opts.CopyGraphOptions.FindSuccessors = successorsForPlatform(plat)
	opts.CopyGraphOptions.PreCopy = ensurePlatformSuccessors(dst, plat)

	// Add logging last
	opts.CopyGraphOptions = logutil.WithLogging(slog.Default(), slog.LevelDebug, &opts.CopyGraphOptions)

	if err := oras.ExtendedCopyGraph(ctx, src, dst, manifest.Descriptor, opts); err != nil {
		// if err := oras.CopyGraph(ctx, src, dst, manifest.Descriptor, opts.CopyGraphOptions); err != nil {
		return fmt.Errorf("copying bottle for platform %s: %w", plat, err)
	}

	return nil
}

// Copy copies all bottle artifacts.
func Copy(ctx context.Context, src oras.ReadOnlyGraphTarget, dst oras.GraphTarget, btl *BottleIndex) error {
	opts := copyOptions()
	opts.CopyGraphOptions.PreCopy = ensureAllSuccessors(dst)

	// Add logging last
	opts.CopyGraphOptions = logutil.WithLogging(slog.Default(), slog.LevelDebug, &opts.CopyGraphOptions)

	if err := oras.ExtendedCopyGraph(ctx, src, dst, btl.Descriptor, opts); err != nil {
		return fmt.Errorf("copying bottles: %w", err)
	}
	return nil
}

func copyOptions() oras.ExtendedCopyGraphOptions {
	return oras.ExtendedCopyGraphOptions{
		// Filter predecessors to Hops metadata
		FindPredecessors: findMetadataPredecessors,
	}
}

type (
	preCopyFunc        func(ctx context.Context, node ocispec.Descriptor) error
	findSuccessorsFunc func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error)
)

func ensureAllSuccessors(dst oras.Target) preCopyFunc {
	return func(ctx context.Context, node ocispec.Descriptor) error {
		// Check if manifest exists in dst
		nodeExists, err := dst.Exists(ctx, node)
		switch {
		// Error checking existence
		case err != nil:
			return fmt.Errorf("checking for node: %w", err)
		// Node exists in the destination
		case nodeExists:
			// List all successors of this node
			// Use the destination to pull from the cache rather than the remote
			successors, err := content.Successors(ctx, dst, node)
			if err != nil {
				return err
			}

			allSuccsExist, err := allExist(ctx, dst, successors)
			switch {
			// Error checking existence
			case err != nil:
				return fmt.Errorf("checking for successor nodes: %w", err)
			// All successors exists in the destination.
			// It is safe to skip this node.
			case allSuccsExist:
				return oras.SkipNode
			// There is at least one missing successor.
			// It is not safe to skip this node.
			default:
				return nil
			}
		// Node does not exist in the destination
		default:
			return nil
		}
	}
}

func ensurePlatformSuccessors(dst oras.Target, plat platform.Platform) preCopyFunc {
	return func(ctx context.Context, node ocispec.Descriptor) error {
		if node.Digest.Encoded() == "2a9f7c8f62d45593810a0fc926f6eec6e71a027774424240485c7d89bce15270" {
			fmt.Println("FOUND IT!!!")
		}
		// Check if manifest exists in dst
		nodeExists, err := dst.Exists(ctx, node)
		switch {
		// Error checking existence
		case err != nil:
			return err
		// Node exists in the destination
		case nodeExists:
			// List platform-specific successors of this node
			// Use the destination to pull from the cache rather than the remote
			successors, err := successorsForPlatform(plat)(ctx, dst, node)
			if err != nil {
				return err
			}

			allSuccsExist, err := allExist(ctx, dst, successors)
			switch {
			// Error checking existence
			case err != nil:
				return fmt.Errorf("checking for successor nodes: %w", err)
			// All successors exists in the destination.
			// It is safe to skip this node.
			case allSuccsExist:
				return oras.SkipNode
			// There is at least one missing successor.
			// It is not safe to skip this node.
			default:
				return nil
			}
		// Node does not exist in the destination
		default:
			return nil
		}
	}
}

func findMetadataPredecessors(ctx context.Context, src content.ReadOnlyGraphStorage, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	return registry.Referrers(ctx, src, desc, hopsspec.ArtifactTypeHopsMetadata)
}

func allExist(ctx context.Context, dst content.ReadOnlyStorage, targets []ocispec.Descriptor) (bool, error) {
	for _, target := range targets {
		exists, err := dst.Exists(ctx, target)
		switch {
		// Error checking existence
		case err != nil:
			return false, err
		// Target exists in the destination
		case exists:
		// Target is missing in the destination
		default:
			return false, nil // return here, there is at least one missing target
		}
	}
	return true, nil
}
