package platform

import (
	"strings"
	"unicode"
	"unicode/utf8"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var macSymbols = map[Platform]string{
	Sequoia:    "15",
	Sonoma:     "14",
	Ventura:    "13",
	Monterey:   "12",
	BigSur:     "11",
	Catalina:   "10.15",
	Mojave:     "10.14",
	HighSierra: "10.13",
	Sierra:     "10.12",
	ElCapitan:  "10.11",
}

// Prefixes for the OCI platform.os.version field.
const (
	linuxVersion = "Ubuntu 22.04"
)

// Converts OCI platform object to Homebrew platform.
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
		// Default to Seqoia if OSVersion is empty.
		if r.OSVersion == "" {
			p = Sequoia
		} else {
			// Iterate over all Mac platforms
			found := false
			for _, checkPlatform := range MacOSPlatforms {
				if version, ok := macSymbols[checkPlatform]; ok {
					if matchVersion("macOS " + version) {
						p = checkPlatform
						found = true
						break
					}
				}
			}
			if !found {
				p = Unsupported
			}
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
	case Arm64Sequoia:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    macSymbols[Sequoia],
		}
	case Arm64Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    macSymbols[Sonoma],
		}
	case Arm64Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    macSymbols[Ventura],
		}
	case Arm64Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    macSymbols[Monterey],
		}
	case Arm64BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			OSVersion:    macSymbols[BigSur],
		}
	case Sequoia:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Sequoia],
		}
	case Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Sonoma],
		}
	case Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Ventura],
		}
	case Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Monterey],
		}
	case BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[BigSur],
		}
	case Catalina:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Catalina],
		}
	case Mojave:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Mojave],
		}
	case HighSierra:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[HighSierra],
		}
	case Sierra:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[Sierra],
		}
	case ElCapitan:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			OSVersion:    macSymbols[ElCapitan],
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

	nextRune, _ := utf8.DecodeRuneInString(after)
	return !unicode.IsDigit(nextRune)
	// return !strings.ContainsAny(after[:1], "1234567890")
}
