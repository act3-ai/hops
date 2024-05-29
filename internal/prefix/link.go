package prefix

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/utils/symlink"
)

// LinkedFiles finds all files in the prefix that link into the given kegs.
func (p Prefix) LinkedFiles(kegs ...string) ([]string, error) {
	linked := []string{}
	for _, pdir := range p.MustExistSubdirectories() {
		err := fs.WalkDir(os.DirFS(pdir), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err // means pdir does not exist
			}

			if d.Type() != fs.ModeSymlink {
				// Ignore non-symlinks
				return nil
			}

			dst, err := os.Readlink(filepath.Join(pdir, path))
			if err != nil {
				return err
			}

			// Check if link points to one of the given kegs
			for _, keg := range kegs {
				if strings.HasPrefix(filepath.Clean(filepath.Join(pdir, dst)), keg) {
					linked = append(linked, filepath.Join(pdir, path))
				}
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("checking %s for symlinks: %w", pdir, err)
		}
	}
	return linked, nil
}

// BrokenLinks finds all broken links in the prefix.
func (p Prefix) BrokenLinks() ([]string, error) {
	broken := []string{}
	for _, pdir := range p.MustExistSubdirectories() {
		err := fs.WalkDir(os.DirFS(pdir), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err // means pdir does not exist
			}

			if d.Type() != fs.ModeSymlink {
				// Ignore non-symlinks
				return nil
			}

			link := filepath.Join(pdir, path)

			// Check if link destination exists
			_, err = os.Stat(link)
			switch {
			case errors.Is(err, fs.ErrNotExist):
				broken = append(broken, link)
				return nil
			case err != nil:
				return err
			default:
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		})
		if err != nil {
			return nil, fmt.Errorf("checking %s for broken symlinks: %w", pdir, err)
		}
	}
	return broken, nil
}

// LinkOptions configure the link step.
type LinkOptions symlink.Options

// FormulaLink links a keg from the Cellar into the prefix.
func (p Prefix) FormulaLink(f formula.Formula, opts *LinkOptions) (links, files int, err error) {
	return p.Link(f.Name(), formula.PkgVersion(f.Version()), opts)
}

// Link links a keg from the Cellar into the prefix.
//
// https://github.com/Homebrew/brew/blob/ea2892f8ee58494623fc3c15cc8ce81b9124e6e6/Library/Homebrew/keg.rb
//
// TODO: make sure etc and var directories are installed correctly:
// https://github.com/Homebrew/brew/blob/caff1359de1ae7ac198fa7081d905d2a535af3a1/Library/Homebrew/formula.rb#L1339
func (p Prefix) Link(name, version string, opts *LinkOptions) (links, files int, err error) {
	if !opts.DryRun {
		var err error
		// Create all "must exist" dirs
		for _, dir := range p.MustExistDirectories() {
			err = errors.Join(err, os.MkdirAll(dir, 0o775))
		}
		if err != nil {
			return links, files, err
		}
	}

	err = p.OptLink(name, version, (*symlink.Options)(opts))
	if err != nil {
		return links, files, err
	}

	kegPath := p.KegPath(name, version)

	// mapper := iter.Mapper[linkedDirectory, int]{MaxGoroutines: runtime.NumCPU()}
	// linkcounts, err := mapper.MapErr(linkedDirectories, func(ld *linkedDirectory) (int, error) {
	// 	return p.linkDir(kegPath, ld.path, ld.modeFunc, opts)
	// })
	// linksum := 0
	// for _, c := range linkcounts {
	// 	linksum += c
	// }
	// if err != nil {
	// 	return linksum, 0, err
	// }

	for _, ld := range linkedDirectories {
		l, err := p.linkDir(kegPath, ld.path, ld.modeFunc, (*symlink.Options)(opts))
		if err != nil {
			return links, files, err
		}
		links += l
	}

	err = symlink.Relative(kegPath, filepath.Join(p.LinkedKegRecords(), name), (*symlink.Options)(opts))
	if err != nil {
		return links, files, err
	}

	return links, files, nil
}

func matches(pattern, name string) bool {
	result, err := doublestar.PathMatch(pattern, name)
	if err != nil {
		panic(err)
	}
	return result
}

type mode string

var (
	modeMkpath   mode = "mkpath"
	modeSkipDir  mode = "skip_dir"
	modeSkipFile mode = "skip_file"
	modeLink     mode = "link"
	modeInfo     mode = "info"
)

type modeFunc func(path string, d fs.DirEntry) mode

