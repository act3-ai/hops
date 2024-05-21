package actions

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/sourcegraph/conc/iter"

	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/brewfile"
	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/formula"
	hopsreg "github.com/act3-ai/hops/internal/hops/registry"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
)

// Images represents the action and its options
type Images struct {
	*Hops
	DependencyOptions formula.DependencyTags

	File      string // path to a Brewfile specifying formulae dependencies
	NoResolve bool   // disable tag resolution
	NoVerify  bool   // disable tag verification
}

// Run runs the action
func (action *Images) Run(ctx context.Context, args ...string) error {
	// Add Brewfile dependencies if requested
	if action.File != "" {
		bf, err := brewfile.Load(action.File)
		if err != nil {
			return err
		}
		args = append(args, bf.Formula...)
	}

	slog.Debug("finding images for", slog.Any("formulae", args))

	formulae, err := action.findDeps(ctx, args)
	if err != nil {
		return err
	}

	images, err := action.listImages(ctx, formulae)
	if err != nil {
		return err
	}

	imageData := strings.Join(images, "\n")
	o.Hai("Images:\n" + imageData)
	// b, err := json.Marshal(formulae)
	// if err != nil {
	// 	return err
	// }

	o.Hai("Writing image list and formula index")

	imagesFile := "hops.images.txt"
	// indexFile := "hops.index.json"

	if err = os.WriteFile(imagesFile, []byte(imageData+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing image list: %w", err)
	}

	// if err = os.WriteFile(indexFile, append(b, []byte("\n")...), 0o644); err != nil {
	// 	return fmt.Errorf("writing formula index: %w", err)
	// }

	fmt.Println("Image list:    " + o.StyleBold(imagesFile))
	// fmt.Println("Formula index: " + o.StyleBold(indexFile))

	return nil
}

func (action *Images) findDeps(ctx context.Context, args []string) ([]formula.PlatformFormula, error) {
	formulary, err := action.FormulaClient(ctx, args)
	if err != nil {
		return nil, err
	}

	names, _ := parseArgs(args)
	roots, err := formula.FetchAllPlatform(ctx, formulary, names, platform.All)
	if err != nil {
		return nil, err
	}

	o.H1("Fetching dependencies...")

	// Build dependency graph
	graph, err := dependencies.WalkAll(ctx, formulary, roots, &action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	dependents := graph.Dependents()
	fmt.Printf("Found %d dependencies\n", len(dependents))

	// Combine root formulae with their dependencies in this list
	return slices.Concat(dependents, roots), nil
}

// listImages lists the images for each formula in the index
func (action *Images) listImages(ctx context.Context, formulae []formula.PlatformFormula) ([]string, error) {
	// Print no-resolve warning
	if action.NoResolve || action.NoVerify {
		o.Poo("Skipping tag resolution")
	}

	// Print no-verify warning
	if action.NoVerify {
		o.Poo("Skipping tag verification")
	}

	reg, err := action.registry()
	if err != nil {
		return nil, err
	}

	mapper := iter.Mapper[formula.PlatformFormula, string]{MaxGoroutines: action.MaxGoroutines()}
	return mapper.MapErr(formulae, func(f *formula.PlatformFormula) (string, error) {
		return action.resolve(ctx, reg, *f)
	})
}

func (action *Images) resolve(ctx context.Context, reg hopsreg.Registry, f formula.Formula) (string, error) {
	image := strings.TrimSuffix(action.Config().Registry.Prefix, "/") + "/" + brewfmt.Repo(f.Name()) + ":" + formula.Tag(f.Version())

	if action.NoVerify {
		// Skip any resolving or verifying
		return image, nil
	}

	// create client for bottle repository
	repo, err := reg.Repository(ctx, f.Name())
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
