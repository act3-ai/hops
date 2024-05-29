package platform

import (
	"errors"
	"fmt"
	"slices"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// ErrInvalidPlatform represents a platform value error.
var ErrInvalidPlatform = errors.New("invalid platform")

// NewErrInvalidPlatform wraps ErrInvalidPlatform with the invalid platform value.
func NewErrInvalidPlatform(value string) error {
	return fmt.Errorf("%q: %w", value, ErrInvalidPlatform)
}

// Platform defines a platform.
type Platform string

// Set implements pflag.Value.
func (p *Platform) Set(s string) error {
	if !IsValid(s) {
		return fmt.Errorf("setting flag value %q: %w", s, ErrInvalidPlatform)
	}
	*p = Platform(s)
	return nil
}

// Type implements pflag.Value.
func (Platform) Type() string {
	return "platform"
}

const (
	Arm64Sonoma   Platform = "arm64_sonoma"   // macOS Sonoma ARM 64-bit
	Arm64Ventura  Platform = "arm64_ventura"  // macOS Sonoma ARM 64-bit
	Arm64Monterey Platform = "arm64_monterey" // macOS Sonoma ARM 64-bit
	Arm64BigSur   Platform = "arm64_big_sur"  // macOS Sonoma ARM 64-bit
	Sonoma        Platform = "sonoma"         // macOS Sonoma x86 64-bit
	Ventura       Platform = "ventura"        // macOS Ventura x86 64-bit
	Monterey      Platform = "monterey"       // macOS Monterey x86 64-bit
	BigSur        Platform = "big_sur"        // macOS Big Sur x86 64-bit
	Catalina      Platform = "catalina"       // macOS Catalina x86 64-bit
	Mojave        Platform = "mojave"         // macOS Mojave x86 64-bit
	HighSierra    Platform = "high_sierra"    // macOS High Sierra x86 64-bit
	X8664Linux    Platform = "x86_64_linux"   // Linux x86 64-bit
	All           Platform = "all"            // All platforms
	Unsupported   Platform = ""               // Unsupported platform (hops defined)
)

// IsValid reports if s is a valid platform.
func IsValid(s string) bool {
	return slices.Contains(ValidPlatformValues, Platform(s))
}

// IsMacOS reports if s is macOS.
func (p Platform) IsMacOS() bool {
	return slices.Contains(MacOSPlatforms, p)
}

// Computed computes the actual included platforms for a Platform value.
func (p Platform) Computed() []Platform {
	switch {
	case !IsValid(p.String()):
		return nil
	case p == All:
		return SupportedPlatforms
	default:
		return []Platform{p}
	}
}

// // Computed computes the actual included platforms for a Platform value.
// func (p Platform) Contains(plat Platform) bool {
// 	switch {
// 	case !IsValid(p.String()):
// 		return nil
// 	case Platform(p) == All:
// 		return SupportedPlatforms
// 	default:
// 		return []Platform{p}
// 	}
// }

// ValidPlatformValues validates flag values.
var ValidPlatformValues = append([]Platform{All}, SupportedPlatforms...)

// SupportedPlatforms contains all supported platforms.
var SupportedPlatforms = []Platform{
	Arm64Sonoma,
	Arm64Ventura,
	Arm64Monterey,
	Arm64BigSur,
	Sonoma,
	Ventura,
	Monterey,
	BigSur,
	Catalina,
	Mojave,
	HighSierra,
	X8664Linux,
}

// MacOSPlatforms contains all known and supported macOS versions.
var MacOSPlatforms = []Platform{
	Arm64Sonoma,
	Arm64Ventura,
	Arm64Monterey,
	Arm64BigSur,
	Sonoma,
	Ventura,
	Monterey,
	BigSur,
	Catalina,
	Mojave,
	HighSierra,
}

// ARM produces the corresponding ARM version of the platform.
func (p Platform) ARM() Platform {
	switch p {
	case Arm64Sonoma:
		return Arm64Sonoma
	case Arm64Ventura:
		return Arm64Ventura
	case Arm64Monterey:
		return Arm64Monterey
	case Arm64BigSur:
		return Arm64BigSur
	case Sonoma:
		return Arm64Sonoma
	case Ventura:
		return Arm64Ventura
	case Monterey:
		return Arm64Monterey
	case BigSur:
		return Arm64BigSur
	case Catalina:
		return Unsupported
	case Mojave:
		return Unsupported
	case HighSierra:
		return Unsupported
	case X8664Linux:
		return Unsupported
	case Unsupported:
		return Unsupported
	case All:
		return All
	default:
		return Unsupported
	}
}

// String implements the fmt.Stringer interface.
func (p Platform) String() string {
	return string(p)
}

// maps darwin versions to Homebrew platform strings.
var darwinVersionToPlatform = map[string]Platform{
	"v17": HighSierra,
	"v18": Mojave,
	"v19": Catalina,
	"v20": BigSur,
	"v21": Monterey,
	"v22": Ventura,
	"v23": Sonoma,
}

var (
	// priority order for arm64 macOS versions.
	orderArm64MacOS = []Platform{
		All,
		Arm64BigSur,
		Arm64Monterey,
		Arm64Ventura,
		Arm64Sonoma,
	}

	// priority order for macOS versions.
	orderAmd64MacOS = []Platform{
		All,
		HighSierra,
		Mojave,
		Catalina,
		BigSur,
		Monterey,
		Ventura,
		Sonoma,
	}

	// priority order for orderAmd64Linux versions.
	orderAmd64Linux = []Platform{
		All,
		X8664Linux,
	}
)

// SelectManifest selects the most viable manifest from an OCI image index.
// The returned index will be -1 if a compatible image is not found.
// The selected platform will not exceed the constraint.
func SelectManifest(index *ocispec.Index, constraint Platform) (ocispec.Descriptor, error) {
	// check for manifest with matching refname
	candidates := make([]Platform, len(index.Manifests))
	for i, manifest := range index.Manifests {
		candidates[i] = FromOCI(manifest.Platform)
	}

	sel := SelectIndex(candidates, constraint)
	if sel < 0 {
		return ocispec.Descriptor{}, errors.New("no manifest in index matches platform " + constraint.String())
	}

	// Return select manifest
	return index.Manifests[sel], nil
}

// SelectManifestIndex selects the most viable manifest from an OCI image index.
// The returned index will be -1 if a compatible image is not found.
// The selected platform will not exceed the constraint.
func SelectManifestIndex(index *ocispec.Index, constraint Platform) int {
	// check for manifest with matching refname
	candidates := make([]Platform, len(index.Manifests))
	for i, manifest := range index.Manifests {
		candidates[i] = FromOCI(manifest.Platform)
	}

	return SelectIndex(candidates, constraint)
}

// SelectIndex selects the most viable platform from a list of candidate platforms.
// The returned index will be -1 if a compatible platform is not found.
// The selected platform will not exceed the constraint.
func SelectIndex(candidates []Platform, constraint Platform) int {
	// Return early if candidates contains the constraint exactly
	if i := slices.Index(candidates, constraint); i != -1 {
		return i
	}

	var max int
	var order []Platform

	m := slices.Index(orderAmd64MacOS, constraint)
	am := slices.Index(orderArm64MacOS, constraint)
	l := slices.Index(orderAmd64Linux, constraint)

	switch {
	// Constrain matches to amd64 macOS versions
	case m != -1:
		order = orderAmd64MacOS
		max = m
	case am != -1:
		order = orderArm64MacOS
		max = am
	case l != -1:
		order = orderAmd64Linux
		max = l
	// Unknown platform, only match "all"
	default:
		order = []Platform{All}
		max = 0
	}

	// Store index of selected platform
	selected := -1
	sorder := -1

	for i, p := range candidates {
		if p == constraint {
			return i // return index if exact match
		}

		corder := slices.Index(order, p)
		switch {
		// candidate platform not compatible with constraint
		case corder == -1:
			continue
		// candidate platform is outside of the constraint, cannot use
		case corder > max:
			continue
		// candidate platform is less satisfactory than current selection
		case corder <= sorder:
			continue
		// candidate platform is more satisfactory than current selection
		case corder > sorder:
			// slog.Log(context.Background(), slog.LevelDebug*2, "selecting new platform candidate",
			// 	slog.String("candidate", p.String()), slog.String("constraint", constraint.String()),
			// 	slog.String("compare", fmt.Sprintf("%d > %d", corder, selected)),
			// )
			selected = i    // select the candidate's index
			sorder = corder // set order comparison to candidate's order
		}
	}
	return selected
}

// SatisfiesOCI checks if the given OCI image platform satisfies the constraint.
// The selected platform will not exceed the constraint.
func SatisfiesOCI(ociplat *ocispec.Platform, constraint Platform) bool {
	return Satisfies(FromOCI(ociplat), constraint)
}

// Satisfies simply checks if the given platform satisfies the constraint.
// The selected platform will not exceed the constraint.
func Satisfies(plat Platform, constraint Platform) bool {
	i := SelectIndex([]Platform{plat}, constraint)
	return i != -1
}
