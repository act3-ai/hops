package env

import (
	"path/filepath"
	"strconv"
)

// NotEmpty reports if name is set and not empty
//
// bash: [[ -n "${name-}" ]]
func NotEmpty(name string) bool {
	return String(name, "") != ""
}

// String is a helper for Or[string]
func String(name, def string) string {
	return Or(name, def, func(envVal string) (string, error) { return envVal, nil })
}

// RequiredString is a helper for Required[string]
func RequiredString(name string) (string, error) {
	return Required(name, func(envVal string) (string, error) { return envVal, nil })
}

// OneOfString is a helper for OneOfOr[string]
func OneOfString(names []string, def string) string {
	return OneOfOr(names, def, func(envVal string) (string, error) { return envVal, nil })
}

// OneOfRequiredString is a helper for OneOfRequired[string]
func OneOfRequiredString(names []string) (string, error) {
	return OneOfRequired(names, func(envVal string) (string, error) { return envVal, nil })
}

// Int is a helper for Or[int]
func Int(name string, def int) int {
	return Or(name, def, strconv.Atoi)
}

// StrictInt is a helper for StrictOr[int]
func StrictInt(name string, def int) (int, error) {
	return StrictOr(name, def, strconv.Atoi)
}

// RequiredInt is a helper for Required[int]
func RequiredInt(name string) (int, error) {
	return Required(name, strconv.Atoi)
}

// Bool is a helper for Or[bool]
func Bool(name string, def bool) bool {
	return Or(name, def, strconv.ParseBool)
}

// StrictBool is a helper for StrictOr[bool]
func StrictBool(name string, def bool) (bool, error) {
	return StrictOr(name, def, strconv.ParseBool)
}

// RequiredBool is a helper for Required[bool]
func RequiredBool(name string) (bool, error) {
	return Required(name, strconv.ParseBool)
}

// StringSlice is a helper for Slice[string]
func StringSlice(name string, def []string, sep string) []string {
	return Slice(name, def, sep, func(envVal string) (string, error) { return envVal, nil })
}

// RequiredStringSlice is a helper for RequiredSlice[string]
func RequiredStringSlice(name string, sep string) ([]string, error) {
	return RequiredSlice(name, sep, func(envVal string) (string, error) { return envVal, nil })
}

// PathSlice grabs the env variable as an array splitting on the default (OS specific) path list separator
func PathSlice(name string, def []string) []string {
	return StringSlice(name, def, string(filepath.ListSeparator))
}
