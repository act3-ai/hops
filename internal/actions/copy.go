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
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/brew"
	"github.com/act3-ai/hops/internal/brewfile"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/dependencies"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/orasutil"
)

type copiedBottle struct {
	repo      oras.GraphTarget
	info      *brewv1.Info
	indexDesc ocispec.Descriptor
	index     *ocispec.Index
}

// Copy represents the action and its options
type Copy struct {
	*Hops
	// DependencyOptions dependencies.Options
	DependencyOptions formula.DependencyTags

	File string // path to a Brewfile specifying formulae dependencies

	From          string // source registry for bottles
	FromOCILayout bool   // use OCI layout directory as source
	FromPlainHTTP bool   // allow insecure connections to source registry without SSL check

	To          string // destination registry for bottles
	ToOCILayout bool   // use OCI layout directory as destination
	ToPlainHTTP bool   // allow insecure connections to destination registry without SSL check
}

// Run runs the action
func (action *Copy) Run(ctx context.Context, args []string) error {
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

		args = append(args, bf.Formula...)
	}

	o.H1("Copying:\n" + strings.Join(args, " "))

	slog.Debug("copying bottles", slog.Any("names", args))

	// Resolve all formulae
	formulae, err := action.resolve(ctx, args)
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
			info: f,
		}
	}

	err = action.copy(ctx, sources, copiedBottles)
	if err != nil {
		return err
	}

	o.Hai(fmt.Sprintf("Copied %d bottles", len(formulae)))

	return nil
}

func (action *Copy) resolve(ctx context.Context, args []string) ([]*brewv1.Info, error) {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return nil, err
	}

	// all, err := action.FetchAll(o.H1, index, formulae...)
	// if err != nil {
	// 	return nil, err
	// }

	// o.H1("Fetching dependencies...")
	// deps, err := dependencies.Walk(ctx, apiwalker.New(index, platform.All), all, &action.DependencyOptions)
	// if err != nil {
	// 	return nil, err
	// }

	// dependents := deps.Dependents()
	// fmt.Printf("Found %d dependencies\n", len(dependents))

	// // Combine root formulae with their dependencies in this list
	// all = append(all, dependents...)

	all, err := action.fetchFromArgs(ctx, args, platform.All)
	if err != nil {
		return nil, err
	}

	store, err := action.FormulaClient(ctx, args)
	if err != nil {
		return nil, err
	}

	o.H1("Fetching dependencies...")
	deps, err := dependencies.WalkAll(ctx, store, all, &action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	dependents := deps.Dependents()
	fmt.Printf("Found %d dependencies\n", len(dependents))

	// Combine root formulae with their dependencies in this list
	all = append(all, dependents...)
	metadata := make([]*brewv1.Info, 0, len(all))
	for _, f := range all {
		md := index.Find(f.Name())
		if md == nil {
			return nil, brew.NewErrFormulaNotFound(f.Name())
		}
		metadata = append(metadata, md)
	}

	return metadata, nil
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
		// IMPORTANT
		// Remove time-sensitive metadata fields
		f.info.Installed = []brewv1.InstalledInfo{} // empty "installed" list
		f.info.TapGitHead = ""                      // remove "tap_git_head" which changes even when formula metadata does not

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

func copyBottleArtifacts(ctx context.Context, src, dst oras.GraphTarget, f *brewv1.Info) (ocispec.Descriptor, error) {
	tag, err := f.ManifestTag(brewv1.Stable)
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

func metadatManifestOptions(name, version string, subject, metadata ocispec.Descriptor) oras.PackManifestOptions {
	docs := "https://formulae.brew.sh/formula/" + name
	return oras.PackManifestOptions{
		Subject: &subject,
		// TODO: add Ruby source as a layer
		Layers: []ocispec.Descriptor{},
		ManifestAnnotations: map[string]string{
			"formulae.brew.sh/version": "v1",
			// this timestamp will be automatically generated by oras.PackManifest() if not specified
			// use a fixed value here in order to have reproducible images
			ocispec.AnnotationCreated:       "1970-01-01T00:00:00Z", // POSIX epoch
			ocispec.AnnotationURL:           docs,
			ocispec.AnnotationDocumentation: docs,
			ocispec.AnnotationSource:        docs + ".json",
			ocispec.AnnotationVersion:       version,
			ocispec.AnnotationVendor:        "hops",
			// gh 2.49.2 metadata
			ocispec.AnnotationTitle: name + " " + version + " metadata",
		},
		ConfigDescriptor: &metadata,
	}
}

func platformMetadatManifestOptions(name, version string, plat platform.Platform, subject, metadata ocispec.Descriptor) oras.PackManifestOptions {
	opts := metadatManifestOptions(name, version, subject, metadata)

	// gh 2.49.2.arm64_monterey metadata
	opts.ManifestAnnotations[ocispec.AnnotationTitle] = name + " " + version + "." + plat.String() + " metadata"
	return opts
}

// pushMetadata pushes metadata for the given manifest
func pushMetadata(ctx context.Context, dst oras.Target, manifestDesc ocispec.Descriptor, f *brewv1.Info) (ocispec.Descriptor, error) { //nolint:unparam
	l := slog.Default()

	var manifestOptions oras.PackManifestOptions

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

		configDesc, err := mustPushMetadataBlob(ctx, dst, hopsspec.MediaTypeFormulaMetadata, platformMetadataJSON)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("pushing platform metadata: %w", err)
		}
		l.Debug("Pushed metadata for platform", logutil.OCIPlatformValue(manifestDesc.Platform), slog.String("digest", configDesc.Digest.String()))

		configDesc.Platform = manifestDesc.Platform // preserve platform

		manifestOptions = platformMetadatManifestOptions(f.FullName, f.Version(), plat, manifestDesc, configDesc)
	} else {
		l = l.With(slog.String("bottle", f.FullName))

		// Marshal formula API information to JSON
		infoJSON, err := json.Marshal(f)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		configDesc, err := mustPushMetadataBlob(ctx, dst, hopsspec.MediaTypeFormulaMetadata, infoJSON)
		if err != nil {
			return ocispec.Descriptor{}, fmt.Errorf("pushing index metadata: %w", err)
		}
		l.Debug("Pushed metadata for index", slog.String("digest", configDesc.Digest.String()))

		manifestOptions = metadatManifestOptions(f.FullName, f.Version(), manifestDesc, configDesc)
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
