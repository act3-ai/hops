package cached

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const (
	FormulaNamesFile   = "formula_names.txt"   // FormulaAliasesFile is the name of the formula name cache
	FormulaAliasesFile = "formula_aliases.txt" // FormulaAliasesFile is the name of the formula aliases cache
)

// FormulaNames represents the contents of the formula names file
type FormulaNames []string

// FormulaAliases represents the cached formula aliases file
// Sorted by the second element (the target name)
type FormulaAliases map[string]string

// LoadFormulaNames loads a formula names list from a file
func LoadFormulaNames(file string) (FormulaNames, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading formula names file %s: %w", file, err)
	}
	fn := strings.Split(string(b), "\n")

	// HACK: Ruby and Go disagree about string comparison
	// the "+" character is sorted different
	// name with + will not be found with Homebrew's given sorting
	slices.Sort(fn)

	return fn, nil
}

// Has reports if the name is found in the formula name list
func (fn FormulaNames) Has(name string) bool {
	_, found := slices.BinarySearch(fn, name)
	return found
}

// Index reports the index of name in the formula name list, -1 if not found
func (fn FormulaNames) Index(name string) int {
	i, found := slices.BinarySearch(fn, name)
	if !found {
		return -1
	}
	return i
}

// WriteFormulaNames writes the formula names to file
func WriteFormulaNames(fn FormulaNames, file string) error {
	err := os.WriteFile(file, []byte(strings.Join(fn, "\n")+"\n"), 0o644)
	if err != nil {
		return fmt.Errorf("writing formula names file: %w", err)
	}
	return nil
}

// LoadFormulaAliases loads a formula aliases list from a file
func LoadFormulaAliases(file string) (FormulaAliases, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("opening formula aliases file %s: %w", file, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '|'
	r.Comment = '#'
	r.FieldsPerRecord = 2 // limit to 2 fields
	r.LazyQuotes = true   // perf
	r.ReuseRecord = true  // perf

	aliases := FormulaAliases{}

	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing formula aliases file %s: %w", file, err)
		}
		aliases[record[0]] = record[1]
	}

	return aliases, nil
}

// MarshalCSV returns the aliases marshaled to a pipe-delimited CSV
func (fa FormulaAliases) MarshalCSV() ([]byte, error) {
	// Write as pipe-delimited CSV
	buf := new(bytes.Buffer)

	// Create list of aliases
	aliases := make([]string, 0, len(fa))
	for k := range fa {
		aliases = append(aliases, k)
	}

	// Sort aliases by real name
	slices.SortStableFunc(aliases, func(alias1, alias2 string) int {
		return strings.Compare(fa[alias1], fa[alias2])
	})

	// Write aliases in order
	for _, alias := range aliases {
		_, err := buf.WriteString(alias + "|" + fa[alias] + "\n")
		if err != nil {
			return nil, err
		}
	}
	return append(buf.Bytes(), []byte("\n")...), nil
}

// WriteFormulaNames writes the formula names to file
func WriteFormulaAliases(fa FormulaAliases, file string) error {
	b, err := fa.MarshalCSV()
	if err != nil {
		return fmt.Errorf("marshalling formula aliases: %w", err)
	}
	err = os.WriteFile(file, b, 0o644)
	if err != nil {
		return fmt.Errorf("writing formula file file: %w", err)
	}
	return nil
}
