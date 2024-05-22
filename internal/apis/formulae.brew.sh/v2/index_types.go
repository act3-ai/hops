package v2

import (
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
)

// Tap represents the v2 Homebrew API
type Tap struct {
	Formulae []*Formula `json:"formulae"`
	Casks    []*Cask    `json:"casks"`
}

// Formula represents a formula entry
type Formula brewv1.Info
