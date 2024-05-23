package apiutil

import (
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

// ConfigDocumentedPath produces a documentable list of configuration file paths.
// Environment variables are not evaluated.
func ConfigDocumentedPath(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join("$XDG_CONFIG_HOME", filepath.Join(parts...)),
		filepath.Join("/", "etc", filepath.Join(parts...)),
	}
}

// ConfigActualPaths returns the list of locations to look for configuration files.
func ConfigActualPaths(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join(xdg.ConfigHome, filepath.Join(parts...)),
		filepath.Join("/", "etc", filepath.Join(parts...)),
	}
	// TODO: also search $XDG_CONFIG_DIRS
}

// DefaultConfigPath is the preferred configuration path.
func DefaultConfigPath(parts ...string) string {
	return filepath.Join(xdg.ConfigHome, filepath.Join(parts...))
}

// ConfigMatchPaths returns the list of paths to validate as configuration files.
func ConfigMatchPaths(parts ...string) []string {
	return []string{
		strings.Join(parts, "-"),
		filepath.Join(parts...),
	}
}
