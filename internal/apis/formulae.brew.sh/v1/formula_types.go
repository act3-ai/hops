package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/platform"
)

// Index represents a formula index listing multiple formulae
type Index []*Info

// Info represents Homebrew API information for a formula
type Info struct {
	PlatformInfo `json:",inline"`
	Variations   map[platform.Platform]*PlatformInfo `json:"variations"`
}

// Info represents Homebrew API information for a formula
type PlatformInfo struct {
	Name                    string                `json:"name"`
	FullName                string                `json:"full_name"`
	Tap                     string                `json:"tap"`
	OldName                 string                `json:"oldname"`
	OldNames                []string              `json:"oldnames"`
	Aliases                 []string              `json:"aliases"`
	VersionedFormulae       []string              `json:"versioned_formulae"`
	Desc                    string                `json:"desc"`
	License                 string                `json:"license"`
	Homepage                string                `json:"homepage"`
	Versions                Versions              `json:"versions"`
	URLs                    map[string]FormulaURL `json:"urls"`
	Revision                int                   `json:"revision"`
	VersionScheme           int                   `json:"version_scheme"`
	Bottle                  map[string]*Bottle    `json:"bottle"`
	PourBottleOnlyIf        *string               `json:"pour_bottle_only_if"`
	KegOnly                 bool                  `json:"keg_only"`
	KegOnlyReason           KegOnlyConfig         `json:"keg_only_reason"`
	Options                 []any                 `json:"options"`
	BuildDependencies       []string              `json:"build_dependencies"`
	Dependencies            []string              `json:"dependencies"`
	TestDependencies        []string              `json:"test_dependencies"`
	RecommendedDependencies []string              `json:"recommended_dependencies"`
	OptionalDependencies    []string              `json:"optional_dependencies"`
	UsesFromMacOS           []any                 `json:"uses_from_macos"`
	UsesFromMacOSBounds     []*MacOSBounds        `json:"uses_from_macos_bounds"`
	Requirements            []*Requirement        `json:"requirements"`
	ConflictsWith           []string              `json:"conflicts_with"`
	ConflictsWithReasons    []string              `json:"conflicts_with_reasons"`
	LinkOverwrite           []string              `json:"link_overwrite"`
	Caveats                 *string               `json:"caveats"`
	Installed               []InstalledInfo       `json:"installed"`
	LinkedKeg               string                `json:"linked_keg"`
	Pinned                  bool                  `json:"pinned"`
	Outdated                bool                  `json:"outdated"`
	Deprecated              bool                  `json:"deprecated"`
	DeprecationDate         *string               `json:"deprecation_date"`
	DeprecationReason       *string               `json:"deprecation_reason"`
	Disabled                bool                  `json:"disabled"`
	DisabledDate            *string               `json:"disable_date"`
	DisabledReason          *string               `json:"disable_reason"`
	PostInstallDefined      bool                  `json:"post_install_defined"`
	Service                 *Service              `json:"service"`
	TapGitHead              string                `json:"tap_git_head"`
	RubySourcePath          string                `json:"ruby_source_path"`
	RubySourceChecksum      map[string]string     `json:"ruby_source_checksum"`
	HeadDependencies        *HeadDependencies     `json:"head_dependencies,omitempty"`
}

const (
	// RubySourceChecksumSha256 is the key for the sha256 checksum of a Formula's Ruby source
	RubySourceChecksumSha256 = "sha256"
)

const (
	PourBottleConditionDefaultPrefix = "default_prefix" // pour bottle condition requiring the default prefix
	PourBottleConditionCLTInstalled  = "clt_installed"  // pour bottle condition requiring the macOS command line tools
)

