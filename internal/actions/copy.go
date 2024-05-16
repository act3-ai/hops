package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"

	hopsspec "github.com/act3-ai/hops/internal/apis/annotations.hops.io"
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/brewfile"
	"github.com/act3-ai/hops/internal/dependencies"
	apiwalker "github.com/act3-ai/hops/internal/dependencies/api"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

type copiedBottle struct {
	repo      oras.GraphTarget
	info      *v1.Info
	indexDesc ocispec.Descriptor
	index     *ocispec.Index
}

// Copy represents the action and its options
type Copy struct {
	*Hops
	DependencyOptions dependencies.Options

	File string // path to a Brewfile specifying formulae dependencies

	From          string // source registry for bottles
	FromOCILayout bool   // use OCI layout directory as source
	FromPlainHTTP bool   // allow insecure connections to source registry without SSL check

	To          string // destination registry for bottles
	ToOCILayout bool   // use OCI layout directory as destination
	ToPlainHTTP bool   // allow insecure connections to destination registry without SSL check
}

// Run runs the action
func (action *Copy) Run(ctx context.Context, names ...string) error {
	if action.To == "" {
		return errors.New("empty destination registry")
	}

	if action.From == "" {
		return errors.New("empty source registry")
	}

	// Add Brewfile dependencies if requested
	if action.File != "" {
		bf, err := brewfile.Load(action.File)
		if err != nil {
			return err
		}

		names = append(names, bf.Formula...)
	}

	o.H1("Copying:\n" + strings.Join(names, " "))

	slog.Debug("copying bottles", slog.Any("names", names))

	// Resolve all formulae
	formulae, err := action.resolve(ctx, names...)
	if err != nil {
		return err
	}

	auth := action.Hops.AuthClient()

	// Initialize source and destination registries
	var srcReg bottle.Registry
	if action.FromOCILayout {
		srcReg = bottle.NewLocal(action.From)
	} else {
		srcReg, err = bottle.NewRegistry(action.From, auth, action.FromPlainHTTP)
		if err != nil {
			return fmt.Errorf("initializing source registry: %w", err)
		}
	}

	var dstReg bottle.Registry
	if action.ToOCILayout {
		dstReg = bottle.NewLocal(action.To)
	} else {
		dstReg, err = bottle.NewRegistry(action.To, auth, action.ToPlainHTTP)
		if err != nil {
			return fmt.Errorf("initializing destination registry: %w", err)
		}
	}

	// Initialize Bottle sources/destinations
	sources := make([]oras.GraphTarget, len(formulae))
	copiedBottles := make([]*copiedBottle, len(formulae))
	for i, f := range formulae {
		sources[i], err = srcReg.Repository(ctx, f.Name)
		if err != nil {
			return fmt.Errorf("creating source for %s: %w", f.Name, err)
		}
		dst, err := dstReg.Repository(ctx, f.Name)
		if err != nil {
			return fmt.Errorf("creating destination for %s: %w", f.Name, err)
		}

		copiedBottles[i] = &copiedBottle{
			repo: dst,
			info: &f.Info,
		}
	}

	err = action.copy(ctx, sources, copiedBottles)
	if err != nil {
		return err
	}

	o.Hai(fmt.Sprintf("Copied %d bottles", len(formulae)))

	return nil
}

func (action *Copy) resolve(ctx context.Context, formulae ...string) ([]*formula.Formula, error) {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return nil, err
	}

	all, err := action.FetchAll(o.H1, index, formulae...)
	if err != nil {
		return nil, err
	}

	o.H1("Fetching dependencies...")
	deps, err := dependencies.Walk(ctx, apiwalker.New(index, platform.All), all, &action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	dependents := deps.Dependents()
	fmt.Printf("Found %d dependencies\n", len(dependents))

	// Combine root formulae with their dependencies in this list
	all = append(all, dependents...)
	return all, nil
}

func (action *Copy) copy(ctx context.Context, sources []oras.GraphTarget, copiedBottles []*copiedBottle) error { //nolint:revive
	// Initialize Goroutine pool to reuse for each stage of the copy
	routines := pool.New().
		WithErrors().
		WithMaxGoroutines(action.MaxGoroutines())

	// Kick off routines to copy bottles
	for i, f := range copiedBottles {
		routines.Go(func() error {
			var err error
			f.indexDesc, err = copyBottleArtifacts(ctx, sources[i], f.repo, f.info)
			if err != nil {
				return err
			}
			return nil
		})
	}
	// TODO: add Ruby source copying here
	// Wait for all Bottles to be copied
	if err := routines.Wait(); err != nil {
		return err
	}

	// Kick off routines to fetch Bottle indexes
	for _, f := range copiedBottles {
		routines.Go(func() error {
			var err error
			f.index, err = orasutil.FetchDecode[ocispec.Index](ctx, f.repo, f.indexDesc)
			if err != nil {
				return fmt.Errorf("loading index: %w", err)
			}
			return nil
		})
	}
	// Wait all for all Bottle indexes to be loaded
	if err := routines.Wait(); err != nil {
		return err
	}

	// Kick off routines to push metadata
	for _, f := range copiedBottles {
		// Push metadata for the bottle index
		routines.Go(func() error {
			// o.Hai("Pushing metadata for " + f.Name)
			slog.Info("Pushing general metadata", slog.String("bottle", f.info.Name))
			if _, err := pushMetadata(ctx, f.repo, f.indexDesc, f.info); err != nil {
				return fmt.Errorf("[%s] failed to push metadata: %w", f.info.Name, err)
			}
			return nil
		})

		for _, manifestDesc := range f.index.Manifests {
			// Push metadata for each platform-specific manifest referenced by the bottle index
			routines.Go(func() error {
				slog.Info("Pushing metadata for platform",
					slog.String("bottle", f.info.Name+"/"+string(platform.FromOCI(manifestDesc.Platform))),
					logutil.OCIPlatformValue(manifestDesc.Platform))

				if _, err := pushMetadata(ctx, f.repo, manifestDesc, f.info); err != nil {
					return fmt.Errorf("[%s] failed to push metadata for platform %s/%s/%s: %w",
						f.info.Name,
						manifestDesc.Platform.OS, manifestDesc.Platform.Architecture, manifestDesc.Platform.OSVersion,
						err)
				}
				return nil
			})
		}
	}
	// Wait for all metadata manifests to be created
	if err := routines.Wait(); err != nil {
		return err
	}

	// Kick off routines to create "latest" tags
	for _, f := range copiedBottles {
		routines.Go(func() error {
			slog.Info("Creating \"latest\" tag", slog.String("bottle", f.info.Name))
			err := f.repo.Tag(ctx, f.indexDesc, "latest")
			if err != nil {
				return fmt.Errorf("creating tag %q for %s:%s", "latest", f.info.Name, f.info.Version())
			}
			return nil
		})
	}
	// Wait for all "latest" tags to be created
	if err := routines.Wait(); err != nil {
		return fmt.Errorf("tagging bottles: %w", err)
	}

	return nil
}

