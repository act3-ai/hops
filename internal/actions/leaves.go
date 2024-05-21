package actions

import (
	"context"
	"fmt"
	"slices"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
)

// Leaves represents the action and its options
type Leaves struct {
	*Hops

	// InstalledOnRequest    bool // Only list leaves that were manually installed
	// InstalledAsDependency bool // Only list leaves that were installed as dependencies
}

// Run runs the action
func (action *Leaves) Run(ctx context.Context) error {
	// List all racks
	kegs, err := action.Prefix().Kegs()
	if err != nil {
		return err
	}

	kegNames := formula.Names(kegs)
	slices.Sort(kegNames)
	kegNames = slices.Compact(kegNames)

	formulae, err := action.fetchFromArgs(ctx, kegNames, platform.SystemPlatform())
	if err != nil {
		return err
	}

	foundDependents := []string{}

	// Iterate over every installed formula, adding its direct dependencies to the list of known deps
	// This will create a list of every formula that is depended on by another installed formula
	// Once all
	for _, f := range formulae {
		fdeps := f.Dependencies()
		if fdeps == nil {
			continue
		}
		for _, d := range fdeps.Required {
			if !slices.Contains(foundDependents, d) {
				foundDependents = append(foundDependents, d)
			}
		}
	}

	// Iterate over every installed formula, if it is not a dependency, print the name
	for _, f := range formulae {
		if !slices.Contains(foundDependents, f.Name()) {
			fmt.Println(f.Name())
		}
	}

	return nil
}