// FormulaURL represents the urls block
type FormulaURL struct {
	URL      string `json:"url"`
	Branch   string `json:"branch,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Revision string `json:"revision,omitempty"`
	Using    string `json:"using,omitempty"`
	Checksum string `json:"checksum,omitempty"`
}

// InstalledInfo represents the installed block
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

// RuntimeDependency represents a required dependency
type RuntimeDependency struct {
	FullName         string `json:"full_name"`
	Version          string `json:"version"`
	Revision         int    `json:"revision"`
	PkgVersionValue  string `json:"pkg_version"`
	DeclaredDirectly bool   `json:"declared_directly"`
}

// Repo implements the same name formatting as other bottle parsing functions
func (rd *RuntimeDependency) Repo() string {
	repo := rd.FullName                       // start with formula name
	repo = strings.ReplaceAll(repo, "@", "/") // replace "@" with "/"
	repo = strings.ReplaceAll(repo, "+", "x") // replace "+" with "x"
	return repo
}

// PkgVersion returns the package version
func (rd *RuntimeDependency) PkgVersion() string {
	v := rd.PkgVersionValue
	if v == "" {
		v = rd.Version
	}
	return v
}

// Tag implements the same tag naming logic as other bottle parsing functions
func (rd *RuntimeDependency) Tag() string {
	v := rd.PkgVersion()
	if rd.Revision != 0 {
		v += fmt.Sprintf("-%d", rd.Revision)
	}
	return v
}

// Variation represents an entry in the variations map
type Variation Info

// type Variation struct {
// 	PourBottleOnlyIf     *PourBottleCondition `json:"pour_bottle_only_if,omitempty"`
// 	KegOnly              bool                 `json:"keg_only,omitempty"`
// 	KegOnlyReason        KegOnlyReasonInfo    `json:"keg_only_reason,omitempty"`
// 	BuildDependencies    []string             `json:"build_dependencies,omitempty"`
// 	Dependencies         []string             `json:"dependencies,omitempty"`
// 	TestDependencies     []string             `json:"test_dependencies,omitempty"`
// 	Requirements         []*Requirement       `json:"requirements,omitempty"`
// 	ConflictsWith        []string             `json:"conflicts_with,omitempty"`
// 	ConflictsWithReasons []string             `json:"conflicts_with_reasons,omitempty"`
// 	HeadDependencies     *HeadDependencies    `json:"head_dependencies,omitempty"`
// 	Caveats              *string              `json:"caveats,omitempty"`
// 	Disabled             bool                 `json:"disabled,omitempty"`
// 	DisabledDate         *string              `json:"disable_date,omitempty"`
// 	DisabledReason       *string              `json:"disable_reason,omitempty"`
// }

// HeadDependencies represents the head_dependencies field
type HeadDependencies struct {
	BuildDependencies       []string       `json:"build_dependencies"`
	Dependencies            []string       `json:"dependencies"`
	TestDependencies        []string       `json:"test_dependencies"`
	RecommendedDependencies []string       `json:"recommended_dependencies"`
	OptionalDependencies    []string       `json:"optional_dependencies"`
	UsesFromMacOS           []any          `json:"uses_from_macos"`
	UsesFromMacOSBounds     []*MacOSBounds `json:"uses_from_macos_bounds"`
}

// Versions represents the available versions
type Versions struct {
	Others map[string]any `json:",inline"`
	Stable string         `json:"stable"`
	Head   *string        `json:"head"`
	Bottle bool           `json:"bottle"`
}

const (
	Stable = "stable" // Key used for stable bottles
)

// Bottle represents the bottle section
type Bottle struct {
	Rebuild int                               `json:"rebuild"`
	RootURL string                            `json:"root_url"`
	Files   map[platform.Platform]*BottleFile `json:"files"`
}

// MacOSBounds represents the uses_from_macos_bounds entries
type MacOSBounds struct {
	Since string `json:"since"`
}

// Requirement represents a requirement
type Requirement struct {
	Name     string   `json:"name"`
	Cask     any      `json:"cask"`
	Download any      `json:"download"`
	Version  string   `json:"version"`
	Contexts []string `json:"contexts"`
	Specs    []string `json:"specs"`
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

// KegOnlyConfig represents the keg_only_reason section
type KegOnlyConfig struct {
	Reason      KegOnlyReason `json:"reason"`
	Explanation string        `json:"explanation"`
}

// BottleFile defines a bottle.files entry
type BottleFile struct {
	Cellar string `json:"cellar"`
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
}

// Relocatable reports if the bottle is relocatable
func (file *BottleFile) Relocatable() bool {
	return file.Cellar == CellarAny || file.Cellar == CellarAnySkipRelocation
}

const (
	CellarAny               = ":any"                 // Signifies bottle is safe to install in the Cellar
	CellarAnySkipRelocation = ":any_skip_relocation" // Signifies bottle is safe to install in the Cellar without relocation
)

// String implements fmt.Stringer
func (info *Info) String() string {
	marshalled, err := json.MarshalIndent(info, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

// Version returns the version of the formula according to Homebrew
//
// Pattern: VERSION[_REVISION]
//
// This version will vary from the formula project's version when
// the "revision" field is set in the formula, which signals the
// formula was updated without changing the version being installed
func (info *PlatformInfo) Version() string {
	// if info.Versions.Stable == nil {
	// 	return "HEAD"
	// }

	tag := info.Versions.Stable
	if info.Revision != 0 {
		tag += fmt.Sprintf("_%d", info.Revision)
	}

	return tag
}

// PossibleNames returns all possible names for the formula
// This is a combination of the current name, old names, and any aliases
func (info *PlatformInfo) PossibleNames() []string {
	names := []string{
		info.Name,
	}
	names = append(names, info.OldNames...)
	names = append(names, info.Aliases...)
	return names
}

// ManifestTag produces the tag of the formula's manifest
func (info *PlatformInfo) ManifestTag(key string) (string, error) {
	bottle := info.Bottle[key]
	if bottle == nil {
		return "", fmt.Errorf("formula %s does not have a %s bottle", info.Name, key)
	}

	version := ""
	if key == Stable {
		if info.Versions.Stable == "" {
			return "", fmt.Errorf("formula %s does not have a %s version", info.Name, key)
		}
		version = info.Versions.Stable
	}

	return brewfmt.Tag(version, info.Revision, bottle.Rebuild), nil
}

// Names returns the names of the listed formulae
func Names(formulae []*Info) []string {
	names := make([]string, len(formulae))
	for i, f := range formulae {
		names[i] = f.Name
	}
	return names
}
