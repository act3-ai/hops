package actions

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
)

// Deps represents the action and its options.
type Deps struct {
	*Hops
	DependencyOptions formula.DependencyTags
	Platform          platform.Platform
}

// Tree runs the action.
func (action *Deps) Run(ctx context.Context, args []string) error {
	graph, err := action.resolve(ctx, args, action.Platform, &action.DependencyOptions)
	if err != nil {
		return err
	}

	deps := formula.Names(graph.Dependencies())
	slices.Sort(deps) // this messes up the ordering, but we do not care
	fmt.Println(strings.Join(deps, "\n"))

	return nil
}

// Tree runs the action.
func (action *Deps) Tree(ctx context.Context, args []string) error {
	deps, err := action.resolve(ctx, args, action.Platform, &action.DependencyOptions)
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
