package actions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
)

// Upgrade represents the action and its options
type Upgrade struct {
	*Hops
}

// Run runs the action
func (action *Upgrade) Run(ctx context.Context, names ...string) error {
	index := action.Index()
	err := formula.AutoUpdate(ctx, index, &action.Config().Homebrew.AutoUpdate)
	if err != nil {
		return err
	}

	// TODO: if no args are passed, run upgrade for all installed formulae
	// if len(formulae) == 0 {
	// 	formulae = brew.List()
	// }

	formulae, err := action.FetchAll(o.H1, index, names...)
	if err != nil {
		return err
	}

	for _, f := range formulae {
		kegs, err := action.Prefix().InstalledKegs(f)
		if err != nil {
			return err
		}

		outOfDate := true
		for _, k := range kegs {
			kegVersion := k.Version()
			switch versionCompare(f.Version(), kegVersion) {
			case 0:
				outOfDate = false
			case 1:
				slog.Debug("out of date keg found", slog.String("version", kegVersion))
			default:
				slog.Debug("keg has newer version", slog.String("version", kegVersion))
			}
		}

		if !outOfDate {
			o.Poo(fmt.Sprintf("%s %s already installed", f.Name, f.Version()))
			continue
		}

		fmt.Println("Upgrading " + f.Version())
		// TODO do install
	}

	return nil
}
