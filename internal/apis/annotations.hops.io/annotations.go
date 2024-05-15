package hopsspec

import ocispec "github.com/opencontainers/image-spec/specs-go/v1"

const (
	// ArtifactTypeHopsMetadata is the artifactType of formula metadata stored in OCI by Hops
	ArtifactTypeHopsMetadata = "application/vnd.hops.formula.metadata.v1"

	// MediaTypeFormulaMetadata is the mediaType of formulae.brew.sh formula metadata stored in OCI by Hops
	MediaTypeFormulaMetadata = "application/vnd.brew.formula.metadata.v1+json"

	// MediaTypeFormulaSource is the mediaType of formula Ruby source stored in OCI by Hops
	MediaTypeFormulaSource = "application/vnd.brew.formula.source.v1"

	// MediaTypeBottleArchiveLayer is the mediaType used for bottle files stored in OCI
	MediaTypeBottleArchiveLayer = ocispec.MediaTypeImageLayerGzip // application/vnd.oci.image.layer.v1.tar+gzip
)
