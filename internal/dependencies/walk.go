package dependencies

import (
	"context"
	"slices"

	"github.com/xlab/treeprint"

	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/brew"
)

// Options for dependency evaluation
type Options struct {
	SkipRecommended bool
	IncludeTest     bool
	IncludeOptional bool
	IncludeBuild    bool
}

// DependencyGraph represents the evaluated dependency graph
type DependencyGraph[T any] struct {
	store         Store[T]
	opts          *Options
	rootKeys      []string                   // list of root formulae
	dependentKeys []string                   // list of dependency names, ordered
	info          map[string]T               // stores dependency information
	trees         map[string]*treeprint.Node // stores dependency trees
}

// Store represents a searchable source of dependency information
type Store[T any] interface {
	Key(entry T) string
	DirectDependencies(ctx context.Context, entry T, opts *Options) ([]T, error)
	// Fetch(keys []string) ([]T, error)
	// DirectDependencies(ctx context.Context, node T, opts *Options) (keys []string, err error)
}

// Dependents returns the list of computed dependencies
func (deps DependencyGraph[T]) Dependents() []T {
	list := make([]T, len(deps.dependentKeys))
	for i, name := range deps.dependentKeys {
		list[i] = deps.info[name]
	}
	return list
}

// Roots returns the list of computed dependencies
func (deps *DependencyGraph[T]) Roots() []T {
	list := make([]T, len(deps.rootKeys))
	for i, name := range deps.rootKeys {
		list[i] = deps.info[name]
	}
	return list
}

// Tree returns a printable tree of dependencies
func (deps *DependencyGraph[T]) Tree(root string) (treeprint.Tree, error) {
	tree, ok := deps.trees[root]
	if !ok {
		return nil, brew.NewErrFormulaNotFound(root)
	}
	return tree, nil
}

// Walk evaluates the dependency graph of all root nodes
func Walk[T any](ctx context.Context, store Store[T], roots []T, opts *Options) (*DependencyGraph[T], error) {
	deps := &DependencyGraph[T]{
		store:         store,
		opts:          opts,
		rootKeys:      []string{},
		dependentKeys: []string{},
		info:          map[string]T{},
		trees:         map[string]*treeprint.Node{},
	}

	for _, f := range roots {
		deps.rootKeys = append(deps.rootKeys, deps.store.Key(f))

		_, err := deps.add(ctx, f)
		if err != nil {
			return deps, err
		}
	}

	return deps, nil
}

// add adds the given Formula to the found dependencies
func (deps *DependencyGraph[T]) add(ctx context.Context, f T) (*treeprint.Node, error) {
	key := deps.store.Key(f)

	// Use trees to check for existence because the tree is added last
	n, ok := deps.trees[key]
	if ok {
		// Already found
		return n, nil
	}

	node := &treeprint.Node{Value: key}

	var children []T
	var err error
	if slices.Contains(deps.rootKeys, key) {
		children, err = deps.store.DirectDependencies(ctx, f, deps.opts)
	} else {
		deps.dependentKeys = append(deps.dependentKeys, key)
		// Don't include indirect test dependencies
		children, err = deps.store.DirectDependencies(ctx, f, deps.opts.withoutTest())
	}
	if err != nil {
		return nil, err
	}

	for _, d := range children {
		child, err := deps.add(ctx, d)
		if err != nil {
			return nil, err
		}

		// Append to list of child nodes
		node.Nodes = append(node.Nodes, child)
	}

	deps.info[key] = f
	deps.trees[key] = node

	// Return my tree once all my children have been accounted for
	return node, nil
}

func (opts *Options) withoutTest() *Options {
	opts2 := opts.clone()
	opts2.IncludeTest = false
	return opts2
}

func (opts *Options) clone() *Options {
	return &Options{
		SkipRecommended: opts.SkipRecommended,
		IncludeTest:     opts.IncludeTest,
		IncludeOptional: opts.IncludeOptional,
		IncludeBuild:    opts.IncludeBuild,
	}
}

// ForOptions computes direct dependencies according to opts
func ForOptions(info *v1.PlatformInfo, opts *Options) []string {
	deps := slices.Clone(info.Dependencies)

	if !opts.SkipRecommended {
		deps = append(deps, info.RecommendedDependencies...)
	}
	if opts.IncludeBuild {
		deps = append(deps, info.BuildDependencies...)
	}
	if opts.IncludeTest {
		deps = append(deps, info.TestDependencies...)
	}
	if opts.IncludeOptional {
		deps = append(deps, info.OptionalDependencies...)
	}

	slices.Sort(deps)           // sort dependencies
	return slices.Compact(deps) // remove duplicates
}

// CategorizedDependencies stores dependencies in lists by kind
type CategorizedDependencies struct {
	Required    []string
	Build       []string
	Test        []string
	Recommended []string
	Optional    []string
}

// ToCategorized converts direct dependencies to a categorized struct
func ToCategorized(info *v1.PlatformInfo) *CategorizedDependencies {
	return &CategorizedDependencies{
		Required:    slices.Clone(info.Dependencies),
		Build:       slices.Clone(info.BuildDependencies),
		Test:        slices.Clone(info.TestDependencies),
		Recommended: slices.Clone(info.RecommendedDependencies),
		Optional:    slices.Clone(info.OptionalDependencies),
	}
}
