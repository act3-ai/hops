package formula

import brewfmt "github.com/act3-ai/hops/internal/brew/fmt"

// Version represents a formula version.
type Version interface {
	Upstream() string
	Revision() int
	Rebuild() int
}

// PkgVersion is a helper for producing the package version.
func PkgVersion(v Version) string {
	return brewfmt.PkgVersion(v.Upstream(), v.Revision())
}

// Tag is a helper for producing the bottle tag.
func Tag(v Version) string {
	return brewfmt.Tag(v.Upstream(), v.Revision(), v.Rebuild())
}
