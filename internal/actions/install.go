package actions

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sourcegraph/conc/iter"

	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/errdef"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/bottle"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/pretty"
	"github.com/act3-ai/hops/internal/utils/iterutil"
)

// Install represents the action and its options.
type Install struct {
	*Hops

	DependencyOptions formula.DependencyTags

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

// Run runs the action.
func (action *Install) Run(ctx context.Context, args ...string) error {
	action.platform = platform.SystemPlatform()
	names := action.SetAlternateTags(args)

	err := action.autoUpdate(ctx)
	if err != nil {
		return err
	}

	installs, err := action.resolveInstalls(ctx, names)
	if err != nil {
		return err
	}

	// Exit here for dry run
	if action.DryRun {
		return nil
	}

	// Exit if there is nothing to install
	if len(installs) == 0 {
		return nil
	}

	// Verify that all bottles can be poured
	err = action.Prefix().CanPourBottles(ctx, installs)
	if err != nil {
		return err
	}

	// Get bottle registry
	reg, err := action.BottleRegistry(ctx)
	if err != nil {
		return err
	}

	// Download all bottles
	bottles, err := bottle.FetchAll(ctx, reg, installs)
	if err != nil {
		return err
	}

	// Install the downloaded bottles
	routines := iter.Iterator[formula.PlatformFormula]{MaxGoroutines: action.MaxGoroutines()}
	err = iterutil.ForEachIdxErr(routines, installs, func(i int, pf *formula.PlatformFormula) error {
		// var err error // create local err variable to avoid race
		btl := bottles[i]
		return errors.Join(
			action.run(ctx, *pf, btl),
			btl.Close(),
		)
	})
	if err != nil {
		return err
	}

	// Print stats on the keg's contents
	o.Hai(fmt.Sprintf("Installed %d formulae in the Cellar:", len(installs)))
	pretty.FormulaInstallStats(action.Prefix(), installs)

	// Finish by printing all caveats again
	for _, f := range installs {
		if caveats := pretty.Caveats(f, action.Prefix()); caveats != "" {
			o.Hai(f.Name() + ": Caveats\n" + caveats)
		}
	}

	return nil
}

// resolveInstalls resolves the list of formulae that will be installed.
func (action *Install) resolveInstalls(ctx context.Context, names []string) ([]formula.PlatformFormula, error) {
	formulary, err := action.Formulary(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch directly-requested formulae
	all, err := formula.FetchAllPlatform(ctx, formulary, names, action.platform)
	if err != nil {
		return nil, err
	}

	// Filter requested formulae into installed and not installed lists
	roots, reinstalls, err := prefix.FilterInstalled(action.Prefix(), all)
	if err != nil {
		return nil, err
	}

	// Direct user to the reinstall command
	for _, f := range reinstalls {
		version := formula.PkgVersion(f)
		o.Poo(heredoc.Docf(`
			%s
			To reinstall %s, run:
			  hops reinstall %s`,
			errdef.NewErrFormulaUpToDate(f.Name(), version).Error(),
			version,
			f.Name()))
	}

	// Exit with no error for reinstalls
	if len(reinstalls) > 0 {
		return nil, nil
	}

	// Ignore dependencies
	// --ignore-dependencies flag
	if action.IgnoreDependencies {
		return roots, nil
	}

	// Build dependency graph
	graph, err := dependencies.Walk(ctx, formulary, roots, action.platform, &action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	// Print all resolved dependencies
	printFormulae(formula.Names(roots), action.DryRun)

	// Filter dependencies into installed and not installed lists
	allDeps := graph.Dependencies()
	slog.Debug("Resolved dependencies",
		slog.Any("dependencies", formula.Names(allDeps)),
		action.DependencyOptions.LogAttr(),
	)
	missingDeps, installedDeps, err := prefix.FilterInstalled(action.Prefix(), allDeps)
	if err != nil {
		return nil, err
	}
	slog.Debug("Validated dependencies", slog.Any("installed", formula.Names(installedDeps)), slog.Any("missing", formula.Names(missingDeps)))

	printDeps(formula.Names(missingDeps), action.DryRun, action.IgnoreDependencies)

	// Only install dependencies
	if action.OnlyDependencies {
		return missingDeps, nil
	}

	// Install dependencies and then requested formulae
	return slices.Concat(missingDeps, graph.Roots()), nil
}

// run is the meat.
func (action *Install) run(_ context.Context, f formula.PlatformFormula, btl io.Reader) error {
	l := slog.Default().With(slog.String("formula", f.Name()))

	// 2. Pour bottle to the Cellar
	l.Info("Pouring bottle", slog.String("file", formula.BottleFileName(f)))
	// slog.Info("Pouring " + b.ArchiveName()) // ex: Pouring cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
	err := action.Prefix().Pour(btl)
	if err != nil {
		return err
	}

	// 3. Link keg to the prefix
	if !f.IsKegOnly() || action.Force {
		l.Info("Linking keg", slog.String("keg", action.Prefix().FormulaKegPath(f))) // ex: Linking cowsay

		lnopts := &prefix.LinkOptions{
			Name:      f.Name(),
			Overwrite: action.Overwrite,
			DryRun:    action.DryRun,
		}

		_, _, err = action.Prefix().FormulaLink(f, lnopts)
		if err != nil {
			return err
		}
	}

	return nil
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
