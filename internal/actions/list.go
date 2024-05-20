package actions

import (
	"context"
	"fmt"
	"strings"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix/keg"
)

// List represents the action and its options
type List struct {
	*Hops

	// // Print formulae with fully-qualified names.
	// // Unless --full-name, --versions or
	// // --pinned are passed, other options (i.e.
	// // -1, -l, -r and -t) are passed to
	// // ls(1) which produces the actual output
	// FullName bool

	// Show the version number for installed
	// formulae, or only the specified formulae if
	// formula are provided
	Versions bool

	// Only show formulae with multiple versions
	// installed
	Multiple bool

	// // List only pinned formulae, or only the
	// // specified (pinned) formulae if formula are
	// // provided. See also pin, unpin
	// Pinned bool

	// // Force output to be one entry per line. This
	// // is the default when output is not to a
	// // terminal
	// OnePerLine bool

	// // List formulae and/or casks in long format.
	// // Has no effect when a formula or cask name is
	// // passed as an argument
	// Long bool

	// // Reverse the order of the formulae and/or
	// // casks sort to list the oldest entries first.
	// // Has no effect when a formula or cask name is
	// // passed as an argument
	// Reverse bool

	// // Sort formulae and/or casks by time modified,
	// // listing most recently modified first. Has no
	// // effect when a formula or cask name is passed
	// // as an argument
	// TimeSort bool
}

// Run runs the action
func (action *List) Run(ctx context.Context, names ...string) error {
	switch {
	case len(names) > 0:
		return action.names(ctx, names)
	case action.Multiple:
		kegs, err := action.Prefix().Kegs()
		if err != nil {
			return err
		}
		multiple(kegs)
	default:
		racks, err := action.Prefix().Racks()
		if err != nil {
			return err
		}

		o.Hai("Installed:")
		for _, r := range racks {
			fmt.Println(r.Name())
		}
	}

	return nil
}

func multiple(kegs []keg.Keg) {
	count := map[string]int{}
	for _, k := range kegs {
		count[k.Name()]++
	}

	for _, k := range kegs {
		if count[k.Name()] > 1 {
			fmt.Println(k.Name() + " " + k.Version())
		}
	}
}

func (action *Hops) resolveNames(ctx context.Context, log func(string), names ...string) ([]*brewv1.Info, error) {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve direct formulae
	return action.FetchAll(log, index, names...)
}

func (action *List) names(ctx context.Context, names []string) error {
	formulae, err := action.resolveNames(ctx, o.Noop, names...)
	if err != nil {
		return err
	}

	for _, f := range formulae {
		kegs, err := action.Prefix().InstalledKegs(f)
		if err != nil {
			return err
		}

		switch {
		case action.Multiple:
			// Only output this info if there are multiple kegs
			if len(kegs) > 1 {
				multiple(kegs)
			}
		case action.Versions:
			for _, k := range kegs {
				fmt.Println(k.Name() + " " + k.Version())
			}
		default:
			for _, k := range kegs {
				kpaths, err := k.Paths()
				if err != nil {
					return err
				}
				fmt.Println(strings.Join(kpaths, "\n"))
			}
		}
	}

	return nil
}
