package platform

import (
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Prefixes for the OCI platform.os.version field.
const (
	linuxVersion      = "Ubuntu 22.04"
	sonomaVersion     = "macOS 14"
	venturaVersion    = "macOS 13"
	montereyVersion   = "macOS 12"
	bigSurVersion     = "macOS 11"
	catalinaVersion   = "macOS 10.15"
	mojaveVersion     = "macOS 10.14"
	highSierraVersion = "macOS 10.13"
)

// Converts platform to OCI platform.
func FromOCI(r *ocispec.Platform) Platform {
	p := Unsupported

	if r == nil {
		return All
	}

	matchVersion := func(versionPrefix string) bool {
		return matchOSVersion(versionPrefix, r.OSVersion)
	}

	switch r.OS {
	case "linux":
		switch r.Architecture {
		case "amd64":
			switch {
			case matchVersion(linuxVersion):
				p = X8664Linux
			default:
				p = X8664Linux // still give it a shot (Homebrew still installs bottles on Ubuntu 18.04/20.04/etc)
			}
		default:
			p = Unsupported
		}
	case "darwin":
		switch {
		case matchVersion(sonomaVersion):
			p = Sonoma
		case matchVersion(venturaVersion):
			p = Ventura
		case matchVersion(montereyVersion):
			p = Monterey
		case matchVersion(bigSurVersion):
			p = BigSur
		case matchVersion(catalinaVersion):
			p = Catalina
		case matchVersion(mojaveVersion):
			p = Mojave
		case matchVersion(highSierraVersion):
			p = HighSierra
		// Default to Sonoma if OSVersion is empty
		case r.OSVersion == "":
			p = Sonoma
		default:
			p = Unsupported
		}

		switch r.Architecture {
		case "amd64":
		case "arm64":
			p = p.ARM()
		default:
			p = Unsupported
		}
	case "":
		switch r.Architecture {
		case "":
			p = All
		default:
			p = Unsupported
		}
	default:
		p = Unsupported
	}
	return p
}

// ociPlatform converts a Platform to an OCI platform.
func (p Platform) ociPlatform() *ocispec.Platform { //nolint:unused
	var r *ocispec.Platform
	switch p {
	case Arm64Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    sonomaVersion,
		}
	case Arm64Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    venturaVersion,
		}
	case Arm64Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    montereyVersion,
		}
	case Arm64BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    bigSurVersion,
		}
	case Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    sonomaVersion,
		}
	case Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    venturaVersion,
		}
	case Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    montereyVersion,
		}
	case BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    bigSurVersion,
		}
	case Catalina:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    catalinaVersion,
		}
	case Mojave:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    mojaveVersion,
		}
	case HighSierra:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    highSierraVersion,
		}
	case X8664Linux:
		r = &ocispec.Platform{
			OS:           "linux",
			Architecture: "amd64",
			OSVersion:    linuxVersion,
		}
	case All, Unsupported:
		return nil
	}
	return r
}

// matchOSVersion reports if osVersion matches versionPrefix.
func matchOSVersion(versionPrefix, osVersion string) bool {
	after, found := strings.CutPrefix(osVersion, versionPrefix)
	if !found {
		return false // no match
	}

	if after == "" {
		return true // perfect match
	}

	// make sure next character is non-numeric
	// ex: versionPrefix := "macOS 14"
	// 	"macOS 14.4.1" ==> true
	// 	"macOS 14.40"  ==> false
	return !strings.ContainsAny(after[:1], "1234567890")
}
