package v1

import (
	"encoding/json"
	"fmt"

	"github.com/act3-ai/hops/internal/apis/formulae.brew.sh/common"
	"github.com/act3-ai/hops/internal/platform"
)

// Index represents a formula index listing multiple formulae.
//
//easyjson:json
type Index []*Info

// Info represents Homebrew API information for a formula.
type Info struct {
	PlatformInfo `json:",inline"`
	Variations   map[platform.Platform]*PlatformInfo `json:"variations"`
}

// PlatformInfo represents Homebrew API information for a formula.
type PlatformInfo struct {
	Name                    string                `json:"name"`
	FullName                string                `json:"full_name,omitempty"` // Deprecated: Evaluate from Tap/Name
	Tap                     string                `json:"tap,omitempty,intern"`
	OldName                 string                `json:"oldname,omitempty"` // Deprecated: Use OldNames list
	OldNames                []string              `json:"oldnames,omitempty"`
	Aliases                 []string              `json:"aliases,omitempty"`
	VersionedFormulae       []string              `json:"versioned_formulae,omitempty"`
	Desc                    string                `json:"desc,omitempty"`
	License                 string                `json:"license,omitempty"`
	Homepage                string                `json:"homepage,omitempty"`
	Versions                Versions              `json:"versions,omitempty"`
	URLs                    map[string]FormulaURL `json:"urls,omitempty"`
	Revision                int                   `json:"revision,omitempty"`
	VersionScheme           int                   `json:"version_scheme,omitempty"`
	Bottle                  map[string]*Bottle    `json:"bottle,omitempty"`
	PourBottleOnlyIf        *string               `json:"pour_bottle_only_if,omitempty"`
	KegOnly                 bool                  `json:"keg_only,omitempty"`
	KegOnlyReason           common.KegOnlyConfig  `json:"keg_only_reason,omitempty"`
	Options                 []any                 `json:"options,omitempty"`
	BuildDependencies       []string              `json:"build_dependencies,omitempty"`
	Dependencies            []string              `json:"dependencies,omitempty"`
	TestDependencies        []string              `json:"test_dependencies,omitempty"`
	RecommendedDependencies []string              `json:"recommended_dependencies,omitempty"`
	OptionalDependencies    []string              `json:"optional_dependencies,omitempty"`
	UsesFromMacOS           []any                 `json:"uses_from_macos,omitempty"`
	UsesFromMacOSBounds     []*MacOSBounds        `json:"uses_from_macos_bounds,omitempty"`
	Requirements            []*Requirement        `json:"requirements,omitempty"`
	ConflictsWith           []string              `json:"conflicts_with,omitempty"`
	ConflictsWithReasons    []string              `json:"conflicts_with_reasons,omitempty"`
	LinkOverwrite           []string              `json:"link_overwrite,omitempty"`
	Caveats                 *string               `json:"caveats,omitempty"`
	Installed               []InstalledInfo       `json:"installed,omitempty"`
	LinkedKeg               string                `json:"linked_keg,omitempty"`
	Pinned                  bool                  `json:"pinned,omitempty"`
	Outdated                bool                  `json:"outdated,omitempty"`
	Deprecated              bool                  `json:"deprecated,omitempty"`
	DeprecationDate         *string               `json:"deprecation_date,omitempty"`
	DeprecationReason       *string               `json:"deprecation_reason,omitempty"`
	Disabled                bool                  `json:"disabled,omitempty"`
	DisabledDate            *string               `json:"disable_date,omitempty"`
	DisabledReason          *string               `json:"disable_reason,omitempty"`
	PostInstallDefined      bool                  `json:"post_install_defined,omitempty"`
	Service                 *common.Service       `json:"service,omitempty"`
	TapGitHead              string                `json:"tap_git_head,omitempty,intern"`
	RubySourcePath          string                `json:"ruby_source_path,omitempty"`
	RubySourceChecksum      map[string]string     `json:"ruby_source_checksum,omitempty"`
	HeadDependencies        *HeadDependencies     `json:"head_dependencies,omitempty"`
}

const (
	// RubySourceChecksumSha256 is the key for the sha256 checksum of a Formula's Ruby source.
	RubySourceChecksumSha256 = "sha256"
)

