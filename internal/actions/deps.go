package actions

import (
	"context"
	"fmt"
	"slices"
	"strings"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/dependencies"
	apiwalker "github.com/act3-ai/hops/internal/dependencies/api"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
)

// Deps represents the action and its options
type Deps struct {
	*Hops
	Standalone        bool
	DependencyOptions dependencies.Options
	Platform          platform.Platform
}

// Tree runs the action
func (action *Deps) Run(ctx context.Context, names ...string) error {
	deps, err := action.eval(ctx, names)
	if err != nil {
		return err
	}

	dependents := formula.Names(deps.Dependents())
	slices.Sort(dependents) // this messes up the ordering, but we do not care
	fmt.Println(strings.Join(dependents, "\n"))

	return nil
}

// Tree runs the action
func (action *Deps) Tree(ctx context.Context, names ...string) error {
	deps, err := action.eval(ctx, names)
	if err != nil {
		return err
	}

	// Print each rooted tree
	for _, root := range deps.Roots() {
		tree, err := deps.Tree(root.Name)
		if err != nil {
			return err
		}
		fmt.Println(tree)
	}

	return nil
}

func (action *Deps) eval(ctx context.Context, names []string) (*dependencies.DependencyGraph[*v1.Info], error) {
	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return nil, err
	}

	formulae, err := action.FetchAll(o.Noop, index, names...)
	if err != nil {
		return nil, err
	}

	graph, err := dependencies.Walk(ctx,
		apiwalker.New(index, action.Platform),
		formulae,
		&action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
