//nolint:all
package formula

import (
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	v3 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v3"
	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/platform"
)

// General Formula types
type (
	Namer interface {
		Name() string
	}

	Versioner interface {
		Version() Version
	}

	// NameVersioner is implemented by all Formula types
	NameVersioner interface {
		Namer
		Versioner
	}
)

// Formula types
type (
	// Formula represents a Homebrew Formula
	Formula interface {
		NameVersioner
		// Info produces basic information about the Formula
		Info() *Info
	}

	// MultiPlatformFormula is implemented by formulae that support multiple platforms
	MultiPlatformFormula interface {
		Formula
		// ForPlatform produces a PlatformFormula for the given platform
		ForPlatform(plat platform.Platform) (PlatformFormula, error)
	}

	// PlatformFormula represents a Homebrew Formula for a specific platform
	PlatformFormula interface {
		Formula
		// Platform produces the platform for this Formula
		Platform() platform.Platform
		// SourceInfo produces information about the Formula's source
		SourceInfo() *SourceInfo
		// Caveats produces the Formula's caveats, if any
		Caveats() string
		// Dependencies lists dependencies on other Formulae
		Dependencies() *TaggedDependencies
		// SystemDependencies lists dependencies on system software
		SystemDependencies() *TaggedDependencies
		// Conflicts lists conflicts with other formulae
		Conflicts() []Conflict
		// LinkOverwrite lists links to be overwritten in the prefix
		LinkOverwrite() []string
		// IsKegOnly reports whether the Formula is keg-only
		IsKegOnly() bool
		// KegOnlyReason produces the reason why a Formula is keg-only
		KegOnlyReason() (reason string)
		// Requirements lists other system requirements
		// Requirements() []any
		// Service produces the Formula's service, if any
		Service() *brewv1.Service
		// Bottle produces information about the Formula's Bottle. Bottle will return nil if the Formula does not provide a Bottle.
		Bottle() *Bottle
	}

	// // FormulaWithInfo represents a Homebrew Formula with full metadata
	// FormulaWithInfo interface {
	// 	Formula
	// 	Info() *Info
	// }

	// // MultiPlatformFormulaWithInfo represents a Homebrew Formula with full metadata
	// MultiPlatformFormulaWithInfo interface {
	// 	MultiPlatformFormula
	// 	FormulaWithInfo
	// }

	// // PlatformFormulaWithInfo represents a Homebrew Formula with full metadata
	// PlatformFormulaWithInfo interface {
	// 	PlatformFormula
	// 	FormulaWithInfo

	// 	// Caveats produces the Formula's caveats, if any
	// 	Caveats() string
	// 	// Dependencies lists dependencies on other Formulae
	// 	Dependencies() Dependencies
	// 	// SystemDependencies lists dependencies on system software
	// 	SystemDependencies() Dependencies
	// 	// Conflicts lists conflicts with other formulae
	// 	Conflicts() []Conflict
	// 	// LinkOverwrite lists links to be overwritten in the prefix
	// 	LinkOverwrite() []string
	// 	// IsKegOnly reports whether the Formula is keg-only
	// 	IsKegOnly() bool
	// 	// KegOnlyReason produces the reason why a Formula is keg-only
	// 	KegOnlyReason() (reason string)
	// 	// Requirements lists other system requirements
	// 	// Requirements() []any
	// 	Service() *brewv1.Service
	// }
)

// // List types
// type (
// 	// Formulae is a list of Formula
// 	Formulae []Formula
// 	// MultiPlatformFormulae is a list of MultiPlatformFormula
// 	MultiPlatformFormulae []MultiPlatformFormula
// 	// PlatformFormulae is a list of PlatformFormula
// 	PlatformFormulae []PlatformFormula

// 	// // FormulaeWithInfo is a list of FormulaWithInfo
// 	// FormulaeWithInfo []FormulaWithInfo
// 	// // MultiPlatformFormulaeWithInfo is a list of MultiPlatformFormulaWithInfo
// 	// MultiPlatformFormulaeWithInfo []MultiPlatformFormulaWithInfo
// 	// // PlatformFormulaeWithInfo is a list of PlatformFormulaWithInfo
// 	// PlatformFormulaeWithInfo []PlatformFormulaWithInfo
// )

// // Metadata types
// type (
// 	FormulaV1 interface {
// 		V1() *brewv1.Info
// 	}
// )

// // formula represents a formula
// type formula struct {
// 	info    Info
// 	version version
// 	name    string
// 	source  SourceInfo
// 	Link    Link
// 	v1 *brewv1.Info
// 	evaluatedPlatforms map[platform.Platform]platformConfig
// }

// // platformFormula is an implementation of PlatformFormula
// type platformFormula struct {
// 	formula
// 	platformConfig
// }

type platformConfig struct {
	caveats             string
	formulaDependencies TaggedDependencies
	systemDependencies  TaggedDependencies // uses_from_macos
	requirements        []v3.Requirement
	conflicts           []Conflict
	bottle              bottle
}

type bottle struct {
	rootURL string
	cellar  string
	url     string
	sha256  string
}

func Names[T Namer](formulae []T) []string {
	names := make([]string, len(formulae))
	for i, f := range formulae {
		names[i] = f.Name()
	}
	return names
}

// BottleFileName returns a short name for the downloaded bottle .tar.gz file for the formula
//
// Pattern: NAME--VERSION[_REVISION][-REBUILD]
//
// Example: cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
func BottleFileName(f PlatformFormula) string {
	version := f.Version()
	return brewfmt.ArchiveFile(
		f.Name(),
		version.Upstream(), version.Revision(), version.Rebuild(),
		f.Platform())
}