package common

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
)

// FormulaDeprecateDisableReasons contains known deprecation and disabling reasons.
//
// https://github.com/Homebrew/brew/blob/master/Library/Homebrew/deprecate_disable.rb
var FormulaDeprecateDisableReasons = map[string]string{
	"does_not_build":      "does not build",
	"no_license":          "has no license",
	"repo_archived":       "has an archived upstream repository",
	"repo_removed":        "has a removed upstream repository",
	"unmaintained":        "is not maintained upstream",
	"unsupported":         "is not supported upstream",
	"deprecated_upstream": "is deprecated upstream",
	"versioned_formula":   "is a versioned formula",
	"checksum_mismatch": heredoc.Doc(`
		was built with an initially released source file that had 
		a different checksum than the current one.
		Upstream's repository might have been compromised.
		We can re-package this once upstream has confirmed that they retagged their release`),
}

// FormulaDeprecateReason describes why a formula was deprecated.
type FormulaDeprecateReason string

// UnmarshalJSON implements json.Unmarshaler.
func (r *FormulaDeprecateReason) UnmarshalJSON(b []byte) error {
	var err error
	*r, err = reasonUnmarshalJSON[FormulaDeprecateReason](b, formulaDeprecateDisableReasons)
	if err != nil {
		return fmt.Errorf("unmarshaling deprecation_reason: %w", err)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (r FormulaDeprecateReason) MarshalJSON() ([]byte, error) {
	b, err := reasonMarshalJSON(r, kegOnlyReasons)
	if err != nil {
		return b, fmt.Errorf("marshaling deprecation_reason: %w", err)
	}
	return b, nil
}

// FormulaDisableReason describes why a formula was disabled.
type FormulaDisableReason string

// UnmarshalJSON implements json.Unmarshaler.
func (r *FormulaDisableReason) UnmarshalJSON(b []byte) error {
	var err error
	*r, err = reasonUnmarshalJSON[FormulaDisableReason](b, formulaDeprecateDisableReasons)
	if err != nil {
		return fmt.Errorf("unmarshaling disabled_reason: %w", err)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (r FormulaDisableReason) MarshalJSON() ([]byte, error) {
	b, err := reasonMarshalJSON(r, kegOnlyReasons)
	if err != nil {
		return b, fmt.Errorf("marshaling disabled_reason: %w", err)
	}
	return b, nil
}

// formulaDeprecateDisableReasons defines known deprecation and disabling reasons.
//
// https://github.com/Homebrew/brew/blob/master/Library/Homebrew/deprecate_disable.rb
var formulaDeprecateDisableReasons = map[string]string{
	":does_not_build":      "does not build",
	":no_license":          "has no license",
	":repo_archived":       "has an archived upstream repository",
	":repo_removed":        "has a removed upstream repository",
	":unmaintained":        "is not maintained upstream",
	":unsupported":         "is not supported upstream",
	":deprecated_upstream": "is deprecated upstream",
	":versioned_formula":   "is a versioned formula",
	":checksum_mismatch": heredoc.Doc(`
		was built with an initially released source file that had 
		a different checksum than the current one.
		Upstream's repository might have been compromised.
		We can re-package this once upstream has confirmed that they retagged their release`),
}

// KegOnlyConfig declares if a formula is keg-only and why.
type KegOnlyConfig struct {
	Reason      KegOnlyReason `json:"reason,omitempty"`
	Explanation string        `json:"explanation,omitempty"`
}

// KegOnlyReason is used for the reason field in keg_only_reason.
type KegOnlyReason string

// UnmarshalJSON implements json.Unmarshaler.
func (r *KegOnlyReason) UnmarshalJSON(b []byte) error {
	var err error
	*r, err = reasonUnmarshalJSON[KegOnlyReason](b, kegOnlyReasons)
	if err != nil {
		return fmt.Errorf("unmarshaling keg_only_reason: %w", err)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (r KegOnlyReason) MarshalJSON() ([]byte, error) {
	b, err := reasonMarshalJSON(r, kegOnlyReasons)
	if err != nil {
		return b, fmt.Errorf("marshaling keg_only_reason: %w", err)
	}
	return b, nil
}

// knownKegOnlyReasons stores the reason aliases.
var kegOnlyReasons = map[string]string{
	":versioned_formula": "this is an alternate version of another formula.",
	":provided_by_macos": "macOS already provides this software and installing another version in parallel can cause all kinds of trouble.",
	":shadowed_by_macos": "macOS provides similar software and installing this software in parallel can cause all kinds of trouble.",
}

// reasonUnmarshalJSON is a helper for json.Unmarshaler functions.
func reasonUnmarshalJSON[T ~string](b []byte, aliases map[string]string) (T, error) {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return T(s), err
	}

	if value, ok := aliases[s]; ok {
		// if s is an alias, set it to the aliased value
		s = value
	}

	return T(s), nil
}

// reasonMarshalJSON is a helper for json.Marshaler functions.
func reasonMarshalJSON[T ~string](r T, aliases map[string]string) ([]byte, error) {
	s := string(r)

	// Check if value is an aliased value
	for name, value := range aliases {
		if s == value {
			s = name // if s is an aliased value, set it to the alias
			break
		}
	}

	// marshal the value as a string
	return json.Marshal(s)
}
