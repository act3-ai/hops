package brewclient

import (
	"maps"
	"slices"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

// APIIndex represents a formula index from the Homebrew API
type APIIndex struct {
	mapped  map[string]*brewv1.Info // full contents indexed by name
	names   []string                // ordered names
	aliases map[string]string       // map of aliases to real names
	// oldnames map[string]string   // map of old names to real names
}

// NewAPIIndex creates a new Index for a Homebrew API source
func NewAPIIndex(index brewv1.Index) *APIIndex {
	a := &APIIndex{
		// Contents: v1.Index{},
		mapped:  make(map[string]*brewv1.Info, len(index)),
		names:   make([]string, len(index)),
		aliases: map[string]string{},
	}

	// Ingest index content
	for i, info := range index {
		a.mapped[info.Name] = info // map name to info for lookups
		a.names[i] = info.Name     // add name to list for iteration
		for _, alias := range info.Aliases {
			a.aliases[alias] = info.Name // map alias to real name for alias lookups
		}
	}

	return a
}

// Find finds a formula
func (index *APIIndex) Find(name string) *brewv1.Info {
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

// List produces the contents of the index
func (index *APIIndex) List() brewv1.Index {
	list := make(brewv1.Index, len(index.names))
	for i, name := range index.names {
		list[i] = index.mapped[name]
	}
	return list
}

// ListNames produces the names in the index
func (index *APIIndex) ListNames() []string {
	return slices.Clone(index.names)
}

// SearchFunc searches the index and returns all formulae hits from the match function
func (index *APIIndex) SearchFunc(match func(*brewv1.Info) bool) []*brewv1.Info {
	hits := []*brewv1.Info{}
	for _, name := range index.names {
		f := index.mapped[name]
		if match(f) {
			hits = append(hits, f)
		}
	}
	return hits
}

// Aliases returns the map of aliases
func (index *APIIndex) Aliases() map[string]string {
	return maps.Clone(index.aliases)
}
