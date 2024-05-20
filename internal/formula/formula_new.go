//nolint:all
package formula

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	v3 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v3"
	"github.com/act3-ai/hops/internal/platform"
)

type formulary interface {
	Fetch(name string) (Formula, error)
}

type concurrentFormulary interface {
	FetchAll(names []string) (Formulae, error)
}

// Formulae is a list of Formulae
type Formulae []Formula

// Formula represents a Homebrew Formula
type Formula interface {
	Name() string
	// Metadata() Metadata
	Version() Version
	Bottle(platform.Platform)
}

// formula represents a formula
type formula struct {
	Metadata

	name       string
	desc       string
	license    string
	homepage   string
	download   Download
	version    version
	RubySource RubySource
	Link       Link

	platforms map[platform.Platform]platformConfig
}

// FromV1 creates a formula from v1 API input
func FromV1(input *v1.Info) Formula {
	f := &formula{
		name:     input.Name,
		desc:     input.Desc,
		license:  input.License,
		homepage: input.Homepage,
		download: Download{
			// URL: ,
		},
		version: version{
			version: input.Versions.Stable,
		},
	}

	// if bottles, ok := input.Bottle[v1.Stable]; ok && bottles != nil {
	// 	f.platforms = make(map[platform.Platform]platformConfig, len(bottles.Files))

	// 	// for p, bf := range bottles.Files {
	// 	// }
	// }

	return f
}

func (f *formula) Name() string {
	return f.name
}

func (f *formula) Version() Version {
	return &f.version
}

func (f *formula) Bottle(plat platform.Platform) {
}

type Metadata struct {
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

type version struct {
	version  string
	revision int
	rebuild  int
}

func (v *version) Upstream() string {
	return v.version
}

func (v *version) Revision() int {
	return v.revision
}

func (v *version) Rebuild() int {
	return v.rebuild
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
	rootURL string
	cellar  string
	url     string
	sha256  string
}

func newFormula(info *v1.Info) *formula {
	stableURL := info.URLs[v1.Stable]
	return &formula{
		Metadata: Metadata{
			Name:     info.Name,
			Desc:     info.Desc,
			License:  info.License,
			Homepage: info.Homepage,
		},
		download: Download{
			URL:      stableURL.URL,
			Revision: stableURL.Revision,
			Tag:      stableURL.Tag,
			Branch:   stableURL.Branch,
			Using:    stableURL.Using,
			Checksum: stableURL.Checksum,
		},
	}
}
