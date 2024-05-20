package formula

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	v3 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v3"
	"github.com/act3-ai/hops/internal/platform"
)

// formula represents a formula
type formula struct {
	metadata

	Download   Download
	Version    Version
	RubySource RubySource
	Link       Link

	platforms map[platform.Platform]platformConfig
}

type metadata struct {
	Name     string
	Desc     string
	License  string
	Homepage string
}

type Download struct {
	URL      string
	Revision string
	Tag      string
	Branch   string
	Using    string
	Checksum string
}

type RubySource struct {
	Path   string
	Sha256 string
}

type Version struct {
	Upstream string
	Revision int
	Rebuild  int
}

type Link struct {
	Overwrite     []string
	KegOnly       bool
	kegonlyReason string
}

type platformConfig struct {
	caveats      string
	dependencies taggedDependencies
	requirements []v3.Requirement
	conflicts    []conflict
	bottle       bottle
}

// TaggedDependencies stores dependencies in lists by tag
type taggedDependencies struct {
	required    []dependency
	build       []dependency
	test        []dependency
	recommended []dependency
	optional    []dependency
}

// dependency represents a dependency
type dependency struct {
	name              string
	useFromMacOS      bool
	sinceMacOSVersion string
}

type conflict struct {
	name   string
	reason string
}

type bottle struct {
	rebuild int
	rootURL string
	cellar  string
	url     string
	sha256  string
}

func newFormula(info *v1.Info) *formula {
	stableURL := info.URLs[v1.Stable]
	return &formula{
		metadata: metadata{
			Name:     info.Name,
			Desc:     info.Desc,
			License:  info.License,
			Homepage: info.Homepage,
		},
		Download: Download{
			URL:      stableURL.URL,
			Revision: stableURL.Revision,
			Tag:      stableURL.Tag,
			Branch:   stableURL.Branch,
			Using:    stableURL.Using,
			Checksum: stableURL.Checksum,
		},
	}
}
