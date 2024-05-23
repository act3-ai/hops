package platform

import (
	"log/slog"
	"strings"

	"github.com/elastic/go-sysinfo"
	"golang.org/x/mod/semver"
)

// fromDarwinVersion maps versions of darwin to amd64 platforms.
func fromDarwinVersion(darwinVersion string) Platform {
	darwinVersion = "v" + strings.TrimPrefix(darwinVersion, "v")
	major := semver.Major(darwinVersion)
	return darwinVersionToPlatform[major]
}

// SystemPlatform detects the host's platform.
func SystemPlatform() Platform {
	host, err := sysinfo.Host()
	if err != nil {
		panic(err)
	}

	log := slog.Default()

	hostAttrs := []slog.Attr{slog.String("platform", host.Info().OS.Platform)}
	if host.Info().Architecture != host.Info().NativeArchitecture {
		hostAttrs = append(hostAttrs,
			slog.String("arch(process)", host.Info().Architecture),
			slog.String("arch(native)", host.Info().NativeArchitecture))
	} else {
		hostAttrs = append(hostAttrs, slog.String("arch", host.Info().Architecture))
	}
	if host.Info().OS.Platform == "darwin" {
		hostAttrs = append(hostAttrs, slog.String("kernel", host.Info().KernelVersion))
	}

	log.Debug("Host details", slog.Any("info", slog.GroupValue(hostAttrs...)))

	logUnsupported := func() {
		slog.Warn("unsupported platform", slog.Any("host", slog.GroupValue(hostAttrs...)))
	}

	switch host.Info().OS.Platform {
	case "darwin": // macOS

		plat := fromDarwinVersion(host.Info().KernelVersion)

		// Handle ARM64 or x86_64 architectures
		switch host.Info().Architecture {
		case "arm64":
			return plat.ARM()
		case "x86_64":
			return plat
		default:
			logUnsupported()
			return Unsupported
		}
	case "windows":
		logUnsupported()
		return Unsupported
	default:
		if host.Info().Architecture != "x86_64" {
			logUnsupported()
			return Unsupported
		}

		// assume linux for all else
		return X8664Linux
	}
}