func copyBottleArtifacts(ctx context.Context, src, dst oras.GraphTarget, f *v1.Info) (ocispec.Descriptor, error) {
	tag, err := f.ManifestTag(v1.Stable)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("resolving tag: %w", err)
	}

	l := slog.Default().With(slog.String("bottle", f.Name+":"+tag))
	l.Info("Copying bottle artifacts")

	opts := oras.DefaultExtendedCopyOptions
	opts.CopyGraphOptions = logutil.WithLogging(l, slog.LevelInfo, &opts.CopyGraphOptions)

	md, err := oras.ExtendedCopy(ctx, src, tag, dst, "", opts)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("copying bottle %q: %w", f.Name, err)
	}

	l.Debug("Copied bottle artifacts", logutil.Descriptor(md))

	return md, nil
}

// pushMetadata pushes metadata for the given manifest
func pushMetadata(ctx context.Context, dst oras.Target, manifestDesc ocispec.Descriptor, f *v1.Info) (ocispec.Descriptor, error) { //nolint:unparam
	l := slog.Default()

	// Remove time-sensitive metadata fields
	f.Installed = []v1.InstalledInfo{} // empty "installed" list
	f.TapGitHead = ""                  // remove "tap_git_head" which changes even when formula metadata does not

	// Standard manifest options
	manifestOptions := oras.PackManifestOptions{
		Subject: &manifestDesc,
		// TODO: add Ruby source as a layer
		Layers: []ocispec.Descriptor{},
		ManifestAnnotations: map[string]string{
			"formulae.brew.sh/version": "v1",
			// this timestamp will be automatically generated by oras.PackManifest() if not specified
			// use a fixed value here in order to have reproducible images
			ocispec.AnnotationCreated: "1970-01-01T00:00:00Z", // POSIX epoch
			ocispec.AnnotationVendor:  "hops",
		},
	}

	if manifestDesc.Platform != nil {
		plat := platform.FromOCI(manifestDesc.Platform)

		l = l.With(slog.String("bottle", f.FullName+"/"+string(plat)))

		platformInfo, err := f.ForPlatform(plat)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		platformMetadataJSON, err := json.Marshal(platformInfo)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		blobDesc, err := mustPushMetadataBlob(ctx, dst, hopsspec.MediaTypeFormulaMetadata, platformMetadataJSON)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("pushing platform metadata: %w", err)
		}
		l.Debug("Pushed metadata for platform", logutil.OCIPlatformValue(manifestDesc.Platform), slog.String("digest", blobDesc.Digest.String()))

		blobDesc.Platform = manifestDesc.Platform // set platform

		manifestOptions.ConfigDescriptor = &blobDesc
		manifestOptions.ManifestAnnotations[ocispec.AnnotationTitle] = fmt.Sprintf("formulae.brew.sh/api/formula/%s.json %s", f.FullName, plat)
	} else {
		l = l.With(slog.String("bottle", f.FullName))

		// Marshal formula API information to JSON
		infoJSON, err := json.Marshal(f)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		blobDesc, err := mustPushMetadataBlob(ctx, dst, hopsspec.MediaTypeFormulaMetadata, infoJSON)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("pushing index metadata: %w", err)
		}
		l.Debug("Pushed metadata for index", slog.String("digest", blobDesc.Digest.String()))

		manifestOptions.ConfigDescriptor = &blobDesc
		manifestOptions.ManifestAnnotations[ocispec.AnnotationTitle] = fmt.Sprintf("formulae.brew.sh/api/formula/%s.json", f.FullName)
	}

	// Create metadata manifest referring to the bottle index
	metadataManifest, err := oras.PackManifest(ctx, dst, oras.PackManifestVersion1_1, hopsspec.ArtifactTypeHopsMetadata, manifestOptions)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	l.Debug("Pushed metadata manifest", slog.String("digest", metadataManifest.Digest.String()))
	return metadataManifest, nil
}

func mustPushMetadataBlob(ctx context.Context, dst oras.Target, mediaType string, metadata []byte) (ocispec.Descriptor, error) {
	d, err := oras.PushBytes(ctx, dst, mediaType, metadata)
	if errors.Is(err, errdef.ErrAlreadyExists) {
		d = content.NewDescriptorFromBytes(mediaType, metadata)
	} else if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("pushing metadata blob: %w", err)
	}
	return d, nil
}