// FormulaURL represents the urls block.
type FormulaURL struct {
	URL      string `json:"url"`
	Branch   string `json:"branch,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Revision string `json:"revision,omitempty"`
	Using    string `json:"using,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}

// InstalledInfo represents the installed block.
type InstalledInfo struct {
	Version               string               `json:"version"`
	UsedOptions           []any                `json:"used_options"`
	BuiltAsBottle         bool                 `json:"built_as_bottle"`
	PouredFromBottle      bool                 `json:"poured_from_bottle"`
	Time                  int                  `json:"time"`
	RuntimeDependencies   []*RuntimeDependency `json:"runtime_dependencies"`
	InstalledAsDependency bool                 `json:"installed_as_dependency"`
	InstalledOnRequest    bool                 `json:"installed_on_request"`
}

// RuntimeDependency represents a required dependency.
type RuntimeDependency struct {
	FullName         string `json:"full_name,omitempty"`
	Version          string `json:"version,omitempty"`
	Revision         int    `json:"revision,omitempty"`
	PkgVersionValue  string `json:"pkg_version,omitempty"`
	DeclaredDirectly bool   `json:"declared_directly,omitempty"`
}

// Variation represents an entry in the variations map.
type Variation PlatformInfo

// HeadDependencies represents the head_dependencies field.
type HeadDependencies struct {
	BuildDependencies       []string       `json:"build_dependencies,omitempty"`
	Dependencies            []string       `json:"dependencies,omitempty"`
	TestDependencies        []string       `json:"test_dependencies,omitempty"`
	RecommendedDependencies []string       `json:"recommended_dependencies,omitempty"`
	OptionalDependencies    []string       `json:"optional_dependencies,omitempty"`
	UsesFromMacOS           []any          `json:"uses_from_macos,omitempty"`
	UsesFromMacOSBounds     []*MacOSBounds `json:"uses_from_macos_bounds,omitempty"`
}

// Versions represents the available versions.
type Versions struct {
	Others map[string]any `json:",inline"`
	Stable string         `json:"stable,omitempty"`
	Head   *string        `json:"head,omitempty"`
	Bottle bool           `json:"bottle,omitempty"`
}

const (
	Stable = "stable" // Key used for stable bottles
)

// Bottle represents the bottle section.
type Bottle struct {
	Rebuild int                               `json:"rebuild,omitempty"`
	RootURL string                            `json:"root_url,omitempty,intern"`
	Files   map[platform.Platform]*BottleFile `json:"files,omitempty"`
}

// MacOSBounds represents the uses_from_macos_bounds entries.
type MacOSBounds struct {
	Since string `json:"since,omitempty"`
}

// Requirement represents a requirement.
type Requirement struct {
	Name     string   `json:"name,omitempty"`
	Cask     any      `json:"cask,omitempty"`
	Download any      `json:"download,omitempty"`
	Version  string   `json:"version,omitempty"`
	Contexts []string `json:"contexts,omitempty"`
	Specs    []string `json:"specs,omitempty"`
}

// if requirement name is confirmed to be a static list, use the below type
// type RequirementName string
// const (
// 	RequirementNameMaximumMacOS RequirementName = "maximum_macos"       // name of requirement specifying maximum macOS version
// 	RequirementNameArch         RequirementName = "arch"                // name of requirement specifying architecture
// 	RequirementNameMacOS        RequirementName = "macos"               // name of requirement specifying macOS version
// 	RequirementNameXCode        RequirementName = "xcode"               // name of requirement specifying XCode
// 	RequirementNameLinux        RequirementName = "linux"               // name of requirement specifying Linux-only
// 	RequirementNameGlibc        RequirementName = "brewedglibcnotolder" // name of requirement specifying glibc version
// 	RequirementNameLinuxKernel  RequirementName = "linuxkernel"         // name of requirement specifying Linux kernel version
// )

// BottleFile defines a bottle.files entry.
type BottleFile struct {
	Cellar string `json:"cellar,omitempty,intern"`
	URL    string `json:"url,omitempty"`
	Sha256 string `json:"sha256,omitempty"`
}

// Relocatable reports if the bottle is relocatable.
func (file *BottleFile) Relocatable() bool {
	return common.CellarRelocatable(file.Cellar)
}

// String implements fmt.Stringer.
func (info *Info) String() string {
	marshalled, err := json.MarshalIndent(info, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

// Version returns the version of the formula according to Homebrew.
//
// Pattern:
//
//	VERSION[_REVISION]
//
// This version will vary from the formula project's version when
// the "revision" field is set in the formula, which signals the
// formula was updated without changing the version being installed.
func (info *PlatformInfo) Version() string {
	tag := info.Versions.Stable
	if info.Revision != 0 {
		tag += fmt.Sprintf("_%d", info.Revision)
	}

	return tag
}

// PossibleNames returns all possible names for the formula.
// This is a combination of the current name, old names, and any aliases.
func (info *PlatformInfo) PossibleNames() []string {
	names := []string{
		info.Name,
	}
	names = append(names, info.OldNames...)
	names = append(names, info.Aliases...)
	return names
}
