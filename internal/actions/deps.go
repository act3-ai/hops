package actions

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/dependencies"
	"github.com/act3-ai/hops/internal/platform"
)

// Deps represents the action and its options
type Deps struct {
	*Hops
	Standalone        bool
	DependencyOptions formula.DependencyTags
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
		tree, err := deps.Tree(root.Name())
		if err != nil {
			return err
		}
		fmt.Println(tree)
	}

	return nil
}

func (action *Deps) eval(ctx context.Context, args []string) (*dependencies.DependencyGraph, error) {
	client, err := action.FormulaClient(ctx, args)
	if err != nil {
		return nil, err
	}

	formulae, err := action.fetchFromArgs(ctx, args, action.Platform)
	if err != nil {
		return nil, err
	}

	graph, err := dependencies.Walk(ctx,
		client,
		formulae,
		action.Platform,
		&action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
