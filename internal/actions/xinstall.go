package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"oras.land/oras-go/v2/errdef"

	"github.com/MakeNowJust/heredoc/v2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sourcegraph/conc/iter"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/bottle/regbottle"
	"github.com/act3-ai/hops/internal/brew"
	"github.com/act3-ai/hops/internal/dependencies"
	regwalker "github.com/act3-ai/hops/internal/dependencies/registry"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/pretty"
)

/*
Workflow:
1. Resolve formula name to manifest descriptor
2. Fetch metadata artifact referring to the manifest
3. Check if formula is installed:
	- If not already installed, add to install list
4. Resolve all dependencies using same workflow as above
5. Install each formula
*/

// XInstall represents the action and its options
type XInstall Install

// Run runs the action
func (action *XInstall) Run(ctx context.Context, names ...string) error {
	_ = action.Homebrew()   // preload
	_ = action.AuthClient() // preload
	action.platform = platform.SystemPlatform()

	// Grab the configured bottle registry
	srcReg, err := action.Registry()
	if err != nil {
		return err
	}

	// Initialize OCI layout cache
	cacheReg := bottle.NewLocal(filepath.Join(action.Hops.Config().Cache, "oci"))

	roots, err := action.resolveRoots(ctx, srcReg, cacheReg, names)
	if err != nil {
		return err
	}

	// Cancel if all roots are installed
	if len(roots) == 0 {
		return nil
	}

	installs, err := action.evaluateInstalls(ctx, srcReg, cacheReg, roots)
	if err != nil {
		return err
	}

	// Maps from descriptors of bottles already downloaded to the kegs they were poured into
	mapper := iter.Mapper[*regbottle.BottleIndex, *v1.Info]{MaxGoroutines: action.Hops.MaxGoroutines()}

	formulae, err := mapper.MapErr(installs,
		func(bp **regbottle.BottleIndex) (*v1.Info, error) {
			b := *bp

			cache, err := cacheReg.Repository(ctx, b.RepositoryName)
			if err != nil {
				return nil, err
			}

			return action.run(ctx, cache, b)
		})
	if err != nil {
		return err
	}

	kegs := []string{}
	installed := []*v1.Info{}
	for _, f := range formulae {
		if f != nil {
			kegs = append(kegs, action.Prefix().KegPath(f.Name, f.Version()))
			installed = append(installed, f)
		}
	}

	// Print stats on the keg's contents
	o.Hai(fmt.Sprintf("Installed %d formulae in the Cellar:", len(installed)))
	pretty.InstallStats(kegs)

	// 3. Finish by printing all caveats again
	for _, b := range installed {
		if caveats := pretty.Caveats(b, action.Prefix()); caveats != "" {
			o.Hai(b.Name + ": Caveats\n" + caveats)
		}
	}

	return nil
}

func (action *XInstall) resolveRoots(ctx context.Context, srcReg bottle.Registry, cache bottle.Registry, args []string) ([]*regbottle.BottleIndex, error) {
	fetchers := iter.Mapper[string, *regbottle.BottleIndex]{
		MaxGoroutines: action.MaxGoroutines(),
	}

	roots, err := fetchers.MapErr(args, func(s *string) (*regbottle.BottleIndex, error) {
		arg := *s

		name, version := parseArg(arg)
		if version == "" {
			version = "latest"
		}

		btl, err := fetchBottle(ctx,
			srcReg, cache,
			name, version, action.platform)
		if err != nil {
			return nil, fmt.Errorf("fetching bottle %s %s: %w", name, version, err)
		}

		return btl, nil
	})
	if err != nil {
		return nil, err
	}

	roots, skipped, err := filterUninstalled(ctx, roots, regbottleInstalled(action.Prefix(), action.platform, cache))
	if err != nil {
		return nil, err
	}

	if err := printReinstallHelper(ctx, action.platform, cache, skipped); err != nil {
		return nil, err
	}

	return roots, nil
}

func filterUninstalled[T any](ctx context.Context, list []*T, isInstalledFunc func(ctx context.Context, entry *T) (bool, error)) (uninstalled []*T, installed []*T, err error) {
	for _, entry := range list {
		isInstalled, err := isInstalledFunc(ctx, entry)
		if err != nil {
			return nil, nil, err
		}

		// Categorize formulae based on install status
		if isInstalled {
			installed = append(installed, entry)
		} else {
			uninstalled = append(uninstalled, entry)
		}
	}

	return uninstalled, installed, nil
}

func regbottleInstalled(p prefix.Prefix, plat platform.Platform, cache bottle.Registry) func(ctx context.Context, btl *regbottle.BottleIndex) (bool, error) {
	return func(ctx context.Context, btl *regbottle.BottleIndex) (bool, error) {
		repo, err := cache.Repository(ctx, btl.RepositoryName)
		if err != nil {
			return false, err
		}
		info, err := btl.PlatformMetadata(ctx, repo, plat)
		if err != nil {
			return false, err
		}
		return formulaInstalled(p)(ctx, &v1.Info{PlatformInfo: *info})
	}
}

func formulaInstalled(p prefix.Prefix) func(ctx context.Context, f *v1.Info) (bool, error) {
	return func(_ context.Context, f *v1.Info) (bool, error) {
		notInstalled, err := p.FormulaOutdated(f)
		if err != nil {
			return false, err
		}
		return !notInstalled, nil
	}
}

