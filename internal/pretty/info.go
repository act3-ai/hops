package pretty

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/act3-ai/hops/internal/apis/receipt.brew.sh"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/utils"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

// Info prints the formula information in pretty format.
func Info(f formula.PlatformFormula, p prefix.Prefix) {
	// Collect information in advance for readability
	var isInstalled bool
	var files int
	var size int64
	var installReceipt *receipt.InstallReceipt

	var err error

	keg := p.KegPath(f.Name(), formula.PkgVersion(f.Version()))

	isInstalled = p.AnyInstalled(f)
	if isInstalled {
		files, size, err = utils.CountDir(keg)
		if err != nil {
			slog.Warn("checking cellar", logutil.ErrAttr(err))
		}

		installReceipt, err = receipt.Load(keg)
		if err != nil {
			slog.Warn("parsing install receipt", logutil.ErrAttr(err))
		}
	}

	versions := []string{}
	if stable := formula.PkgVersion(f.Version()); stable != "" {
		// TODO:
		// if f.Versions.Bottle {
		stable += " (bottled)"
		// }
		versions = append(versions, stable)
	}
	// TODO?
	// if f.Versions.Head != nil {
	// 	versions = append(versions, *f.Versions.Head)
	// }
	lines := []string{
		fmt.Sprintf(
			"%s: %s",
			o.StyleBold(f.Name()), strings.Join(versions, ", ")),
		f.Info().Desc,
		o.StyleUnderline(f.Info().Homepage),
	}
	// TODO: deprecate/disable
	// if f.Deprecated && f.DeprecationReason != nil {
	// 	lines = append(lines, fmt.Sprintf("Deprecated because it %s!", *f.DeprecationReason))
	// }
	// if f.Disabled {
	// 	lines = append(lines, fmt.Sprintf("Disabled because it %s!", *f.DisabledReason))
	// }

	if conflicts := f.Conflicts(); len(conflicts) > 0 {
		lines = append(lines, "Conflicts with:")
		for _, conflict := range conflicts {
			lines = append(lines, fmt.Sprintf("  %s (because %s)", conflict.Name, conflict.Reason))
		}
	}

	if isInstalled {
		lines = append(lines, fmt.Sprintf(
			"%s (%d files, %s) *",
			keg, files, humanize.Bytes(uint64(size))))

		if installReceipt != nil {
			line := "  "
			if installReceipt.PouredFromBottle {
				line += "Poured from bottle "
			} else {
				line += "Installed "
			}

			if installReceipt.LoadedFromAPI {
				line += "using the formulae.brew.sh API "
			}

			t := time.Unix(int64(installReceipt.Time), 0).Local()
			line += fmt.Sprintf("on %s at %s", t.Format(time.DateOnly), t.Format(time.TimeOnly))
			lines = append(lines, line)
		} else {
			lines = append(lines, "  Missing install receipt")
		}
	} else {
		lines = append(lines, "Not installed")
	}

	lines = append(lines,
		// "From: "+o.StyleUnderline(TapNameToURL(f.Tap)),
		"License: "+f.Info().License)
	o.H1(strings.Join(lines, "\n"))

	// platinfo, err := f.ForPlatform(plat)
	Deps(f, p)
	// if err != nil {
	// 	slog.Warn("evaluating platform metadata", o.ErrAttr(err))
	// } else {
	// }
	if caveats := Caveats(f, p); caveats != "" {
		o.Hai("Caveats\n" + caveats)
	}

	// Analytics are not implemented yet
}

// Deps prints dependency information.
func Deps(f formula.PlatformFormula, p prefix.Prefix) {
	deps := f.Dependencies()
	lines := []string{}
	if len(deps.Build) > 0 {
		lines = append(lines, "Build: "+formatDependencyList(deps.Build, p))
	}
	if len(deps.Required) > 0 {
		lines = append(lines, "Required: "+formatDependencyList(deps.Required, p))
	}
	if len(deps.Test) > 0 {
		lines = append(lines, "Test: "+formatDependencyList(deps.Test, p))
	}
	if len(deps.Recommended) > 0 {
		lines = append(lines, "Recommended: "+formatDependencyList(deps.Recommended, p))
	}
	if len(deps.Optional) > 0 {
		lines = append(lines, "Optional: "+formatDependencyList(deps.Optional, p))
	}
	o.Hai("Dependencies\n" + strings.Join(lines, "\n"))
}

func formatDependencyList(deps []string, p prefix.Prefix) string {
	// Print dependencies
	depnames := []string{}
	for _, dep := range deps {
		depnames = append(depnames, FormulaName(dep, p))
	}
	return strings.Join(depnames, ", ")
}

// FormulaName formats a formula's name based on its installed status.
func FormulaName(name string, p prefix.Prefix) string {
	prefixes, err := p.InstalledKegsByName(name)
	switch {
	case err != nil:
		o.Poo("Checking prefixes: " + err.Error())
		return o.PrettyUninstalled(name)
	case len(prefixes) == 0:
		return o.PrettyUninstalled(name)
	default:
		return o.PrettyInstalled(name)
	}
}

// TapNameToURL converts a tap name to its repository URL using Homebrew's tap naming shortcut.
func TapNameToURL(name string) string {
	pieces := strings.SplitN(name, "/", 2)
	if len(pieces) != 2 {
		return name
	}
	return "https://github.com/" + pieces[0] + "/homebrew-" + pieces[1]
}
