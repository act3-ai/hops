package hopsreg

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
)

const (
	// AnnotationMetadataVersion is the annotation key used to describe the metadata version.
	AnnotationMetadataVersion = "formulae.brew.sh/version"

	// MetadataVersionV1 is the value of the "formulae.brew.sh/version" annotation for the v1 API.
	MetadataVersionV1 = "v1"

	// MetadataVersionV2 is the value of the "formulae.brew.sh/version" annotation for the v2 API.
	MetadataVersionV2 = "v2"

	// MetadataVersionV3 is the value of the "formulae.brew.sh/version" annotation for the v3 API.
	MetadataVersionV3 = "v3"
)

// Registry stores Bottles.
type Registry interface {
	Repository(ctx context.Context, name string) (oras.GraphTarget, error)
}

// ListTags lists the tags available in a repository, only if the repository supports listing tags.
func ListTags(ctx context.Context, repo oras.ReadOnlyGraphTarget) ([]string, error) {
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
