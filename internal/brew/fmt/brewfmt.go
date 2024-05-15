package brewfmt

import (
	"fmt"
	"strings"

	"github.com/act3-ai/hops/internal/platform"
)

// Repo produces a repository name from a name
//
// See the implementation for the edge cases that make this output different than the formula's name.
//
// Pattern: NAME[/PINNED_VERSION]
func Repo(name string) string {
	repo := name                              // start with name
	repo = strings.ReplaceAll(repo, "@", "/") // replace "@" with "/"
	repo = strings.ReplaceAll(repo, "+", "x") // replace "+" with "x"
	return repo
}

// Tag produces a bottle tag for the given version information
//
// Pattern: VERSION[_REVISION][-REBUILD]
//
// This tag will vary from the formula version when the "rebuild"
// field is set in the formula's bottle, which signals the bottle has
// been rebuilt and retagged without changing the version
func PkgVersion(version string, revision int) string {
	pkgv := version    // start with version
	if revision != 0 { // append revision number if nonzero
		pkgv += fmt.Sprintf("_%d", revision)
	}
	return pkgv
}

// Tag produces a bottle tag for the given version information
//
// Pattern: VERSION[_REVISION][-REBUILD]
//
// This tag will vary from the formula version when the "rebuild"
// field is set in the formula's bottle, which signals the bottle has
// been rebuilt and retagged without changing the version
func Tag(version string, revision, rebuild int) string {
	tag := PkgVersion(version, revision) // start with pkg version
	if rebuild != 0 {                    // append rebuild number if nonzero
		tag += fmt.Sprintf("-%d", rebuild)
	}
	return tag
}

// ArchiveFile produces the bottle archive filename from the given information
//
// Pattern: NAME--VERSION[_REVISION][-REBUILD]
//
// Example: cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
func ArchiveFile(name, version string, revision, rebuild int, plat platform.Platform) string {
	// Construct main name
	archive := fmt.Sprintf("%s--%s.%s.bottle",
		name,
		PkgVersion(version, revision),
		plat,
	)

	// Add rebuild if non-zero
	if rebuild != 0 {
		archive += fmt.Sprintf(".%d", rebuild)
	}

	// Add extension
	return archive + ".tar.gz"
}
