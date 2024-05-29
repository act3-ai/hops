package v3

import "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/common"

// Cask represents a cask's metadata.
type Cask struct {
	Token            string                             `json:"token"`
	Name             string                             `json:"name"`
	Description      string                             `json:"desc"`
	Homepage         string                             `json:"homepage"`
	URL              string                             `json:"url"`
	Version          string                             `json:"version"`
	Sha256           string                             `json:"sha256"`
	Artifacts        []common.CaskArtifact              `json:"artifacts"`
	RubySourcePath   string                             `json:"ruby_source_path"`
	RubySourceSHA256 string                             `json:"ruby_source_sha256"`
	URLSpecs         map[string]URLSpec                 `json:"url_specs,omitempty"`
	DependsOn        map[string]common.CaskDependencies `json:"depends_on"`
}

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
