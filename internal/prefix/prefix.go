package prefix

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/sourcegraph/conc/iter"
	"golang.org/x/mod/semver"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/utils"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/symlink"
)

// NewErrNoSuchKeg creates an error when a keg is not found.
func (p Prefix) NewErrNoSuchKeg(name string) error {
	return fmt.Errorf("no such keg: %s", filepath.Join(string(p), name))
}

// Prefix represents a Homebrew prefix.
type Prefix string

// String implements fmt.Stringer.
func (p Prefix) String() string {
	return string(p)
}

const (
	DarwinArm64Default Prefix = "/opt/homebrew"              // the default prefix on ARM64 macOS
	DarwinAmd64Default Prefix = "/usr/local"                 // the default prefix on x86 macOS
	LinuxDefault       Prefix = "/home/linuxbrew/.linuxbrew" // the default prefix on linux
)

// Default returns the default Homebrew prefix value.
func Default() Prefix {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return DarwinArm64Default
		}
		return DarwinAmd64Default
	case "linux":
		return LinuxDefault
	default:
		// Other platform just return Linux, it will error out elsewhere if needed
		return LinuxDefault
	}
}

// EnvOrDefault returns the Homebrew prefix value
// set by the HOMEBREW_PREFIX environment variable
// or the default Homebrew prefix value.
func EnvOrDefault() Prefix {
	// Allow environment variable to override
	prefix, ok := os.LookupEnv("HOMEBREW_PREFIX")
	if ok && prefix != "" {
		return Prefix(prefix)
	}
	return Default()
}

// Cellar.
func (p Prefix) Cellar() string {
	return filepath.Join(string(p), "Cellar")
}

// CompatibleWithCellar reports whether the Formula's Bottle is compatible with the Prefix's Cellar.
func (p Prefix) CompatibleWithCellar(f formula.PlatformFormula) error {
	want := ""
	if btl := f.Bottle(); btl != nil {
		want = btl.Cellar
	}
	got := p.Cellar()

	switch want {
	// Empty string means Bottle is relocatable (or cannot be checked for compatibility).
	case "":
		return nil
	// Compatible with configured Cellar.
	case got:
		return nil
	// Cannot be relocated and is not compatible with configured Cellar.
	default:
		return errors.New(heredoc.Docf(`
			bottle for %s may be incompatible with your settings
			  HOMEBREW_CELLAR: %s (yours is %s)
			  HOMEBREW_PREFIX: %s (yours is %s)`,
			f.Name(),
			want, got,
			filepath.Dir(want), p.String(),
		))
	}
}

// KegPath produces the keg for the given name and version.
func (p Prefix) KegPath(name, version string) string {
	return filepath.Join(p.Cellar(), name, version)
}

// FormulaKegPath produces the keg for a formula.
func (p Prefix) FormulaKegPath(f formula.Formula) string {
	return p.KegPath(f.Name(), formula.PkgVersion(f))
}

// Opt.
func (p Prefix) Opt() string {
	return filepath.Join(string(p), "opt")
}

// OptRecord.
func (p Prefix) OptRecord(name string) string {
	return filepath.Join(string(p), "opt", name)
}

// Library.
func (p Prefix) Library() string {
	return filepath.Join(string(p), "Library")
}

// ShimsPath.
func (p Prefix) ShimsPath() string {
	return filepath.Join(p.Library(), "Homebrew", "shims")
}

// DataPath.
func (p Prefix) DataPath() string {
	return filepath.Join(p.Library(), "Homebrew", "data")
}

// LinkedKegRecords.
func (p Prefix) LinkedKegRecords() string {
	return filepath.Join(string(p), "var", "homebrew", "linked")
}

// PinnedKegRecords.
func (p Prefix) PinnedKegRecords() string {
	return filepath.Join(string(p), "var", "homebrew", "pinned")
}

// Locks.
func (p Prefix) Locks() string {
	return filepath.Join(string(p), "var", "homebrew", "locks")
}

