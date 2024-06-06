package brewformulary

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"

	api "github.com/act3-ai/hops/internal/apis/formulae.brew.sh"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/errdef"
	"github.com/act3-ai/hops/internal/formula"
)

// PreloadedFormulary defines the formulary's capabilities.
type PreloadedFormulary interface {
	formula.Formulary
	ListNames() []string
}

// V1Cache represents formula data cached from the Homebrew API.
type V1Cache struct {
	mapped  map[string]*brewv1.Info // full contents indexed by name
	names   []string                // ordered names
	aliases map[string]string       // map of aliases to real names
	renames map[string]string       // map of old names to real names
}

// FetchFormula implements formula.Formulary.
func (index *V1Cache) FetchFormula(_ context.Context, name string) (formula.MultiPlatformFormula, error) {
	data := index.Find(name)
	if data == nil {
		return nil, errdef.NewFormulaNotFoundError(name)
	}
	return formula.FromV1(data), nil
}

// cacheV1 creates a new Index for a Homebrew API source.
func cacheV1(index []*brewv1.Info) *V1Cache {
	a := &V1Cache{
		mapped:  make(map[string]*brewv1.Info, len(index)),
		names:   make([]string, len(index)),
		aliases: map[string]string{},
		renames: map[string]string{},
	}

	// Ingest index content
	for i, info := range index {
		a.mapped[info.Name] = info // map name to info for lookups
		a.names[i] = info.Name     // add name to list for iteration
		for _, alias := range info.Aliases {
			a.aliases[alias] = info.Name // map alias to real name for alias lookups
		}
		for _, renamed := range info.OldNames {
			a.renames[renamed] = info.Name // map old name to real name for rename lookups
		}
	}

	return a
}

// Find finds a formula.
func (index *V1Cache) Find(name string) *brewv1.Info {
	// Look up the name
	f, ok := index.mapped[name]
	if ok {
		return f
	}

	// Look up as alias
	rname, ok := index.aliases[name]
	if ok {
		return index.Find(rname)
	}

	return nil
}

// List produces the contents of the index.
func (index *V1Cache) List() brewv1.Index {
	list := make(brewv1.Index, len(index.names))
	for i, name := range index.names {
		list[i] = index.mapped[name]
	}
	return list
}

// ListNames produces the names in the index.
func (index *V1Cache) ListNames() []string {
	return slices.Clone(index.names)
}

// SearchFunc searches the index and returns all formulae hits from the match function.
func (index *V1Cache) SearchFunc(match func(*brewv1.Info) bool) []*brewv1.Info {
	hits := []*brewv1.Info{}
	for _, name := range index.names {
		f := index.mapped[name]
		if match(f) {
			hits = append(hits, f)
		}
	}
	return hits
}

// Aliases returns the map of aliases.
func (index *V1Cache) Aliases() map[string]string {
	return maps.Clone(index.aliases)
}

func writeAPICache(cached *V1Cache, dir string) error {
	// Create parent directory
	err := os.MkdirAll(dir, 0o775)
	if err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	err = api.WriteFormulaNames(cached.ListNames(), namesFile(dir))
	if err != nil {
		return err
	}

	err = api.WriteFormulaAliases(cached.Aliases(), aliasesFile(dir))
	if err != nil {
		return err
	}

	return nil
}
