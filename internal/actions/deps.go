package actions

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/act3-ai/hops/internal/dependencies"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
)

// Deps represents the action and its options.
type Deps struct {
	*Hops
	Standalone        bool
	DependencyOptions formula.DependencyTags
	Platform          platform.Platform
}

// Tree runs the action.
func (action *Deps) Run(ctx context.Context, names ...string) error {
	graph, err := action.eval(ctx, names)
	if err != nil {
		return err
	}

	deps := formula.Names(graph.Dependencies())
	slices.Sort(deps) // this messes up the ordering, but we do not care
	fmt.Println(strings.Join(deps, "\n"))

	return nil
}

// Tree runs the action.
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
	names := action.SetAlternateTags(args)

	formulary, err := action.Formulary(ctx)
	if err != nil {
		return nil, err
	}

	formulae, err := formula.FetchAllPlatform(ctx, formulary, names, action.Platform)
	if err != nil {
		return nil, err
	}

	graph, err := dependencies.Walk(ctx,
		formulary,
		formulae,
		action.Platform,
		&action.DependencyOptions)
	if err != nil {
		return nil, err
	}

	return graph, nil
}
