package formula

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
)

// Formula represents a formula
type Formula struct {
	v1.Info // API information about the formula
}

// New creates a new Formula
func New(info *v1.Info) *Formula {
	return &Formula{
		Info: *info,
	}
}

// ManifestTag produces the tag of the formula's manifest
func (f *Formula) Repo() string {
	return brewfmt.Repo(f.FullName)
}

// ToV1 converts a list of formulae to a v1 API form
func ToV1(formulae ...*Formula) v1.Index {
	result := make(v1.Index, len(formulae))

	for i, f := range formulae {
		result[i] = &f.Info
	}

	return result
}

// FromV1 converts a list of formulae to a v1 API form
func FromV1(formulae ...*v1.Info) []*Formula {
	result := make([]*Formula, len(formulae))

	for i, f := range formulae {
		result[i] = New(f)
	}

	return result
}

// Names returns the names of the listed formulae
func Names(formulae []*Formula) []string {
	names := make([]string, len(formulae))
	for i, f := range formulae {
		names[i] = f.Name
	}
	return names
}
