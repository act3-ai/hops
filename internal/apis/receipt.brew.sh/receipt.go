package receipt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	tab "github.com/act3-ai/hops/internal/apis/sh.brew.tab"
)

// InstallReceiptFile is the name of the file.
const InstallReceiptFile = "INSTALL_RECEIPT.json"

// InstallReceipt represents the INSTALL_RECEIPT.json file.
type InstallReceipt struct {
	HomebrewVersion       string                      `json:"homebrew_version"`
	UsedOptions           []any                       `json:"used_options"`
	UnusedOptions         []any                       `json:"unused_options"`
	BuiltAsBottle         bool                        `json:"built_as_bottle"`
	PouredFromBottle      bool                        `json:"poured_from_bottle"`
	LoadedFromAPI         bool                        `json:"loaded_from_api"`
	InstalledAsDependency bool                        `json:"installed_as_dependency"`
	InstalledOnRequest    bool                        `json:"installed_on_request"`
	ChangedFiles          []string                    `json:"changed_files"`
	Time                  uint                        `json:"time"`
	SourceModifiedTime    uint                        `json:"source_modified_time"`
	Compiler              string                      `json:"compiler"`
	Aliases               []string                    `json:"aliases"`
	RuntimeDependencies   []*brewv1.RuntimeDependency `json:"runtime_dependencies"`
	Source                Source                      `json:"source"`
	Arch                  string                      `json:"arch"`
	BuiltOn               tab.BuiltOn
}

// Source section of the receipt.
type Source struct {
	Spec       string          `json:"spec"`
	Versions   brewv1.Versions `json:"versions"`
	Path       string          `json:"path"`
	TapGitHead string          `json:"tap_git_head"`
	Tap        string          `json:"tap"`
}

// Load loads the INSTALL_RECEIPT.json for a keg.
//
// A return value of nil, nil signifies that no install receipt was found.
func Load(keg string) (*InstallReceipt, error) {
	b, err := os.ReadFile(filepath.Join(keg, InstallReceiptFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("reading install receipt: %w", err)
	}

	r := &InstallReceipt{}
	err = json.Unmarshal(b, r)
	if err != nil {
		return nil, fmt.Errorf("parsing install receipt: %w", err)
	}

	return r, nil
}

// CreateInstallReceipt creates an install receipt for the formula.
func CreateInstallReceipt(version string, requested []string, info *brewv1.Info) *InstallReceipt {
	isDep := true
	for _, r := range requested {
		for _, n := range info.PossibleNames() {
			if n == r {
				isDep = false
			}
		}
	}

	return &InstallReceipt{
		HomebrewVersion:       "hops+" + version,
		BuiltAsBottle:         true,
		PouredFromBottle:      true,
		LoadedFromAPI:         true,
		InstalledAsDependency: isDep,
		InstalledOnRequest:    !isDep,
		// Time:                  time.Now(),
		Aliases: info.Aliases,
	}
}

// NewInstallReceipt creates an install receipt for a formula.
func NewInstallReceipt(info *brewv1.Info, t *tab.Tab, requested bool, hopsVersion string) *InstallReceipt {
	return &InstallReceipt{
		HomebrewVersion:       "hops+" + hopsVersion,
		BuiltAsBottle:         true,
		PouredFromBottle:      true,
		LoadedFromAPI:         true,
		InstalledAsDependency: !requested,
		InstalledOnRequest:    requested,
		ChangedFiles:          t.ChangedFiles,
		Time:                  uint(time.Now().Unix()),
		SourceModifiedTime:    t.SourceModifiedTime,
		Compiler:              t.Compiler,
		Aliases:               info.Aliases,
		RuntimeDependencies:   t.RuntimeDependencies,
		Source: Source{
			Spec:     brewv1.Stable,
			Versions: info.Versions,
			Path:     "HOMEBREW_PREFIX/Library/Taps/homebrew/homebrew-core/Formula/g/git.rb",
		},
		Arch:    t.Arch,
		BuiltOn: t.BuiltOn,
	}
}
