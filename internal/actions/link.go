package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
)

// Link represents the action and its options.
type Link struct {
	*Hops

	platform platform.Platform

	Overwrite bool // Delete files that already exist in the prefix while linking
	DryRun    bool // List files which would be linked or deleted by hops link --overwrite without actually linking or deleting any files
	Force     bool // Allow keg-only formulae to be linked
	Head      bool // Link the HEAD version of the formula if it is installed
}

// Run runs the action.
func (action *Link) Run(ctx context.Context, args []string) error {
	if action.platform == "" {
		action.platform = platform.SystemPlatform()
	}

	names := action.SetAlternateTags(args)

	store, err := action.Formulary(ctx)
	if err != nil {
		return err
	}

	formulae, err := formula.FetchAllPlatform(ctx, store, names, action.platform)
	if err != nil {
		return err
	}

	// Confirm no keg-only formulae were requested without --force
	for _, f := range formulae {
		if !action.Prefix().AnyInstalled(f) {
			return action.Prefix().NewErrNoSuchKeg(f.Name())
		}
		if f.IsKegOnly() && !action.Force {
			return errors.New(f.Name() + " is keg-only and must be linked with " + o.StyleBold("--force"))
		}
	}

	for _, f := range formulae {
		if action.DryRun {
			o.Hai(fmt.Sprintf("Would create the following links for %s:", f.Name()))
		} else {
			o.Hai("Linking " + f.Name()) // ex: Linking cowsay
		}
		links, _, err := action.Prefix().FormulaLink(f, &prefix.LinkOptions{
			Name:      f.Name(),
			Overwrite: action.Overwrite,
			DryRun:    action.DryRun,
		})
		if err != nil {
			return err
		}

		if !action.DryRun {
			fmt.Printf("Linking %s... %d symlinks created.\n", action.Prefix().FormulaKegPath(f), links)
		}
	}

	return nil
}

// Unlink represents the action and its options.
type Unlink struct {
	*Hops

	DryRun bool // List files which would be unlinked without actually unlinking or deleting any files
}

// Run runs the action.
func (action *Unlink) Run(ctx context.Context, args ...string) error {
	formulae, err := action.fetchFromArgs(ctx, args, platform.SystemPlatform())
	if err != nil {
		return err
	}

	// List all installed kegs
	// TODO: make sure the formula is actually linked by checking the homebrew/var/link dir first
	kegs := make([]string, 0, len(formulae))
	for _, f := range formulae {
		fkegs, err := action.Prefix().InstalledKegs(f)
		if err != nil {
			return err
		}
		if len(fkegs) == 0 {
			return action.Prefix().NewErrNoSuchKeg(f.Name())
		}
		for _, k := range fkegs {
			kegs = append(kegs, k.String())
		}
	}

	if action.DryRun {
		o.Hai("Would unlink the following kegs:\n" + strings.Join(kegs, "\n"))
	} else {
		o.Hai("Unlinking the following kegs:\n" + strings.Join(kegs, "\n"))
	}

	links, err := action.Prefix().LinkedFiles(kegs...)
	if err != nil {
		return err
	}

	// Return here for dry run
	if action.DryRun {
		o.Hai("Would remove the following links:\n" + strings.Join(links, "\n"))
		return nil
	}

	for _, l := range links {
		err = os.Remove(l)
		if err != nil {
			return fmt.Errorf("removing link %s: %w", l, err)
		}
	}

	fmt.Printf("%d symlinks removed.\n", len(links))

	return nil
}

func (action *Hops) resolve(ctx context.Context, args []string, plat platform.Platform, tags *formula.DependencyTags) (*dependencies.DependencyGraph, error) {
	// Detect the platform if unset
	if plat == "" {
		plat = platform.SystemPlatform()
	}

	names := action.SetAlternateTags(args)

	formulary, err := action.Formulary(ctx)
	if err != nil {
		return nil, err
	}

	formulae, err := formula.FetchAllPlatform(ctx, formulary, names, plat)
	if err != nil {
		return nil, err
	}

	return dependencies.Walk(ctx, formulary, formulae, plat, tags)
}
