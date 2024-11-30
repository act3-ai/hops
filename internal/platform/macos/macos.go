// Package macos contains utilities for checking macOS system capabilities.
//
// https://en.wikipedia.org/wiki/MacOS
package macos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/act3-ai/hops/internal/utils"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

// Known macOS versions.
//
// Source: https://en.wikipedia.org/wiki/Template:MacOS_versions
//
//revive:disable:exported
const (
	Unknown      Version = iota - 1 // Unknown version
	Jaguar       Version = iota + 5 // Mac OS X Jaguar
	Panther                         // Mac OS X Panther
	Tiger                           // Mac OS X Tiger
	Leopard                         // Mac OS X Leopard
	SnowLeopard                     // Mac OS X Snow Leopard
	Lion                            // Mac OS X Lion
	MountainLion                    // OS X Mountain Lion
	Mavericks                       // OS X Mavericks
	Yosemite                        // OS X Yosemite
	ElCapitan                       // OS X El Capitan
	Sierra                          // macOS Sierra
	HighSierra                      // macOS High Sierra
	Mojave                          // macOS Mojave
	Catalina                        // macOS Catalina
	BigSur                          // macOS Big Sur
	Monterey                        // macOS Monterey
	Ventura                         // macOS Ventura
	Sonoma                          // macOS Sonoma
	Sequoia                         // macOS Sequoia
	//revive:enable:exported
)

// Version represents a macOS version.
//
// The macOS release is represented by the version of the darwin kernel it shipped with.
type Version int

// Darwin returns the darwin kernel version.
func (v Version) Darwin() int {
	return int(v)
}

// index produces the 0-based index for the version.
// macOS versions in this package start at 6.
func (v Version) index() int {
	return v.Darwin() - 6
}

// Name of each version.
var names = []string{
	"Jaguar",
	"Panther",
	"Tiger",
	"Leopard",
	"Snow Leopard",
	"Lion",
	"Mountain Lion",
	"Mavericks",
	"Yosemite",
	"El Capitan",
	"Sierra",
	"High Sierra",
	"Mojave",
	"Catalina",
	"Big Sur",
	"Monterey",
	"Ventura",
	"Sonoma",
	"Sequoia",
}

// Name produces the short name of the version.
func (v Version) Name() string {
	if n, ok := utils.IndexIfOK(names, v.index()); ok {
		return n
	}
	return "UNKNOWN"
}

// FullName produces the full name of the version.
func (v Version) FullName() string {
	n := v.Name()
	switch {
	case n == "UNKNOWN":
		return n
	case v < MountainLion:
		return "Mac OS X " + n
	case v >= MountainLion && v < Sierra:
		return "OS X " + n
	default:
		return "macOS " + n
	}
}

// OS release versions, which differ from the darwin kernel's version.
var osVersions = []string{
	"10.2", // Jaguar
	"10.3", // Panther
	"10.4", // etc.
	"10.5",
	"10.6",
	"10.7",
	"10.8",
	"10.9",
	"10.10",
	"10.11",
	"10.12",
	"10.13",
	"10.14",
	"10.15",
	"11",
	"12",
	"13",
	"14", // Sonoma
	"15", // Sequoia
}

// OSVersion produces the OS version.
func (v Version) OSVersion() string {
	if n, ok := utils.IndexIfOK(osVersions, v.index()); ok {
		return n
	}
	return "UNKNOWN"
}

// SupportsARM reports whether the version supports ARM applications.
func (v Version) SupportsARM() bool {
	return v >= BigSur
}

// Supports64Bit reports whether the version supports 64-bit applications.
func (v Version) Supports64Bit() bool {
	return v >= Panther
}

// Supports32Bit reports whether the version supports 32-bit applications.
func (v Version) Supports32Bit() bool {
	return v <= Mojave
}

// CLTInstalled reports whether the XCode Command Line Tools are installed.
func CLTInstalled(ctx context.Context) (bool, error) {
	// xcode-select -p >/dev/null;
	cmd := exec.CommandContext(ctx, "xcode-select", "-p")
	o, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug("command failed",
			slog.String("command", cmd.String()),
			slog.String("output", string(o)),
			logutil.ErrAttr(err),
		)
		return false, fmt.Errorf("running: %s\nerror:\n%w", cmd.String(), err)
	}

	return true, nil
}

// CLTInstalled reports whether the XCode Command Line Tools are installed.
func (v Version) CLTInstalledVersion(ctx context.Context) (bool, error) {
	if v <= Mavericks {
		return false, errors.New("cannot check for XCode Command Line Tools on " + v.FullName())
	}
	return CLTInstalled(ctx)
}
