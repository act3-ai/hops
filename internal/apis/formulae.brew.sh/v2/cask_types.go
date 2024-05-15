package v2

// Cask represents a cask entry
type Cask struct {
	Token              string         `json:"token"`
	FullToken          string         `json:"full_token"`
	OldTokens          []string       `json:"old_tokens"`
	Tap                string         `json:"tap"`
	Name               []string       `json:"name"`
	Desc               string         `json:"desc"`
	Homepage           string         `json:"homepage"`
	URL                string         `json:"url"`
	URLSpecs           map[string]any `json:"url_specs"`
	Appcast            any            `json:"appcast"`
	Version            string         `json:"version"`
	Installed          string         `json:"installed"`
	InstalledTime      string         `json:"installed_time"`
	BundleVersion      string         `json:"bundle_version"`
	BundleShortVersion string         `json:"bundle_short_version"`
	Outdated           bool           `json:"outdated"`
	Sha256             string         `json:"sha256"`
	Artifacts          []struct {
		App []string `json:"app,omitempty"`
		Zap []struct {
			Trash []string `json:"trash,omitempty"`
		} `json:"zap,omitempty"`
	} `json:"artifacts"`
	Caveats   *string `json:"caveats"`
	DependsOn struct {
		MacOS map[string][]string `json:"macos"`
	} `json:"depends_on"`
	ConflictsWith struct {
		Cask []string `json:"cask"`
	} `json:"conflicts_with"`
	Container          any     `json:"container"`
	AutoUpdates        bool    `json:"auto_updates"`
	Deprecated         bool    `json:"deprecated"`
	DeprecationDate    *string `json:"deprecation_date"`
	DeprecationReason  *string `json:"deprecation_reason"`
	Disabled           bool    `json:"disabled"`
	DisableDate        *string `json:"disable_date"`
	DisableReason      *string `json:"disable_reason"`
	TapGitHead         string
	Languages          []string `json:"languages"`
	RubySourcePath     string   `json:"ruby_source_path"`
	RubySourceChecksum struct {
		Sha256 string `json:"sha256"`
	} `json:"ruby_source_checksum"`
}
