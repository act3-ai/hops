package actions

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
)

// Upgrade represents the action and its options.
type Upgrade struct {
	*Hops
}

// Run runs the action.
func (action *Upgrade) Run(ctx context.Context, args ...string) error {
	var toCheck []formula.PlatformFormula
	var err error
	switch {
	case len(args) > 0:
		toCheck, err = action.fetchFromArgs(ctx, args, platform.SystemPlatform())
		if err != nil {
			return err
		}
	default:
		kegs, err := action.Prefix().Kegs()
		if err != nil {
			return err
		}

		kegNames := formula.Names(kegs)
		// Sort and remove duplicates
		slices.Sort(kegNames)
		kegNames = slices.Compact(kegNames)

		toCheck, err = action.fetchFromArgs(ctx, kegNames, platform.SystemPlatform())
		if err != nil {
			return err
		}
	}

	toUpgrade := make([]formula.PlatformFormula, 0, len(toCheck))
	for _, f := range toCheck {
		kegs, err := action.Prefix().InstalledKegs(f)
		if err != nil {
			return err
		}

		outOfDate := true
		for _, k := range kegs {
			kegVersion := k.Version()
			updated := formula.PkgVersion(f.Version())
			switch versionCompare(updated, kegVersion) {
			case 0:
				outOfDate = false
			case 1:
				slog.Debug("out of date keg found", slog.String("version", kegVersion))
			default:
				slog.Debug("keg has newer version", slog.String("version", kegVersion))
			}
		}

		if !outOfDate {
			o.Poo(fmt.Sprintf("%s %s already installed", f.Name(), f.Version()))
			continue
		}

		fmt.Println("Upgrading " + formula.PkgVersion(f.Version()))
		toUpgrade = append(toUpgrade, f)
	}

	// TODO do install
	fmt.Println("Would upgrade the following: " + strings.Join(formula.Names(toUpgrade), ", "))

	return nil
}