func printReinstallHelper(ctx context.Context, plat platform.Platform, cache bottle.Registry, installedBottles []*regbottle.BottleIndex) error {
	for _, btl := range installedBottles {
		repo, err := cache.Repository(ctx, btl.RepositoryName)
		if err != nil {
			return err
		}

		f, err := btl.PlatformMetadata(ctx, repo, plat)
		if err != nil {
			return err
		}

		o.Poo(reinstallHelper(f.Name, f.Version()))
	}
	return nil
}

func reinstallHelper(name, version string) string {
	return heredoc.Docf(`
		%s
		To reinstall %s, run:
		  hops reinstall %s`,
		brew.NewErrFormulaUpToDate(name, version).Error(),
		version,
		name)
}

// type InstallDependencyOptions struct {
// 	Dependencies       dependencies.Options
// 	IgnoreDependencies bool
// 	OnlyDependencies   bool
// }

func (action *XInstall) evaluateInstalls(ctx context.Context, srcReg bottle.Registry, cache bottle.Registry, roots []*regbottle.BottleIndex) ([]*regbottle.BottleIndex, error) {
	// --ignore-dependencies flag
	if action.IgnoreDependencies {
		return roots, nil
	}

	depwalker := regwalker.New(srcReg, cache, action.platform, action.MaxGoroutines())

	walkedDeps, err := dependencies.Walk(ctx, depwalker, roots, &action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	missingDependents, _, err := filterUninstalled(ctx, walkedDeps.Dependents(), regbottleInstalled(action.Prefix(), action.platform, cache))
	if err != nil {
		return nil, err
	}

	if action.OnlyDependencies {
		return missingDependents, nil
	}

	return slices.Concat(missingDependents, roots), nil
}

func fetchBottle(ctx context.Context, srcReg bottle.Registry, dstReg bottle.Registry, name, version string, plat platform.Platform) (*regbottle.BottleIndex, error) {
	src, err := srcReg.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	dst, err := dstReg.Repository(ctx, name)
	if err != nil {
		return nil, err
	}

	btl, err := regbottle.ResolveVersion(ctx, src, version)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, errors.Join(err, listAvailableTags(ctx, src, name))
		}
		return nil, err
	}

	if err := regbottle.CopyTargetPlatform(ctx, src, dst, btl, plat); err != nil {
		return nil, err
	}

	return btl, nil
}

func (action *XInstall) run(ctx context.Context, store bottle.Repository, b *regbottle.BottleIndex) (*v1.Info, error) {
	info, err := b.PlatformMetadata(ctx, store, action.platform)
	if err != nil {
		return nil, err
	}

	f := &v1.Info{PlatformInfo: *info}
	outdated, err := action.Prefix().FormulaOutdated(f)
	if err != nil {
		return nil, err
	}

	if !outdated {
		o.Hai(fmt.Sprintf("%s %s is already installed and up-to-date.", info.FullName, f.Version()))
		return nil, nil
	}

	btl, err := bottle.FromFormula(f, v1.Stable, action.platform)
	if err != nil {
		return nil, err
	}

	if !btl.CompatibleWithCellar(action.Prefix().Cellar()) {
		slog.Warn(newIncompatibleCellarError(btl.Name, btl.File.Cellar, action.Prefix().Cellar()).Error())
	}

	bottleDesc, err := b.ResolveBottle(ctx, store, action.platform)
	if err != nil {
		return nil, err
	}

	bottleFile := bottleDesc.Digest.Encoded()
	if v, ok := bottleDesc.Annotations[ocispec.AnnotationTitle]; ok {
		bottleFile = v
	}

	slog.Info("Pouring bottle", slog.String("file", bottleFile))

	// The path the bottle will be unzipped to
	kegPath := filepath.Join(action.Prefix().Cellar(), f.Name)
	if err := os.RemoveAll(kegPath); err != nil {
		return nil, fmt.Errorf("removing old keg: %w", err)
	}

	r, err := store.Fetch(ctx, bottleDesc)
	if err != nil {
		return nil, fmt.Errorf("fetching bottle: %w", err)
	}
	defer r.Close()

	err = bottle.PourReader(r, action.Prefix().Cellar())
	if err != nil {
		return nil, fmt.Errorf("pouring bottle %s: %w", bottleFile, err)
	}

	// 3. Link keg to the prefix
	if !f.KegOnly || action.Force {
		slog.Info("Linking " + f.Name) // ex: Linking cowsay

		lnopts := &prefix.LinkOptions{
			Name:      f.Name,
			Overwrite: action.Overwrite,
			DryRun:    action.DryRun,
		}

		_, _, err = action.Prefix().Link(f.Name, f.Version(), lnopts)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func parseArg(arg string) (name, version string) {
	fields := strings.SplitN(arg, "=", 2)
	if len(fields) == 1 {
		return fields[0], ""
	}
	return fields[0], fields[1]
}

func listAvailableTags(ctx context.Context, repo bottle.Repository, name string) error {
	tags, err := bottle.ListTags(ctx, repo)
	if err != nil {
		o.Poo(fmt.Sprintf("[%s] Could not list available tags", name))
		return err
	}

	o.Hai(fmt.Sprintf("[%s] Found %d tags", name, len(tags)))
	if len(tags) > 0 {
		slices.Reverse(tags)
		if len(tags) > 10 {
			fmt.Println("Available tags:\n\t" + strings.Join(tags[:11], "\n\t") + "\n\t…")
			// fmt.Println("Available tags:\n\t" + strings.Join(tags[:11], "\n\t") + "\n\t...")
		} else {
			fmt.Println("Available tags:\n\t" + strings.Join(tags, "\n\t"))
		}
	}

	return nil
}
