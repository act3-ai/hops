package v2

import (
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

// Info represents the v2 Homebrew API
type Info struct {
	Formulae []*Formula `json:"formulae"`
	Casks    []*Cask    `json:"casks"`
}

// Formula represents a formula entry
type Formula v1.Info
