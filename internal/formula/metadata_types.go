package formula

import (
	"log/slog"
	"slices"

	"github.com/act3-ai/hops/internal/platform"
)

// Metadata types.
type (
	// Info defines the most general information for a Formula.
	Info struct {
		Desc     string
		License  string
		Homepage string
	}

	// SourceInfo defines source information for a Formula.
	SourceInfo struct {
		URL      string
		Using    string
		Checksum string
		Git      GitSource
		Ruby     RubySource
	}

	// GitSource defines the Git source for a Formula.
	GitSource struct {
		Revision string
		Tag      string
		Branch   string
	}

	// RubySource defines the Ruby source for a Formula.
	RubySource struct {
		Path   string
		Sha256 string
	}

	// Conflict defines a conflicting package.
	Conflict struct {
		Name   string
		Reason string
	}

	// Bottle defines bottle metadata.
	Bottle struct {
		RootURL    string
		Sha256     string
		Cellar     string
		Platform   platform.Platform // stores Bottle platform, which can vary from PlatformFormula.Platform() iff the Bottle is for "all" platforms.
		PourOnlyIf string            // pour_bottle_only_if rule
	}
)

// version is an implementation of Version.
type version struct {
	version  string
	revision int
	rebuild  int
}

// Upstream implements Version.
func (v *version) Upstream() string {
	return v.version
}

// Revision implements Version.
func (v *version) Revision() int {
	return v.revision
}

// Rebuild implements Version.
func (v *version) Rebuild() int {
	return v.rebuild
}

// Dependency types.
type (
	// DependencyTags defines the available dependency tags.
	DependencyTags struct {
		IncludeBuild    bool
		IncludeTest     bool
		SkipRecommended bool
		IncludeOptional bool
	}

	// TaggedDependencies stores dependencies in lists by tag.
	TaggedDependencies struct {
		Required    []string
		Build       []string
		Test        []string
		Recommended []string
		Optional    []string
	}
)

// // taggedDependencies stores dependencies in lists by tag.
// type taggedDependencies struct {
// 	required    []string
// 	build       []string
// 	test        []string
// 	recommended []string
// 	optional    []string
// }

// ForTags implements Dependencies.
func (deps *TaggedDependencies) ForTags(tags *DependencyTags) []string {
	result := slices.Clone(deps.Required)

	if tags.IncludeBuild {
		result = append(result, deps.Build...)
	}

	if tags.IncludeTest {
		result = append(result, deps.Test...)
	}

	if tags.SkipRecommended {
		result = append(result, deps.Recommended...)
	}

	if tags.IncludeOptional {
		result = append(result, deps.Optional...)
	}

	return result
}

// LogAttr formats the tags as a slog.Attr.
func (deps *DependencyTags) LogAttr() slog.Attr {
	return slog.Group(
		"tags",
		slog.Bool("build", deps.IncludeBuild),
		slog.Bool("test", deps.IncludeTest),
		slog.Bool("recommended", !deps.SkipRecommended),
		slog.Bool("optional", deps.IncludeOptional),
	)
}
