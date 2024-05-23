package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/act3-ai/hops/internal/platform"
)

// Uninstall represents the action and its options.
type Uninstall struct {
	*Hops
}

// Run runs the action.
func (action *Uninstall) Run(ctx context.Context, args []string) error {
	formulae, err := action.fetchFromArgs(ctx, args, platform.SystemPlatform())
	if err != nil {
		return err
	}

	// List all installed kegs
	kegs := make([]string, 0, len(args))
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

	// Uninstall the found kegs
	err = action.Prefix().Uninstall(kegs...)
	if err != nil {
		return err
	}

	for _, k := range kegs {
		kparent := filepath.Dir(k)
		err = os.Remove(kparent)
		if err != nil {
			return fmt.Errorf("removing prefix directory %s: %w", kparent, err)
		}
	}

	/*
		Add something like the following:
			Warning: The following may be dbus configuration files and have not been removed!
			If desired, remove them manually with `rm -rf`:
			  /opt/homebrew/etc/dbus-1
	*/

	return nil
}
