package v3

import (
	"github.com/act3-ai/hops/internal/apis/formulae.brew.sh/common"
	"github.com/act3-ai/hops/internal/platform"
)

// Formula represents a formula's metadata.
type Formula struct {
	PlatformFormula `json:",inline"`
	Variations      map[platform.Platform]PlatformFormula `json:"variations"`
}

// PlatformFormula represents a formula's metadata for a specific platform.
type PlatformFormula struct {
	Desc     string `json:"desc"`
	License  string `json:"license"`
	Homepage string `json:"homepage"`
	URLs     map[string]struct {
		URL      string `json:"url"`
		Revision string `json:"revision,omitempty"`
		Tag      string `json:"tag,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Using    string `json:"using,omitempty"`
		Checksum string `json:"checksum,omitempty"`
	} `json:"urls"`
	PostInstallDefined bool                 `json:"post_install_defined"`
	RubySourcePath     string               `json:"ruby_source_path"`
	RubySourceSHA256   string               `json:"ruby_source_sha256"`
	LinkOverwrite      []string             `json:"link_overwrite,omitempty"`
	Revision           int                  `json:"revision,omitempty"`
	KegOnlyReason      common.KegOnlyConfig `json:"keg_only_reason,omitempty"`
	PourBottleOnlyIf   string               `json:"pour_bottle_only_if,omitempty"`
	Caveats            string               `json:"caveats,omitempty"`
	Service            common.Service       `json:"service,omitempty"`
	VersionScheme      int                  `json:"version_scheme,omitempty"`
	Version            string               `json:"version"`
	Bottle             Bottle               `json:"bottle"`
	VersionedFormulae  []string             `json:"versioned_formulae,omitempty"`
	DeprecationDate    string               `json:"deprecation_date,omitempty"`
	DeprecationReason  string               `json:"deprecation_reason,omitempty"`
	DisabledDate       string               `json:"disable_date,omitempty"`
	DisabledReason     string               `json:"disable_reason,omitempty"`
	Dependencies       Dependencies         `json:"dependencies,omitempty"`
	HeadDependencies   Dependencies         `json:"head_dependencies,omitempty"`
	Requirements       []Requirement        `json:"requirements,omitempty"`
	Conflicts          Conflicts            `json:",inline"`
}

// Variation represents a platform-specific variation to the formula's metadata.
type Variation struct {
	// v3 caveats can only be set to a string that overwrites the general caveats.
	Caveats          string        `json:"caveats,omitempty"`
	Dependencies     Dependencies  `json:"dependencies,omitempty"`
	HeadDependencies Dependencies  `json:"head_dependencies,omitempty"`
	Requirements     []Requirement `json:"requirements,omitempty"`
	Conflicts        Conflicts     `json:",inline"`
}

// Bottle represents the bottle section.
type Bottle struct {
	Rebuild int                              `json:"rebuild"`
	RootURL string                           `json:"root_url"`
	Files   map[platform.Platform]BottleFile `json:"files"`
}

// BottleFile defines a bottle.files entry.
type BottleFile struct {
	Cellar string `json:"cellar"`
	Sha256 string `json:"sha256"`
}

// Dependencies represents a collection of dependencies.
type Dependencies map[string]*DependencyConfig

// DependencyConfig provides additional context for a dependency.
type DependencyConfig struct {
	Tags          []string     `json:"tags,omitempty"`
	UsesFromMacOS *MacOSBounds `json:"uses_from_macos,omitempty"`
}

// MacOSBounds constrains a macOS dependency.
type MacOSBounds struct {
	Since string `json:"since,omitempty"`
}

// Requirement represents a requirement.
type Requirement struct {
	Name     string   `json:"name"`
	Cask     any      `json:"cask"`
	Download any      `json:"download"`
	Version  *string  `json:"version"`
	Contexts []string `json:"contexts"`
	Specs    []string `json:"specs"`
}

// Conflicts specifies formula conflicts.
type Conflicts struct {
	ConflictsWith        []string `json:"conflicts_with,omitempty"`
	ConflictsWithReasons []string `json:"conflicts_with_reasons,omitempty"`
}
