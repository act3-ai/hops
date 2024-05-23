package v3

// Cask represents a cask's metadata.
type Cask struct {
	Token            string             `json:"token"`
	Name             string             `json:"name"`
	Description      string             `json:"desc"`
	Homepage         string             `json:"homepage"`
	URL              string             `json:"url"`
	Version          string             `json:"version"`
	Sha256           string             `json:"sha256"`
	Artifacts        []CaskArtifact     `json:"artifacts"`
	RubySourcePath   string             `json:"ruby_source_path"`
	RubySourceSHA256 string             `json:"ruby_source_sha256"`
	URLSpecs         map[string]URLSpec `json:"url_specs,omitempty"`
	DependsOn        struct {
		MacOS map[string][]string `json:"macos"`
	} `json:"depends_on"`
}

// CaskArtifact represents an artifact stanza defined by a cask.
//
// https://docs.brew.sh/Cask-Cookbook#stanza-descriptions
type CaskArtifact map[string][]any

// URLSpec the url_specs entry.
type URLSpec struct {
	Verified  string            `json:"verified,omitempty"`
	Using     string            `json:"using,omitempty"`   // one of "post" or "homebrew_curl"
	Cookies   map[string]string `json:"cookies,omitempty"` // 4 examples
	Referer   string            `json:"referer,omitempty"`
	Header    any               `json:"header,omitempty"`     // string or array of strings
	UserAgent string            `json:"user_agent,omitempty"` // 37 examples
	Data      map[string]string `json:"data,omitempty"`       // 2 examples: segger-jlink, segger-ozone
}