var (
	alwaysMkpath modeFunc = func(_ string, _ fs.DirEntry) mode {
		return modeMkpath
	}

	alwaysSkipDir modeFunc = func(_ string, _ fs.DirEntry) mode {
		return modeSkipDir
	}

	alwaysLink modeFunc = func(_ string, _ fs.DirEntry) mode {
		return modeLink
	}

	// shareModeFunc represents Homebrew's rules when linking the directory.
	shareModeFunc modeFunc = func(path string, _ fs.DirEntry) mode {
		// Iterate over these, it's easier
		matchKegSharePaths := func(path string) bool {
			for _, p := range KegSharePaths {
				if matches("share/"+p, path) {
					return true
				}
			}
			return false
		}

		switch {
		case matches("share/info/*.{info,dir}", path):
			return modeInfo
		case matches("share/icons/**/icon-theme.cache", path):
			return modeSkipFile
		case
			matches("share/{locale,man}/{??,C,POSIX}{_??,}{.*,}{@?*,}", path),
			matches("share/icons/**", path),
			matches("share/zsh/**", path),
			matches("share/fish/**", path),
			matches("share/lua/**", path),
			matches("share/guile/**", path),
			matchKegSharePaths(path):
			return modeMkpath
		default:
			return modeLink
		}
	}

	// libModeFunc represents Homebrew's rules when linking the directory.
	libModeFunc modeFunc = func(path string, _ fs.DirEntry) mode {
		switch {
		case matches("lib/charset.alias", path):
			return modeSkipFile
		case
			matches("lib/pkgconfig", path),
			matches("lib/cmake", path),
			matches("lib/dtrace", path),
			matches("lib/gdk-pixbuf/**", path),
			matches("lib/ghc", path),
			matches("lib/gio/**", path),
			matches("lib/lua/**", path),
			matches("lib/mecab/**", path),
			matches("lib/node/**", path),
			matches("lib/ocaml/**", path),
			matches("lib/perl5/**", path),
			matches("lib/php", path),
			matches("lib/python{2,3}.d/**", path),
			matches("lib/R/**", path),
			matches("lib/ruby/**", path):
			return modeMkpath
		default:
			return modeLink
		}
	}

	// frameworksModeFunc represents Homebrew's rules when linking the directory.
	frameworksModeFunc modeFunc = func(path string, _ fs.DirEntry) mode {
		switch {
		case
			matches("Frameworks/*.framework", path),
			matches("Frameworks/*.framework/Versions", path):
			return modeMkpath
		default:
			return modeLink
		}
	}
)

type linkedDirectory struct {
	path     string
	modeFunc modeFunc
}

var linkedDirectories = []linkedDirectory{
	{
		path:     "etc",
		modeFunc: alwaysMkpath,
	},
	{
		path:     "bin",
		modeFunc: alwaysSkipDir,
	},
	{
		path:     "sbin",
		modeFunc: alwaysSkipDir,
	},
	{
		path:     "include",
		modeFunc: alwaysLink,
	},
	{
		path:     "share",
		modeFunc: shareModeFunc,
	},
	{
		path:     "lib",
		modeFunc: libModeFunc,
	},
	{
		path:     "Frameworks",
		modeFunc: frameworksModeFunc,
	},
}

// linkDir walks all files and directories under the root directory within the keg.
// modeFunc is called for each entry to tells linkDir whether to create the entry, link to it, or skip it.
func (p Prefix) linkDir(kegPath string, root string, modeFunc func(path string, d fs.DirEntry) mode, opts *symlink.Options) (int, error) {
	links := 0
	return links, fs.WalkDir(os.DirFS(kegPath), root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// Skip non-existent directories
			return nil
		case err != nil:
			// Return other stat errors
			return fmt.Errorf("link %q: %w", path, err)
		}

		src := filepath.Join(kegPath, path)
		dst := filepath.Join(string(p), path)

		switch {
		case d.Type() == fs.ModeSymlink || d.Type().IsRegular():
			if filepath.Base(path) == ".DS_Store" ||
				isPycFile(path) {
				return nil
			}

			linked, err := linkFile(modeFunc(path, d), src, dst, opts)
			if err != nil {
				return err
			}

			if linked {
				links++
			}

			return nil
		case d.IsDir():
			// check if dst exists and is not a symlink
			if dstInfo, _ := os.Lstat(dst); dstInfo != nil &&
				dstInfo.IsDir() && dstInfo.Mode().Type() != os.ModeSymlink {
				// Continue walking its tree
				return nil
			}

			// no need to put .app bundles in the path, the user can just use
			// spotlight, or the open command and actual mac apps use an equivalent
			if filepath.Ext(path) == ".app" {
				return fs.SkipDir
			}

			linked, err := linkDir(modeFunc(path, d), src, dst, opts)
			if err != nil {
				return err
			}

			if linked {
				links++
			}

			return nil
		default:
			slog.Debug("skipping unsupported keg file type",
				slog.String("type", d.Type().String()))
			return nil
		}
	})
}

// handles file linking based on mode m.
func linkFile(m mode, src, dst string, opts *symlink.Options) (bool, error) {
	switch m {
	case modeSkipFile:
		return false, nil
	case modeInfo:
		if filepath.Base(src) == "dir" {
			return false, nil
		}

		// make relative symlink
		err := symlink.Relative(src, dst, opts)
		if err != nil {
			return false, err
		}

		// call "dst.install_info"
		return true, nil
	default:
		// make relative symlink
		return true, symlink.Relative(src, dst, opts)
	}
}

// handles directory linking based on mode m.
func linkDir(m mode, src, dst string, opts *symlink.Options) (bool, error) {
	switch m {
	case modeSkipDir:
		// skip this entire tree
		return false, fs.SkipDir
	case modeMkpath:
		// check what "resolve_any_conflicts" does
		err := os.MkdirAll(dst, 0o775)
		if err != nil {
			return false, err
		}

		// continue walking this tree
		return false, nil
	default:
		// make relative symlink
		err := symlink.Relative(src, dst, opts)
		if err != nil {
			return false, err
		}

		// once a symlink has been made, the rest of this tree does not need walked
		return true, fs.SkipDir
	}
}
