package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/dustin/go-humanize"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// FmtSemver forces a string formatted as a semver version
func FmtSemver(s string) string {
	return "v" + strings.TrimPrefix(s, "v")
}

// TerminalWidth returns the width of the terminal, using fallback if it can't determine width
func TerminalWidth(fallback int) int {
	tty := termenv.DefaultOutput().TTY()
	if tty == nil {
		return fallback
	}
	width, _, err := term.GetSize(int(tty.Fd()))
	if err != nil {
		return fallback
	}
	return width
}

// PrettyBytes returns the number of bytes formatted for human readability
func PrettyBytes(size int64) string {
	return strings.ReplaceAll(humanize.Bytes(uint64(size)), " ", "")
}

// ParseRegistryDomain
func ParseRegistryDomain(rawURL string) string {
	for _, prefix := range []string{"https://", "http://"} {
		rawURL = strings.TrimPrefix(rawURL, prefix)
	}
	return strings.Replace(rawURL, "/v2/", "/", 1)
}

// ConfigSearchPath returns the list of locations to look for configuration files
func ConfigSearchPath(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join(xdg.ConfigHome, filepath.Join(parts...)),
		filepath.Join("/", "etc", filepath.Join(parts...)),
	}
	// TODO we should consider searching $XDG_CONFIG_DIRS as well
}

// DefaultConfigPath is the path we would save the configuration to if needed.  In a sense it is the preferred configuration path.
func DefaultConfigPath(parts ...string) string {
	return filepath.Join(xdg.ConfigHome, filepath.Join(parts...))
}

// ConfigValidatePath returns the list of paths to validate as configuration files
func ConfigValidatePath(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join(parts...),
	}
}

// AssertStrings converts a slice of any into a slice of strings
func AssertStrings(as []any) ([]string, error) {
	ss := make([]string, 0, len(as))
	for _, a := range as {
		s, ok := a.(string)
		if !ok {
			return nil, fmt.Errorf("expected list of strings, got %T in list", a)
		}
		ss = append(ss, s)
	}
	return ss, nil
}

// CountDir returns the number of files in a directory and the total size
func CountDir(dir string) (int, int64, error) {
	files, size := 0, int64(0)
	return files, size, fs.WalkDir(os.DirFS(dir), ".", func(_ string, d fs.DirEntry, err error) error {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil
		case err != nil:
			return err
		}

		if d.IsDir() {
			return nil
		}

		files++

		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			size += info.Size()
		}

		return nil
	})
}

// StrOrStrSlice represents the command run for a service
type StrOrStrSlice []string

// UnmarshalJSON implements json.Unmarshaler
func (s *StrOrStrSlice) UnmarshalJSON(b []byte) error {
	var i any
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	switch v := i.(type) {
	case string:
		*s = []string{v}
	case []any:
		tmp := make([]string, 0, len(v))
		for _, i := range v {
			if j, ok := i.(string); ok {
				tmp = append(tmp, j)
			} else {
				return fmt.Errorf("non-string type %T in slice", i)
			}
		}
		*s = tmp
	// case map[string]any:
	// 	fmt.Println("map found:" + string(b))
	default:
		return fmt.Errorf("cannot unmarshal %s into string or string slice", i)
	}

	return nil
}

// // StrOrStrSlice represents the command run for a service
// type StrOrStrSlice []string

// // UnmarshalJSON implements json.Unmarshaler
// func (s *StrOrStrSlice) UnmarshalJSON(b []byte) error {
// 	var i any
// 	err := json.Unmarshal(b, &i)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("bytes: " + string(b))

// 	switch v := i.(type) {
// 	case string:
// 		*s = []string{v}
// 	case []any:
// 		tmp := make([]string, 0, len(v))
// 		for _, i := range v {
// 			if j, ok := i.(string); ok {
// 				tmp = append(tmp, j)
// 			} else {
// 				return fmt.Errorf("non-string type %T in slice", i)
// 			}
// 		}
// 		*s = tmp
// 	// case map[string]any:
// 	// 	fmt.Println("map found:" + string(b))
// 	default:
// 		return fmt.Errorf("cannot unmarshal %s into string or string slice", i)
// 	}

// 	return nil
// }

// // UnmarshalYAML implements yaml.Unmarshaler
// func (s *StrOrStrSlice) UnmarshalYAML(b []byte) error {
// 	var i any
// 	err := yaml.Unmarshal(b, &i)
// 	if err != nil {
// 		return err
// 	}

// 	switch v := i.(type) {
// 	case string:
// 		*s = []string{v}
// 	case []any:
// 		tmp := make([]string, 0, len(v))
// 		for _, i := range v {
// 			if j, ok := i.(string); ok {
// 				tmp = append(tmp, j)
// 			} else {
// 				return fmt.Errorf("%s", "non-string Type in Run slice")
// 			}
// 		}
// 		*s = tmp
// 	default:
// 		return fmt.Errorf("unknown Type for Run value")
// 	}

// 	return nil
// }

// BytesAreEmptyIsh checks if the provided byte slice is practically empty
// Can handle cases where a nil pointer or empty struct were marshalled to YAML
func BytesAreEmptyIsh(b []byte) bool {
	return b == nil ||
		bytes.Equal(bytes.TrimSpace(b), []byte("{}")) ||
		bytes.Equal(bytes.TrimSpace(b), []byte("null"))
}

// FilterNil filters nil entries from a list
func FilterNil[T any](list []*T) []*T {
	result := []*T{}
	for _, entry := range list {
		if entry != nil {
			result = append(result, entry)
		}
	}
	return result
}
