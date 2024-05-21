package regbottle

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"

	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

// successorsForPlatform creates a platform-specific function to list successors used by Hops
func successorsForPlatform(plat platform.Platform) func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	if plat == platform.All {
		return content.Successors
	}

	return func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		switch desc.MediaType {
		case ocispec.MediaTypeImageIndex,
			"application/vnd.docker.distribution.manifest.list.v2+json":
			index, err := orasutil.FetchDecode[ocispec.Index](ctx, fetcher, desc)
			if err != nil {
				return nil, err
			}

			// Return nodes for platform and subject
			sel, err := platform.SelectManifest(index, plat)
			if err != nil {
				return nil, err
			}
			nodes := []ocispec.Descriptor{sel}
			if index.Subject != nil {
				nodes = append(nodes, *index.Subject)
			}
			return nodes, nil
		case ocispec.MediaTypeImageManifest,
			"application/vnd.docker.distribution.manifest.v2+json":
			manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, fetcher, desc)
			if err != nil {
				return nil, err
			}

			// Return nodes for config of metadata type and subject
			var nodes []ocispec.Descriptor

			// Copy subjects
			if manifest.Subject != nil {
				nodes = append(nodes, *manifest.Subject)
			}

			// Copy config if it's interesting
			if manifest.Config.MediaType == "application/vnd.brew.formula.metadata.v1+json" {
				nodes = append(nodes, manifest.Config)
			}

			// Copy layers
			return append(nodes, manifest.Layers...), nil
		default:
			return nil, nil
		}
	}
}

// metadataSuccessorsForPlatform creates a platform-specific function to list successors used as metadata by Hops
func metadataSuccessorsForPlatform(plat platform.Platform) func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	if plat == platform.All {
		return metadataSuccessors
	}

	return func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		switch desc.MediaType {
		case "application/vnd.docker.distribution.manifest.v2+json",
			ocispec.MediaTypeImageManifest:
			manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, fetcher, desc)
			if err != nil {
				return nil, err
			}

			// Return nodes for config of metadata type and subject
			var nodes []ocispec.Descriptor
			if manifest.Subject != nil {
				nodes = append(nodes, *manifest.Subject)
			}

			// Copy config if it's interesting
			if manifest.Config.MediaType == "application/vnd.brew.formula.metadata.v1+json" {
				nodes = append(nodes, manifest.Config)
			}

			return nodes, nil
		case ocispec.MediaTypeImageIndex,
			"application/vnd.docker.distribution.manifest.list.v2+json":
			index, err := orasutil.FetchDecode[ocispec.Index](ctx, fetcher, desc)
			if err != nil {
				return nil, err
			}

			// Return nodes for platform and subject
			sel, err := platform.SelectManifest(index, plat)
			if err != nil {
				return nil, err
			}
			nodes := []ocispec.Descriptor{sel}
			if index.Subject != nil {
				nodes = append(nodes, *index.Subject)
			}
			return nodes, nil
		}
		return nil, nil
	}
}

// metadataSuccessors lists successors used as metadata by Hops
func metadataSuccessors(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
	switch desc.MediaType {
	case "application/vnd.docker.distribution.manifest.v2+json",
		ocispec.MediaTypeImageManifest:
		manifest, err := orasutil.FetchDecode[ocispec.Manifest](ctx, fetcher, desc)
		if err != nil {
			return nil, err
		}

		// Return nodes for config of metadata type and subject
		var nodes []ocispec.Descriptor
		if manifest.Subject != nil {
			nodes = append(nodes, *manifest.Subject)
		}
		// Copy config if it's interesting
		if manifest.Config.MediaType == "application/vnd.brew.formula.metadata.v1+json" {
			nodes = append(nodes, manifest.Config)
		}
		return nodes, nil
	case ocispec.MediaTypeImageIndex,
		"application/vnd.docker.distribution.manifest.list.v2+json":
		index, err := orasutil.FetchDecode[ocispec.Index](ctx, fetcher, desc)
		if err != nil {
			return nil, err
		}

		nodes := []ocispec.Descriptor{}
		if index.Subject != nil {
			nodes = append(nodes, *index.Subject)
		}
		return nodes, nil
	}
	return nil, nil
}
