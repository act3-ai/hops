package prefix

import (
	"path/filepath"
)

// Rack represents a rack in the Cellar.
type Rack string

// String implements fmt.Stringer.
func (r Rack) String() string {
	return string(r)
}

// Name implements formula.Namer.
func (r Rack) Name() string {
	return filepath.Base(r.String())
}

// // Kegs returns the kegs available in a rack
// func (r Rack) Kegs() ([]fs.DirEntry, error) {
// }
