package platform

import (
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	LinuxVersion      = "Ubuntu 22.04"
	SonomaVersion     = "macOS 14"
	VenturaVersion    = "macOS 13"
	MontereyVersion   = "macOS 12"
	BigSurVersion     = "macOS 11"
	CatalinaVersion   = "macOS 10.15"
	MojaveVersion     = "macOS 10.14"
	HighSierraVersion = "macOS 10.13"
)

// Converts platform to OCI platform
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
			case matchVersion(LinuxVersion):
				p = X8664Linux
			default:
				// slog.Warn("Unsupported Linux OS version",
				// 	slog.String("want", LinuxVersion),
				// 	slog.String("got", r.OSVersion))
				p = X8664Linux // still give it a shot (Homebrew still installs bottles on Ubuntu 18.04/20.04/etc)
			}
		default:
			p = Unsupported
		}
	case "darwin":
		switch {
		case matchVersion(SonomaVersion):
			p = Sonoma
		case matchVersion(VenturaVersion):
			p = Ventura
		case matchVersion(MontereyVersion):
			p = Monterey
		case matchVersion(BigSurVersion):
			p = BigSur
		case matchVersion(CatalinaVersion):
			p = Catalina
		case matchVersion(MojaveVersion):
			p = Mojave
		case matchVersion(HighSierraVersion):
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

// Converts platform to OCI platform
func (p Platform) OCIPlatform() *ocispec.Platform {
	var r *ocispec.Platform
	switch p {
	case Arm64Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			// OSVersion:    "",
		}
	case Arm64Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			// OSVersion:    "",
		}
	case Arm64Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			// OSVersion:    "",
		}
	case Arm64BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "arm64",
			// OSVersion:    "",
		}
	case Sonoma:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case Ventura:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case Monterey:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case BigSur:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case Catalina:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case Mojave:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case HighSierra:
		r = &ocispec.Platform{
			OS:           "darwin",
			Architecture: "amd64",
			// OSVersion:    "",
		}
	case X8664Linux:
		r = &ocispec.Platform{
			OS:           "linux",
			Architecture: "amd64",
		}
	case All, Unsupported:
		return nil
	}
	return r
}

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
