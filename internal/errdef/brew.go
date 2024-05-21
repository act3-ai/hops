package errdef

import (
	"errors"
	"fmt"
)

// ErrFormulaNotFound represents a formula search error
var ErrFormulaNotFound = errors.New("formula not found")

// NewErrFormulaNotFound wraps ErrFormulaNotFound with the formula name
func NewErrFormulaNotFound(formula string) error {
	return fmt.Errorf("%q: %w", formula, ErrFormulaNotFound)
}

// ErrFormulaUpToDate reports a formula already installed
var ErrFormulaUpToDate = errors.New("already installed and up-to-date")

// NewErrFormulaUpToDate wraps ErrFormulaUpToDate with the formula name
func NewErrFormulaUpToDate(name, version string) error {
	return fmt.Errorf("%s %s is %w", name, version, ErrFormulaUpToDate)
}
