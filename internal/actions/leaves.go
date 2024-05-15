package actions

import (
	"context"
	"fmt"
	"slices"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
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
	plat := platform.SystemPlatform()

	formulae, err := action.resolveInstalled(ctx)
	if err != nil {
		return err
	}

	deps := []string{}

	// Iterate over every installed formula, adding its direct dependencies to the list of known deps
	// This will create a list of every formula that is depended on by another installed formula
	// Once all
	for _, f := range formulae {
		platinfo, err := f.Info.ForPlatform(plat)
		if err != nil {
			return err
		}

		for _, d := range platinfo.Dependencies {
			if !slices.Contains(deps, d) {
				deps = append(deps, d)
			}
		}
	}

	// Iterate over every installed formula, if it is not a dependency, print the name
	for _, f := range formulae {
		if !slices.Contains(deps, f.Name) {
			fmt.Println(f.Name)
		}
	}

	return nil
}

func (action *Leaves) resolveInstalled(ctx context.Context) ([]*formula.Formula, error) {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return nil, err
	}

	// List all racks
	racks, err := action.Prefix().Racks()
	if err != nil {
		return nil, err
	}

	rackNames := make([]string, 0, len(racks))
	for _, r := range racks {
		rackNames = append(rackNames, r.Name())
	}

	return action.FetchAll(o.Noop, index, rackNames...)
}
