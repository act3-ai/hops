package formula

import brewfmt "github.com/act3-ai/hops/internal/brew/fmt"

// Version represents a formula version.
type Version interface {
	Upstream() string
	Revision() int
	Rebuild() int
}

// PkgVersion is a helper for producing the package version.
func PkgVersion(obj Versioner) string {
	v := obj.Version()
	return brewfmt.PkgVersion(v.Upstream(), v.Revision())
}

// Tag is a helper for producing the bottle tag.
func Tag(obj Versioner) string {
	v := obj.Version()
	return brewfmt.Tag(v.Upstream(), v.Revision(), v.Rebuild())
}

// func ParseVersion() (version string, packageRevision, bottleRebuild int)
