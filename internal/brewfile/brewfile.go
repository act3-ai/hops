// Package brewfile implements a simple Brewfile parser.
package brewfile

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Brewfile represents a "brew bundle" Brewfile.
type Brewfile struct {
	Tap     []string // Homebrew Taps
	Formula []string // Homebrew Formulae
	// CaskArgs  []string // Arguments passed to
	// Cask      []string
	// MAS       []string
	// Whalebrew []string
	// VSCode    []string
}

// Load loads a Brewfile.
func Load(path string) (*Brewfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening Brewfile %s: %w", path, err)
	}
	defer file.Close()

	bf, err := Parse(file)
	if err != nil {
		return nil, fmt.Errorf("loading Brewfile %s: %w", path, err)
	}
	return bf, nil
}

// Parse parses a Brewfile from an io.Reader.
func Parse(r io.Reader) (*Brewfile, error) {
	bf := &Brewfile{
		Tap:     []string{},
		Formula: []string{},
		// CaskArgs:  []string{},
		// Cask:      []string{},
		// MAS:       []string{},
		// Whalebrew: []string{},
		// VSCode:    []string{},
	}

	// Iterate through lines of the file
	scanner := bufio.NewScanner(r)
	i := 0
	for scanner.Scan() {
		line := scanner.Text() // read line into string
		i++

		// Skip empty lines and comment lines
		if strings.TrimSpace(line) == "" ||
			strings.HasPrefix(line, "#") {
			continue
		}

		args := strings.Fields(line)

		// Skip lines with too few arguments
		if len(args) < 2 {
			continue
		}

		// Identifies lines kind
		kind := args[0]
		// Remove trailing comma and quotes
		name := strings.Trim(strings.TrimSuffix(args[1], ","), `"'`)

		switch kind {
		case "tap":
			bf.Tap = append(bf.Tap, name)
		case "brew":
			bf.Formula = append(bf.Formula, name)
		case "cask_args", "cask", "mas", "whalebrew", "vscode":
			// Catch all recognized fields, warn that they are being ignored
			slog.Warn("skipping unsupported Brewfile entry", slog.String("kind", kind), slog.Int("lineNumber", i), slog.String("line", line))
			continue
		// case "cask_args":
		// 	// Append unmodified, don't want to parse it more than this
		// 	bf.CaskArgs = append(bf.CaskArgs, args[1])
		// case "cask":
		// 	bf.Cask = append(bf.Cask, name)
		// case "mas":
		// 	bf.MAS = append(bf.MAS, name)
		// case "whalebrew":
		// 	bf.Whalebrew = append(bf.Whalebrew, name)
		// case "vscode":
		// 	bf.VSCode = append(bf.VSCode, name)
		default:
			// slog.Debug("skipping unsupported Brewfile line", slog.String("line", line))
			// continue
			return nil, fmt.Errorf("could not parse Brewfile:%d: unrecognized argument \"%s\" in line: %s", i, kind, line)
		}
	}

	return bf, nil
}