// OptLink.
func (p Prefix) OptLink(name, version string, opts *symlink.Options) error {
	optRecord := p.OptRecord(name)

	kegPath := p.KegPath(name, version)

	err := symlink.Relative(kegPath, optRecord, opts)
	if err != nil {
		return err
	}

	// TODO
	// // Refresh all alias links
	// for _, alias := range info.Aliases {
	// 	aliasOptRecord := p.OptRecord(alias)
	// 	// Create new link from alias to the keg
	// 	err := symlink.Relative(kegPath, aliasOptRecord, opts)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// // Refresh all old name links
	// for _, oldName := range info.OldNames {
	// 	oldNameOptRecord := p.OptRecord(oldName)
	// 	// Create new link from old name to the keg
	// 	err := symlink.Relative(kegPath, oldNameOptRecord, opts)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// KegKegLinkDirectories.
func KegKegLinkDirectories() []string {
	return []string{
		"bin", "etc", "include", "lib", "sbin", "share", "var",
	}
}

var (
	// KegLinkDirectories.
	KegLinkDirectories = []string{
		"bin", "etc", "include", "lib", "sbin", "share", "var",
	}

	// MustExistSubdirectories.
	MustExistSubdirectories = []string{
		"bin", "etc", "include", "lib", "sbin", "share",
	}

	// KegSharePaths returns the share paths of a keg (Keg:SHARE_PATHS).
	// These paths relative to the keg's share directory should always be real directories in the prefix, never symlinks.
	KegSharePaths = []string{
		"aclocal", "doc", "info", "java", "locale", "man",
		"man/man1", "man/man2", "man/man3", "man/man4",
		"man/man5", "man/man6", "man/man7", "man/man8",
		"man/cat1", "man/cat2", "man/cat3", "man/cat4",
		"man/cat5", "man/cat6", "man/cat7", "man/cat8",
		"applications", "gnome", "gnome/help", "icons",
		"mime-info", "pixmaps", "sounds", "postgresql",
	}
)

// MustExistSubdirectories.
func (p Prefix) MustExistSubdirectories() []string {
	return []string{
		filepath.Join(string(p), "bin"),
		filepath.Join(string(p), "etc"),
		filepath.Join(string(p), "include"),
		filepath.Join(string(p), "lib"),
		filepath.Join(string(p), "sbin"),
		filepath.Join(string(p), "share"),
		filepath.Join(string(p), "opt"),
		p.LinkedKegRecords(),
	}
}

// MustExistDirectories.
func (p Prefix) MustExistDirectories() []string {
	return append(
		p.MustExistSubdirectories(),
		p.Cellar(),
	)
}

// file extensions for elisp files.
// elispExtensions = []string{".el", ".elc"}

// file extensions for pyc files.
var pycExtensions = []string{".pyc", ".pyo"}

func isPycFile(path string) bool {
	// Don't link pyc or pyo files because Python overwrites these
	// cached object files and next time brew wants to link, the
	// file is in the way.
	return slices.Contains(pycExtensions, filepath.Ext(path)) && strings.Contains(path, "/site-packages/")
}

// file extensions for libtool files.
// libtoolExtensions = []string{".la", ".lai"}

// AnyInstalled reports if any versions of the given formula are installed.
func (p Prefix) AnyInstalled(f formula.Formula) bool {
	prefix, err := p.InstalledKegs(f)
	if err != nil {
		slog.Warn("checking installed prefixes", logutil.ErrAttr(err))
	}
	return len(prefix) > 0
}

// InstalledKegs returns all currently installed prefix directories.
func (p Prefix) InstalledKegs(f formula.Formula) ([]Keg, error) {
	return p.InstalledKegsByName(f.Name())
}

// InstalledPrefixes returns all currently installed prefix directories.
func (p Prefix) InstalledKegsByName(names ...string) ([]Keg, error) {
	prefixes := []struct {
		dir   string
		entry fs.DirEntry
	}{}

	sortedPrefixes := func() []Keg {
		// Sort by basenames
		sort.Slice(prefixes, func(i, j int) bool {
			return prefixes[i].entry.Name() < prefixes[j].entry.Name()
		})

		// Flatten to list of keg dirs
		sorted := []Keg{}
		for _, p := range prefixes {
			sorted = append(sorted, Keg(filepath.Join(p.dir, p.entry.Name())))
		}

		return sorted
	}

	for _, name := range names {
		dir := filepath.Join(p.Cellar(), name)
		entries, err := os.ReadDir(dir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return sortedPrefixes(), err
		}

		for _, entry := range entries {
			prefixes = append(prefixes, struct {
				dir   string
				entry fs.DirEntry
			}{
				dir:   dir,
				entry: entry,
			})
		}
	}

	return sortedPrefixes(), nil
}

// FormulaOutdated reports whether the formula is outdated.
func (p Prefix) FormulaOutdated(f formula.Formula) (bool, error) {
	installedPrefixes, err := p.InstalledKegs(f)
	if err != nil {
		return true, err
	}

	latest := formula.PkgVersion(f)

	outdated := true

	// Try to find an up-to-date keg
	for _, installedPrefix := range installedPrefixes {
		installedVersion := filepath.Base(string(installedPrefix))

		l := slog.Default().With(slog.String("latest", latest), slog.String("installed", installedVersion))

		// Check if the installed version is newer or up-to-date
		switch semver.Compare(latest, installedVersion) {
		case -1:
			l.Debug("Found keg with greater version than found version")
			outdated = false
		case 0:
			outdated = false
		default:
		}
	}

	return outdated, nil
}

// FilterInstalledGeneric categorizes Formulae by install status.
func FilterInstalled[T formula.Formula](p Prefix, list []T) (uninstalled []T, installed []T, err error) {
	for _, entry := range list {
		missing, err := p.FormulaOutdated(entry)
		switch {
		case err != nil:
			return nil, nil, err
		case missing:
			uninstalled = append(uninstalled, entry)
		default:
			installed = append(installed, entry)
		}
	}
	return uninstalled, installed, nil
}

// FormulaOutdated reports whether the formula is outdated.
func (p Prefix) FormulaOutdatedFromName(name, latest string) (bool, error) {
	installedPrefixes, err := p.InstalledKegsByName(name)
	if err != nil {
		return true, err
	}

	outdated := true

	// Try to find an up-to-date keg
	for _, installedPrefix := range installedPrefixes {
		installedVersion := filepath.Base(string(installedPrefix))

		l := slog.Default().With(slog.String("keg", installedVersion), slog.String("latest", latest))

		// Check if the installed version is newer or up-to-date
		switch semver.Compare(latest, installedVersion) {
		case -1:
			l.Debug("found keg with newer version, reverting")
		case 0:
			l.Debug("found up to date keg")
			outdated = false
		default:
			l.Debug("found out of date keg")
		}
	}

	return outdated, nil
}

// Uninstall removes the keg and any symlinks into the keg.
func (p Prefix) Uninstall(kegs ...string) error {
	links, err := p.LinkedFiles(kegs...)
	if err != nil {
		return err
	}

	for _, l := range links {
		// Remove the link
		err = os.Remove(l)
		if err != nil {
			return err
		}
	}

	for _, keg := range kegs {
		files, size, err := utils.CountDir(keg)
		if err != nil {
			return err
		}
		fmt.Printf("Uninstalling %s... (%d files, %s)\n", keg, files, utils.PrettyBytes(size))

		// Remove the keg
		err = os.RemoveAll(keg)
		if err != nil {
			return fmt.Errorf("uninstalling %s: %w", keg, err)
		}
	}

	return nil
}

// Racks returns the list of available racks.
func (p Prefix) Racks() ([]Rack, error) {
	racks := []Rack{}

	err := p.forEachRack(
		func(r fs.DirEntry, _ []fs.DirEntry) {
			racks = append(racks, Rack(filepath.Join(p.Cellar(), r.Name())))
		})
	if err != nil {
		return nil, err
	}

	return racks, nil
}

// forEachRack iterates over each rack.
func (p Prefix) forEachRack(fn func(rack fs.DirEntry, kegs []fs.DirEntry)) error {
	racks, err := os.ReadDir(p.Cellar())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	mapper := iter.Mapper[fs.DirEntry, []fs.DirEntry]{}

	rackkegs, err := mapper.MapErr(racks, func(ep *fs.DirEntry) ([]fs.DirEntry, error,
	) {
		e := *ep

		if e.Type() == os.ModeSymlink {
			// continue // filter out symlinks
			return nil, nil
		}

		if strings.HasPrefix(e.Name(), ".") {
			// continue // filter out hidden files and dirs
			return nil, nil
		}

		rack := filepath.Join(p.Cellar(), e.Name())
		kegs, err := os.ReadDir(rack)
		if err != nil {
			return nil, fmt.Errorf("checking for kegs in %s: %w", rack, err)
		}

		if len(kegs) == 0 {
			return nil, nil // skip empty racks
		}

		return kegs, nil
	})
	if err != nil {
		return err
	}

	for i, kegs := range rackkegs {
		if len(kegs) == 0 {
			continue // skip empty racks
		}

		fn(racks[i], kegs)
	}

	// for _, e := range racks {
	// 	if e.Type() == os.ModeSymlink {
	// 		continue // filter out symlinks
	// 	}

	// 	if strings.HasPrefix(e.Name(), ".") {
	// 		continue // filter out hidden files and dirs
	// 	}

	// 	rack := filepath.Join(p.Cellar(), e.Name())
	// 	kegs, err := os.ReadDir(rack)
	// 	if err != nil {
	// 		return fmt.Errorf("checking for kegs in %s: %w", rack, err)
	// 	}

	// 	if len(kegs) == 0 {
	// 		continue // skip empty racks
	// 	}

	// 	fn(e, kegs)
	// }

	return nil
}

// Kegs returns the list of available kegs.
func (p Prefix) Kegs() ([]Keg, error) {
	ks := []Keg{}

	err := p.forEachRack(func(rack fs.DirEntry, kegs []fs.DirEntry) {
		for _, k := range kegs {
			ks = append(ks, Keg(filepath.Join(p.Cellar(), rack.Name(), k.Name())))
		}
	})
	if err != nil {
		return nil, err
	}

	return ks, nil
}
