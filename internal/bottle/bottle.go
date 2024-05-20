package bottle

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/opencontainers/go-digest"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/platform"
)

// Store downloads bottles
type Store interface {
	Path(bottle *Bottle) string                         // path to a downloaded bottle
	Source(bottle *Bottle) string                       // source for a bottle
	Download(ctx context.Context, bottle *Bottle) error // downloads a bottle
}

// Bottle represents a bottle
type Bottle struct {
	Name     string             // bottle name
	version  string             // bottle version
	Platform platform.Platform  // bottle platform
	Rebuild  int                // rebuild count
	RootURL  string             // root URL for the bottle
	File     *brewv1.BottleFile // info on the tar.gz file
	Digest   digest.Digest      // Digest of the bottle's file
}

// FromFormula initializes a Bottle from a formula
func FromFormula(f *brewv1.Info, key string, plat platform.Platform) (*Bottle, error) {
	bottle := f.Bottle[key]
	if bottle == nil {
		return nil, fmt.Errorf("formula %s does not have a %s bottle", f.Name, key)
	}

	// Load file for the requested platform
	file := bottle.Files[plat]
	if file == nil {
		// Load cross-platform file as fallback
		file = bottle.Files[platform.All]
		if file == nil {
			return nil, fmt.Errorf("formula %s does not have a bottle for platform %s", f.Name, plat)
		}
	}

	// Pre-parse the digest
	d, err := digest.Parse("sha256:" + file.Sha256)
	if err != nil {
		return nil, fmt.Errorf("parsing bottle digest: %w", err)
	}

	return &Bottle{
		Name:     f.Name,
		version:  f.Version(),
		Platform: plat,
		Rebuild:  bottle.Rebuild,
		RootURL:  bottle.RootURL,
		File:     file,
		Digest:   d,
	}, nil
}

// Manifest loads the first bottle defined in the formula
// Not for installation use
func Manifest(f *brewv1.Info, key string) (string, error) {
	bottle := f.Bottle[key]
	if bottle == nil {
		return "", fmt.Errorf("formula %s does not have a %s bottle", f.Name, key)
	}

	// Stores found manifests
	manifests := []string{}

	// Load each platform to ensure they are in agreement
	// If only one platform was rebuilt, there will be two manifests for the same bottle
	// (that's not how the Homebrew project works, but for sanity...)
	for plat := range bottle.Files {
		b, err := FromFormula(f, key, plat)
		if err != nil {
			return "", err
		}

		m := fmt.Sprintf("%s:%s", b.Repo(), b.Tag())
		if slices.Contains(manifests, m) {
			continue // skip already-added manifests (which means bottles are all in agreement)
		}

		manifests = append(manifests, m)
	}
	if len(manifests) > 1 {
		// Return an error if there are multiple manifests
		return manifests[0], fmt.Errorf("multiple manifests found for bottle %s: %s", f.Name, strings.Join(manifests, ", "))
	}

	return manifests[0], nil
}

// implements a short id for error messages
func (b *Bottle) id() string {
	return b.Name + "@" + b.Tag()
}

// Tag returns the expected tag for the bottle
//
// Pattern: VERSION[_REVISION][-REBUILD]
//
// This tag will vary from the formula version when the "rebuild"
// field is set in the formula's bottle, which signals the bottle has
// been rebuilt and retagged without changing the version
func (b *Bottle) Tag() string {
	return brewfmt.Tag(b.version, 0, b.Rebuild)
}

// Repo produces the repository name for a bottle
//
// See the implementation for the edge cases that make this output different than the formula's name.
//
// Pattern: NAME[/PINNED_VERSION]
func (b *Bottle) Repo() string {
	return brewfmt.Repo(b.Name)
}

// Repo produces the repository name for a bottle
//
// Pinned formulae have their names modified, replacing the "@" with a slash.
//
// Pattern: NAME[/PINNED_VERSION]
func (b *Bottle) KegPath() string {
	return filepath.Join(b.Name, b.version)
}

// ArchiveName returns a short name for the downloaded bottle .tar.gz file for the formula
//
// Pattern: NAME--VERSION[_REVISION][-REBUILD]
//
// Example: cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
func (b *Bottle) ArchiveName() string {
	return brewfmt.ArchiveFile(b.Name, b.version, 0, b.Rebuild, b.Platform)
}

// LinkName returns the name of the symlink to the downloaded bottle .tar.gz file for the formula
//
// Pattern: NAME--VERSION
//
// Example: cowsay--3.04_1
func (b *Bottle) LinkName() string {
	return b.Name + "--" + b.version
}
