package actions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sourcegraph/conc/iter"
	"oras.land/oras-go/v2/registry/remote/retry"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/dependencies"
	apiwalker "github.com/act3-ai/hops/internal/dependencies/api"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/pretty"
)

// Install represents the action and its options
type Install struct {
	*Hops

	DependencyOptions dependencies.Options

	platform platform.Platform // store target platform

	// Install formulae without checking for previously installed keg-only
	// or non-migrated versions.
	Force bool

	// Show what would be installed, but do not actually install anything.
	DryRun bool

	// An unsupported Homebrew development option to skip installing any
	// dependencies of any kind. If the dependencies are not already
	// present, the formula will have issues. If you're not developing
	// Homebrew, consider adjusting your PATH rather than using this option.
	IgnoreDependencies bool

	// Install the dependencies with specified options but do not install
	// the formula itself
	OnlyDependencies bool

	// Print install times for each package at the end of the run
	// DisplayTimes bool

	// Delete files that already exist in the prefix while linking
	Overwrite bool
}

// Run runs the action
func (action *Install) Run(ctx context.Context, names ...string) error {
	brew := action.Homebrew()
	action.platform = platform.SystemPlatform()

	installs, err := action.resolveInstalls(ctx, names)
	if err != nil {
		return err
	}

	// Exit here for dry run
	if action.DryRun {
		return nil
	}

	if len(installs) == 0 {
		return nil
	}

	store := bottle.NewIndexStore(
		action.AuthHeaders(),
		retry.DefaultClient,
		action.AuthClient(),
		brew.Cache)

	// Use an iterator to start concurrent installs of each bottle
	m := iter.Mapper[*formula.Formula, string]{MaxGoroutines: action.MaxGoroutines()}
	results, err := m.MapErr(installs,
		func(f **formula.Formula) (string, error) {
			return action.run(ctx, store, *f)
		})
	if err != nil {
		return err
	}

	kegs := []string{}
	for _, keg := range results {
		if keg != "" {
			kegs = append(kegs, keg)
		}
	}

	// Print stats on the keg's contents
	o.Hai(fmt.Sprintf("Installed %d formulae in the Cellar:", len(kegs)))
	pretty.InstallStats(kegs)

	// todo: fix so it does not print caveats of already up-to-date formulae
	// 3. Finish by printing all caveats again
	for _, f := range installs {
		if caveats := pretty.Caveats(f, action.Prefix()); caveats != "" {
			o.Hai(f.Name + ": Caveats\n" + caveats)
		}
	}

	return nil
}

// resolveInstalls resolves the list of formulae that will be installed
func (action *Install) resolveInstalls(ctx context.Context, names []string) ([]*formula.Formula, error) {
	index := action.Index()
	err := formula.AutoUpdate(ctx, index, &action.Config().Homebrew.AutoUpdate)
	if err != nil {
		return nil, err
	}

	// Resolve direct formulae
	roots, err := action.FetchAll(o.H1, index, names...)
	if err != nil {
		return nil, err
	}

	roots, skipped, err := filterUninstalled(ctx, roots, formulaInstalled(action.Prefix()))
	if err != nil {
		return nil, err
	}
	for _, f := range skipped {
		o.Poo(reinstallHelper(f.Name, f.Version()))
	}

	// Ignore dependencies
	if action.IgnoreDependencies {
		return roots, nil
	}

	graph, err := dependencies.Walk(ctx,
		apiwalker.New(index, action.platform),
		roots,
		&action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	printFormulae(formula.Names(roots), action.DryRun)

	dependents := graph.Dependents()
	printDeps(formula.Names(dependents), action.DryRun, action.IgnoreDependencies)

	// Only install dependencies
	if action.OnlyDependencies {
		return dependents, nil
	}

	// Install dependencies and then requested formulae
	return slices.Concat(dependents, roots), nil
}

// run is the meat
func (action *Install) run(ctx context.Context, store *bottle.IndexStore, f *formula.Formula) (string, error) {
	outdated, err := action.Prefix().FormulaOutdated(f)
	if err != nil {
		return "", err
	}

	if !outdated {
		o.Hai(fmt.Sprintf("%s %s is already installed and up-to-date.", f.FullName, f.Version()))
		return "", nil
	}

	b, err := bottle.FromFormula(f, v1.Stable, action.platform)
	if err != nil {
		return "", err
	}

	if !b.CompatibleWithCellar(action.Prefix().Cellar()) {
		// incompatibleCellarWarning(bottles[i], brew.Cellar())
		slog.Warn(newIncompatibleCellarError(b.Name, b.File.Cellar, action.Prefix().Cellar()).Error())
	}
	slog.Info("Downloading " + store.Source(b))

	// 1. Download bottle to the cache
	err = store.Download(ctx, b)
	if err != nil {
		return "", err
	}

	// 2. Pour bottle to the Cellar
	slog.Info("Pouring " + b.ArchiveName()) // ex: Pouring cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
	err = bottle.PourFile(store.Path(b), b, action.Prefix().Cellar())
	if err != nil {
		return "", err
	}

	// 3. Link keg to the prefix
	if !f.KegOnly || action.Force {
		slog.Info("Linking " + b.Name) // ex: Linking cowsay

		lnopts := &prefix.LinkOptions{
			Name:      b.Name,
			Overwrite: action.Overwrite,
			DryRun:    action.DryRun,
		}

		_, _, err = action.Prefix().Link(f.Name, f.Version(), lnopts)
		if err != nil {
			return "", err
		}
	}

	return action.Prefix().KegPath(f.Name, f.Version()), nil
}

func printFormulae(roots []string, dryrun bool) {
	fword := "formulae"
	flist := wordwrap.String(strings.Join(roots, " "), o.Width)
	if len(roots) == 1 {
		fword = "formula"
	}

	switch {
	// No formulae
	case len(roots) == 0:
	// Print formulae
	case dryrun:
		o.Hai(fmt.Sprintf("Would install %d %s:\n%s", len(roots), fword, flist))
	// Print formulae
	default:
		o.H1(fmt.Sprintf("Installing %d %s:\n%s", len(roots), fword, flist))
	}
}

func printDeps(deps []string, dryrun, ignoredeps bool) {
	dword := "dependencies"
	dlist := wordwrap.String(strings.Join(deps, " "), o.Width)
	if len(deps) == 1 {
		dword = "dependency"
	}

	switch {
	// No deps
	case len(deps) == 0:
	// Print ignored deps
	case dryrun && ignoredeps:
		o.Hai(fmt.Sprintf("Would install %d %s:\n%s", len(deps), dword, dlist))
	// Warn ignored deps
	case ignoredeps:
		o.Poo(fmt.Sprintf("Ignoring %d %s", len(deps), dword))
	// Print deps
	case dryrun:
		o.Hai(fmt.Sprintf("Would install %d %s:\n%s", len(deps), dword, dlist))
	// Print deps
	default:
		o.H1(fmt.Sprintf("Installing %d %s:\n%s", len(deps), dword, dlist))
	}
}

func newIncompatibleCellarError(name, wantCellar, cellar string) error {
	return errors.New(heredoc.Docf(`
		bottle for %s may be incompatible with your settings
		  HOMEBREW_CELLAR: %s (yours is %s)
		  HOMEBREW_PREFIX: %s (yours is %s)`,
		name,
		wantCellar, cellar,
		filepath.Dir(wantCellar), filepath.Dir(cellar),
	))
}
