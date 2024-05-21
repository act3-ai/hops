package brewformulary

import (
	"context"

	brew "github.com/act3-ai/hops/internal/brew"
	"github.com/act3-ai/hops/internal/formula"
)

// NewFormulary creates a pre-loaded formulary
func NewFormulary(index formula.Index) (formula.Formulary, error) {
	return NewPreloaded(index)
}

// NewPreloaded creates a pre-loaded formulary
func NewPreloaded(index formula.Index) (*Preloaded, error) {
	return &Preloaded{
		index: index,
	}, nil
}

// Preloaded is a formulary with the full contents of the Homebrew API
type Preloaded struct {
	index formula.Index
}

// Fetch implements formula.Formulary.
func (f *Preloaded) Fetch(_ context.Context, name string) (formula.MultiPlatformFormula, error) {
	data := f.index.Find(name)
	if data == nil {
		return nil, brew.NewErrFormulaNotFound(name)
	}
	return formula.FromV1(data), nil
}
