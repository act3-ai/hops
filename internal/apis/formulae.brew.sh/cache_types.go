package api

import (
	"github.com/act3-ai/hops/internal/apis/formulae.brew.sh/cached"
)

const (
	CachedFormulaNamesFile   = cached.FormulaNamesFile   // FormulaAliasesFile is the name of the formula name cache
	CachedFormulaAliasesFile = cached.FormulaAliasesFile // FormulaAliasesFile is the name of the formula aliases cache
)

// FormulaNames represents the contents of the formula names file
type FormulaNames cached.FormulaNames

// FormulaAliases represents the cached formula aliases file
// Sorted by the second element (the target name)
type FormulaAliases cached.FormulaAliases

// LoadFormulaNames loads a formula names list from a file
var LoadFormulaNames = cached.LoadFormulaNames

// WriteFormulaNames writes the formula names to file
var WriteFormulaNames = cached.WriteFormulaNames

// LoadFormulaAliases loads a formula aliases list from a file
var LoadFormulaAliases = cached.LoadFormulaAliases

// WriteFormulaNames writes the formula names to file
var WriteFormulaAliases = cached.WriteFormulaAliases
