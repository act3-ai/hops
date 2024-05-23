package keg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Keg represents a keg as its path.
type Keg string

// String implement fmt.Stringer.
func (k Keg) String() string {
	return string(k)
}

// Name produces the name of the formula.
func (k Keg) Name() string {
	dir, _ := filepath.Split(k.String())
	return filepath.Base(dir)
}

// Version produces the version of the keg.
func (k Keg) Version() string {
	return filepath.Base(k.String())
}

// FS opens an fs.FS in the keg.
func (k Keg) FS() fs.FS {
	return os.DirFS(k.String())
}

// Paths returns a list of paths in the keg.
func (k Keg) Paths() ([]string, error) {
	paths := []string{}
	err := fs.WalkDir(k.FS(), ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, filepath.Join(k.String(), path))
		return nil
	})
	if err != nil {
		return paths, fmt.Errorf("walking keg paths: %w", err)
	}
	return paths, nil
}
