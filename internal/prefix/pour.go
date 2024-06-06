package prefix

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/act3-ai/hops/internal/apis/formulae.brew.sh/common"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform/macos"
)

// CanPourBottle verifies that the PlatformFormula's bottle can be poured.
func (p Prefix) CanPourBottle(ctx context.Context, f formula.PlatformFormula) error {
	switch f.Bottle().PourOnlyIf {
	// Bottle can always be poured
	case "":
		return nil
	// Bottle can be poured with the default prefix
	case common.PourBottleConditionDefaultPrefix:
		if p != Default() {
			return errors.New(
				"cannot pour bottle for " + f.Name() +
					`: incompatible prefix`)
		}
		return nil
	// Bottle can be poured with the XCode Command Line Tools installed
	case common.PourBottleConditionCLTInstalled:
		// skip check on non-macOS systems
		if !f.Platform().IsMacOS() {
			return nil
		}

		ok, err := macos.CLTInstalled(ctx)
		switch {
		// Could not check
		case err != nil:
			return err
		// CLT are not installed
		case !ok:
			return errors.New(
				"cannot pour bottle for " + f.Name() +
					`: XCode Command Line Tools are not installed`)
		// CLT are installed
		default:
			return nil
		}
	// Custom rule, skip the rule while warning
	default:
		slog.Warn(`Skipping unknown "pour_bottle_only_if" condition`,
			slog.String("bottle", f.Name()),
			slog.String("condition", f.Bottle().PourOnlyIf))
		return nil
	}
}

// CanPourBottles verifies that all formulae's bottles can be poured.
func (p Prefix) CanPourBottles(ctx context.Context, formulae []formula.PlatformFormula) error {
	var err error
	for _, f := range formulae {
		err = errors.Join(err, p.CanPourBottle(ctx, f))
	}
	return err
}

// Pour pours a Bottle into the Cellar.
func (p Prefix) Pour(btl io.Reader) error {
	// Untar the bottle
	if err := untar(btl, p.Cellar()); err != nil {
		return fmt.Errorf("pouring bottle: %w", err)
	}
	return nil
}

// untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files.
func untar(r io.Reader, dst string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case errors.Is(err, io.EOF):
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		// cwe-22: validate that the path does not contain ".."
		// already handled by the error returned from tr.Next(), but here for redundancy.
		case !filepath.IsLocal(header.Name):
			return errors.New("archive contains path traversal, cannot write to path " + header.Name)
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		case tar.TypeDir:
			info, err := os.Stat(target)
			switch {
			// path does not exist
			case errors.Is(err, os.ErrNotExist):
				// create dir
				if err := os.MkdirAll(target, 0o775); err != nil {
					return err
				}
			// unknown path error
			case err != nil:
				return fmt.Errorf("checking destination: %w", err)
			// path exists but is a file
			case !info.IsDir():
				return fmt.Errorf("creating directory %s: destination is a file", target)
			// directory exists
			default:
				return nil
			}
		case tar.TypeReg:
			// if it's a file create it
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("copying file from tar archive: %w", err)
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			err = f.Close()
			if err != nil {
				return fmt.Errorf("closing copied file: %w", err)
			}
		case tar.TypeSymlink:
			_, err := os.Stat(target)
			switch {
			// path already exists
			case err == nil:
				return nil
			// unknown path error
			case !errors.Is(err, os.ErrNotExist):
				return fmt.Errorf("checking destination: %w", err)
			// cwe-22: the created symlink can point outside of the archive directory.
			// creating the symlink itself is not the concern, but subsequent files
			// could be created in the symlinked location, overwriting system files.
			// evaluate locality from the symlink file's directory
			case !filepath.IsLocal(filepath.Join(filepath.Dir(header.Name), header.Linkname)):
				return errors.New("archive contains path traversal, cannot create symlink " + target + " -> " + header.Linkname)
			// create symlink
			default:
				err = os.Symlink(header.Linkname, target)
				if err != nil {
					return fmt.Errorf("creating symlink: %w", err)
				}
			}
		}
	}
}
