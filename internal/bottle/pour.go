package bottle

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PourFile pours the bottle downloaded by the store into the cellar
func PourFile(path string, b *Bottle, cellar string) error {
	// The path the bottle will be unzipped to
	kegPath := filepath.Join(cellar, b.KegPath())
	err := os.RemoveAll(kegPath)
	if err != nil {
		return fmt.Errorf("removing old keg: %w", err)
	}

	// Open the bottle file
	bottleFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening bottle: %w", err)
	}
	defer bottleFile.Close()

	// Untar the bottle
	err = untar(bottleFile, cellar)
	if err != nil {
		return fmt.Errorf("pouring bottle: %w", err)
	}

	return nil
}

// PourReader pours the bottle into the cellar
func PourReader(b io.Reader, cellar string) error {
	// Untar the bottle
	if err := untar(b, cellar); err != nil {
		return fmt.Errorf("pouring bottle: %w", err)
	}
	return nil
}

// CompatibleWithCellar reports if the bottle is compatible with the given cellar
func (b *Bottle) CompatibleWithCellar(cellar string) bool {
	return b.File.Relocatable() || string(b.File.Cellar) == cellar
}

// untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
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
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		case tar.TypeDir:
			// if its a dir and it doesn't exist create it
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o775); err != nil {
					return err
				}
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
			// if it's a symlink and it doesn't exist create it
			if _, err := os.Stat(target); err != nil {
				err = os.Symlink(header.Linkname, target)
				if err != nil {
					return fmt.Errorf("creating symlink: %w", err)
				}
			}
		}
	}
}
