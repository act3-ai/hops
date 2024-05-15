package brewfile

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Brewfile represents a brew bundle Brewfile
type Brewfile struct {
	Tap       []string // Homebrew Taps
	Brew      []string // Homebrew Formulae
	CaskArgs  []string // Arguments passed to
	Cask      []string
	MAS       []string
	Whalebrew []string
	VSCode    []string
}

// Load loads a Brewfile
func Load(name string) (*Brewfile, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("%s %w", name, err)
	}
	defer file.Close()

	bf := &Brewfile{
		Tap:       []string{},
		Brew:      []string{},
		CaskArgs:  []string{},
		Cask:      []string{},
		MAS:       []string{},
		Whalebrew: []string{},
		VSCode:    []string{},
	}

	// Iterate through lines of the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text() // read line into string

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
			bf.Brew = append(bf.Brew, name)
		case "cask_args":
			// Append unmodified, don't want to parse it more than this
			bf.CaskArgs = append(bf.CaskArgs, args[1])
		case "cask":
			bf.Cask = append(bf.Cask, name)
		case "mas":
			bf.MAS = append(bf.MAS, name)
		case "whalebrew":
			bf.Whalebrew = append(bf.Whalebrew, name)
		case "vscode":
			bf.VSCode = append(bf.VSCode, name)
		default:
			slog.Debug("skipping unknown line", slog.String("line", line))
			continue
		}
	}

	return bf, nil
}
