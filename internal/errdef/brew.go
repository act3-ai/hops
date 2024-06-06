package errdef

import (
	"strconv"
)

// FormulaNotFoundError is emitted when a a formula could not be found.
type FormulaNotFoundError struct {
	name string
}

// Error implements error.
func (err FormulaNotFoundError) Error() string {
	return "no available formula with the name " + strconv.Quote(err.name)
}

// NewFormulaNotFoundError produces a FormulaNotFoundError.
func NewFormulaNotFoundError(name string) error {
	return FormulaNotFoundError{name: name}
}

// FormulaUpToDateError reports a formula already installed.
type FormulaUpToDateError struct {
	name    string
	version string
}

// Error implements error.
func (err FormulaUpToDateError) Error() string {
	return err.name + " " + err.version + " is already installed and up-to-date"
}

// NewFormulaUpToDateError produces a FormulaUpToDateError.
func NewFormulaUpToDateError(name, version string) error {
	return FormulaUpToDateError{
		name:    name,
		version: version,
	}
}
