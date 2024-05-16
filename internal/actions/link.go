package actions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix"
)

// Link represents the action and its options
type Link struct {
	*Hops

	Overwrite bool // Delete files that already exist in the prefix while linking
	DryRun    bool // List files which would be linked or deleted by hops link --overwrite without actually linking or deleting any files
	Force     bool // Allow keg-only formulae to be linked
	Head      bool // Link the HEAD version of the formula if it is installed
}

// Run runs the action
func (action *Link) Run(ctx context.Context, names ...string) error {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return err
	}

	formulae, err := action.FetchAll(o.H1, index, names...)
	if err != nil {
		return err
	}

	// Confirm no keg-only formulae were requested without --force
	for _, f := range formulae {
		if !action.Prefix().AnyInstalled(f) {
			return action.Prefix().NewErrNoSuchKeg(f.Name)
		}
		if f.KegOnly && !action.Force {
			return fmt.Errorf("%s is keg-only and must be linked with %s", f.Name, o.StyleBold("--force"))
		}
	}

	for _, f := range formulae {
		if action.DryRun {
			o.Hai(fmt.Sprintf("Would create the following links for %s:", f.FullName))
		} else {
			o.Hai("Linking " + f.FullName) // ex: Linking cowsay
		}
		links, _, err := action.Prefix().Link(f.Name, f.Version(), &prefix.LinkOptions{
			Name:      f.Name,
			Overwrite: action.Overwrite,
			DryRun:    action.DryRun,
		})
		if err != nil {
			return err
		}

		if !action.DryRun {
			fmt.Printf("Linking %s... %d symlinks created.\n", action.Prefix().KegPath(f.Name, f.Version()), links)
		}
	}

	return nil
}

// Unlink represents the action and its options
type Unlink struct {
	*Hops

	DryRun bool // List files which would be unlinked without actually unlinking or deleting any files
}

// Run runs the action
func (action *Unlink) Run(ctx context.Context, names ...string) error {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return err
	}

	formulae, err := action.FetchAll(o.H1, index, names...)
	if err != nil {
		return err
	}

	// List all installed kegs
	// TODO: make sure the formula is actually linked by checking the homebrew/var/link dir first
	kegs := make([]string, 0, len(names))
	for _, f := range formulae {
		fkegs, err := action.Prefix().InstalledKegs(f)
		if err != nil {
			return err
		}
		if len(fkegs) == 0 {
			return action.Prefix().NewErrNoSuchKeg(f.Name)
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
