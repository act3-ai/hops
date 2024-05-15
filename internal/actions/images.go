package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/sourcegraph/conc/iter"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/brewfile"
	"github.com/act3-ai/hops/internal/dependencies"
	apiwalker "github.com/act3-ai/hops/internal/dependencies/api"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
)

// Images represents the action and its options
type Images struct {
	*Hops
	DependencyOptions dependencies.Options

	File      string // path to a Brewfile specifying formulae dependencies
	NoResolve bool   // disable tag resolution
	NoVerify  bool   // disable tag verification
}

// Run runs the action
func (action *Images) Run(ctx context.Context, formulae ...string) error {
	// Add Brewfile dependencies if requested
	if action.File != "" {
		bf, err := brewfile.Load(action.File)
		if err != nil {
			return err
		}

		formulae = append(formulae, bf.Brew...)
	}

	slog.Debug("finding images for", slog.Any("formulae", formulae))

	index, err := action.findDeps(ctx, formulae...)
	if err != nil {
		return err
	}

	images, err := action.listImages(ctx, index)
	if err != nil {
		return err
	}

	imageData := strings.Join(images, "\n")
	o.Hai("Images:\n" + imageData)
	b, err := json.Marshal(index)
	if err != nil {
		return err
	}

	o.Hai("Writing image list and formula index")

	imagesFile := "hops.images.txt"
	indexFile := "hops.index.json"

	if err = os.WriteFile(imagesFile, []byte(imageData+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing image list: %w", err)
	}

	if err = os.WriteFile(indexFile, append(b, []byte("\n")...), 0o644); err != nil {
		return fmt.Errorf("writing formula index: %w", err)
	}

	fmt.Println("Image list:    " + o.StyleBold(imagesFile))
	fmt.Println("Formula index: " + o.StyleBold(indexFile))

	return nil
}

func (action *Images) findDeps(ctx context.Context, formulae ...string) ([]*formula.Formula, error) {
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
	deps, err := dependencies.Walk(ctx,
		apiwalker.New(index, platform.All),
		all,
		&action.DependencyOptions)
	// deps, err := formula.WalkDependencies(index, &action.DependencyOptions, all...)
	if err != nil {
		return nil, err
	}

	dependents := deps.Dependents()
	fmt.Printf("Found %d dependencies\n", len(dependents))

	// Combine root formulae with their dependencies in this list
	all = append(all, dependents...)
	return all, nil
}

// listImages lists the images for each formula in the index
func (action *Images) listImages(ctx context.Context, formulae []*formula.Formula) ([]string, error) {
	// Print no-resolve warning
	if action.NoResolve || action.NoVerify {
		o.Poo("Skipping tag resolution")
	}

	// Print no-verify warning
	if action.NoVerify {
		o.Poo("Skipping tag verification")
	}

	reg, err := action.Registry()
	if err != nil {
		return nil, err
	}

	mapper := iter.Mapper[*formula.Formula, string]{MaxGoroutines: action.MaxGoroutines()}
	return mapper.MapErr(formulae, func(f **formula.Formula) (string, error) {
		return action.resolve(ctx, reg, *f)
	})
}

func (action *Images) resolve(ctx context.Context, reg bottle.Registry, f *formula.Formula) (string, error) {
	m, err := bottle.Manifest(f, v1.Stable)
	if err != nil {
		return "", fmt.Errorf("computing bottle manifest: %w", err)
	}

	image := path.Join(action.Config().Registry.Prefix, m)

	if action.NoVerify {
		// Skip any resolving or verifying
		return image, nil
	}

	// create client for bottle repository
	repo, err := reg.Repository(ctx, f.Name)
	if err != nil {
		return image, err
	}

	o.Hai("Verifying " + image)

	// verify the bottle's tag by resolving it to a descriptor
	desc, err := repo.Resolve(ctx, image)
	if err != nil {
		return "", fmt.Errorf("verifying bottle tag: %w", err)
	}

	if action.NoResolve {
		// Add the image without the appending digest
		return image, nil
	}

	fmt.Println("Resolved digest " + string(desc.Digest))
	return image + "@" + string(desc.Digest), nil
}
