package dependencies

import (
	"context"
	"fmt"
	"slices"

	"github.com/xlab/treeprint"

	"github.com/act3-ai/hops/internal/errdef"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
)

// DependencyGraph represents the evaluated dependency graph.
type DependencyGraph struct {
	rootKeys      []string                           // list of root formulae
	dependentKeys []string                           // list of dependency names, ordered
	formulae      map[string]formula.PlatformFormula // stores dependency information
	trees         map[string]*treeprint.Node         // stores dependency trees
}

// Dependents returns the list of computed dependencies.
func (deps DependencyGraph) Dependents() []formula.PlatformFormula {
	list := make([]formula.PlatformFormula, len(deps.dependentKeys))
	for i, name := range deps.dependentKeys {
		list[i] = deps.formulae[name]
	}
	return list
}

// Roots returns the list of computed dependencies.
func (deps *DependencyGraph) Roots() []formula.PlatformFormula {
	list := make([]formula.PlatformFormula, len(deps.rootKeys))
	for i, name := range deps.rootKeys {
		list[i] = deps.formulae[name]
	}
	return list
}

// Tree returns a printable tree of dependencies.
func (deps *DependencyGraph) Tree(root string) (treeprint.Tree, error) {
	tree, ok := deps.trees[root]
	if !ok {
		return nil, errdef.NewErrFormulaNotFound(root)
	}
	return tree, nil
}

// WalkPlatform evaluates the dependency graph of all root nodes for a specific platform.
func Walk(ctx context.Context, store formula.Formulary, roots []formula.PlatformFormula, plat platform.Platform, tags *formula.DependencyTags) (*DependencyGraph, error) {
	deps := &DependencyGraph{
		rootKeys:      []string{},
		dependentKeys: []string{},
		formulae:      map[string]formula.PlatformFormula{},
		trees:         map[string]*treeprint.Node{},
	}

	for _, f := range roots {
		deps.rootKeys = append(deps.rootKeys, f.Name())

		_, err := deps.add(ctx, store, f, plat, tags)
		if err != nil {
			return deps, err
		}
	}

	return deps, nil
}

// WalkAll evaluates the dependency graph of all root nodes.
//
// If dependencies vary by platform, all possible dependencies will be included.
func WalkAll(ctx context.Context, store formula.Formulary, roots []formula.PlatformFormula, tags *formula.DependencyTags) (*DependencyGraph, error) {
	return Walk(ctx, store, roots, platform.All, tags)
}

// add adds the given Formula to the found dependencies.
func (deps *DependencyGraph) add(ctx context.Context, store formula.Formulary, f formula.PlatformFormula, plat platform.Platform, tags *formula.DependencyTags) (*treeprint.Node, error) {
	key := f.Name()

	// Use trees to check for existence because the tree is added last
	n, ok := deps.trees[key]
	if ok {
		// Already found
		return n, nil
	}

	node := &treeprint.Node{Value: key}

	children := f.Dependencies().ForTags(tags)

	if !slices.Contains(deps.rootKeys, key) {
		deps.dependentKeys = append(deps.dependentKeys, key)
	}

	childformulae, err := formula.FetchAllPlatform(ctx, store, children, plat)
	if err != nil {
		return nil, err
	}

	for _, d := range childformulae {
		switch d := d.(type) {
		case formula.PlatformFormula:
			// Don't include indirect test dependencies
			child, err := deps.add(ctx, store, d, plat, withoutTest(tags))
			if err != nil {
				return nil, err
			}

			// Append to list of child nodes
			node.Nodes = append(node.Nodes, child)
		default:
			return nil, fmt.Errorf("no dependency information for formula %s", d.Name())
		}
	}

	deps.formulae[key] = f
	deps.trees[key] = node

	// Return my tree once all my children have been accounted for
	return node, nil
}

func withoutTest(tags *formula.DependencyTags) *formula.DependencyTags {
	return &formula.DependencyTags{
		IncludeBuild:    tags.IncludeBuild,
		IncludeTest:     false,
		SkipRecommended: tags.SkipRecommended,
		IncludeOptional: tags.IncludeOptional,
	}
}
