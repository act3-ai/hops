package formula

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

// Names returns the names of the listed formulae
func Names(formulae []*v1.Info) []string {
	names := make([]string, len(formulae))
	for i, f := range formulae {
		names[i] = f.Name
	}
	return names
}
