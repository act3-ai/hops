package v1beta1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adrg/xdg"

	"github.com/act3-ai/hops/internal/apis/apiutil"
	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/utils/env"
)

var (
	// configurationFileParts is the parts used to assemble
	// the config file name and default paths.
	configurationFileParts = []string{"hops", "config.yaml"}

	// ConfigurationMatchFiles is a list of patterns used to match a config file for schema validation.
	ConfigurationMatchFiles = apiutil.ConfigDocumentedPath(configurationFileParts...)

	// ConfigurationFile is the default config file location.
	ConfigurationFile = apiutil.DefaultConfigPath(configurationFileParts...)

	// ConfigurationSearchFiles is the possible config file locations in descending priority order.
	ConfigurationSearchFiles = apiutil.ConfigActualPaths(configurationFileParts...)

	// UnevaluatedConfigurationSearchFiles are used for display.
	UnevaluatedSearchFiles = []string{
		strings.Join(configurationFileParts, "-"),
		filepath.Join("$XDG_CONFIG_HOME", filepath.Join(configurationFileParts...)),
		filepath.Join("/", "etc", filepath.Join(configurationFileParts...)),
	}

	// ConfigurationEnvPrefix is the prefix for configuration environment variables.
	ConfigurationEnvPrefix = "HOPS"

	// ConfigurationEnvName is the environment variable name that overrides the search paths.
	ConfigurationEnvName = "HOPS_CONFIG"
)

// Configuration represents the Hops CLI's configuration file.
type Configuration struct {
	// Path prefix for installed packages. Default value depends on OS/Arch.
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" env:"PREFIX"`

	// Path used for caches.
	Cache string `json:"cache,omitempty" yaml:"cache,omitempty" env:"CACHE" envDefault:"$XDG_CACHE_HOME/hops"`

	// Configuration shared from Homebrew.
	Homebrew brewenv.Configuration `json:"homebrew,omitempty" yaml:"homebrew,omitempty" envPrefix:"HOMEBREW_"`

	// Registry configures a Hops-compatible registry for Bottles.
	Registry RegistryConfig `json:"registry,omitempty" yaml:"registry,omitempty" envPrefix:"REGISTRY_"`
}

// RegistryConfig configures a Hops-compatible registry for Bottles.
type RegistryConfig struct {
	// Prefix is the prefix for all Bottle repositories
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" env:",inline"`

	// CAFile sets the server certificate authority file for the remote registry
	// CAFile string `json:"caFile,omitempty" yaml:"caFile,omitempty" env:"CA_FILE"`

	// DistributionSpec sets OCI distribution spec version and API option for target. options: v1.1-referrers-api, v1.1-referrers-tag
	DistributionSpec string `json:"distributionSpec,omitempty" yaml:"distributionSpec,omitempty" env:"DISTRIBUTION_SPEC"`

	// Headers adds custom headers to requests
	Headers []string `json:"headers,omitempty" yaml:"headers,omitempty" env:"HEADERS"`

	// Insecure	allows connections to SSL registry without certs
	// Insecure bool `json:"insecure,omitempty" yaml:"insecure,omitempty" env:"INSECURE"`

	// OCILayout sets the registry as an OCI image layout
	OCILayout bool `json:"ociLayout,omitempty" yaml:"ociLayout,omitempty" env:"OCI_LAYOUT"`

	// // Username sets the registry username
	// Username string
	// // Password sets the registry password
	// Password string

	// PlainHTTP allows insecure connections to registry without SSL check
	PlainHTTP bool `json:"plainHTTP,omitempty" yaml:"plainHTTP,omitempty" env:"PLAIN_HTTP"`

	// Config sets the path of the authentication file for the registry
	Config string `json:"config,omitempty" yaml:"config,omitempty" env:"CONFIG"`

	// Resolve sets customized DNS for registry, formatted in host:port:address[:address_port]
	// Resolve string `json:"resolve,omitempty" yaml:"resolve,omitempty" env:"RESOLVE"`
}

// ConfigurationDefault defaults the object's fields.
func ConfigurationDefault(cfg *Configuration) {
	if cfg.Prefix == "" {
		cfg.Prefix = prefix.Default().String()
	}

	if cfg.Cache == "" {
		cfg.Cache = filepath.Join(xdg.CacheHome, "hops")
	}

	if cfg.Registry.DistributionSpec == "" {
		cfg.Registry.DistributionSpec = "v1.1-referrers-api"
	}

	if cfg.Homebrew.API.AutoUpdate.Secs == nil {
		cfg.Homebrew.API.AutoUpdate.Secs = new(int)
		*(cfg.Homebrew.API.AutoUpdate.Secs) = brewenv.DefaultAutoUpdateSecs
	}

	if cfg.Homebrew.API.Domain == "" {
		cfg.Homebrew.API.Domain = "https://formulae.brew.sh/api"
	}

	// Do not default the registry prefix
	// ghcr.io/homebrew/core is not usable as a Hops registry
	// Could change this down the line if support is added for
	// "unsafe" installation without metadata or if Homebrew's
	// registry starts including the metadata in some way
	// if cfg.Registry.Prefix == "" {
	// 	cfg.Registry.Prefix = "ghcr.io/homebrew/core"
	// }
}

// ConfigurationEnvOverrides overrides the configuration with environment variables.
func ConfigurationEnvOverrides(cfg *Configuration) {
	cfg.Prefix = env.OneOfString([]string{
		ConfigurationEnvPrefix + "_PREFIX",
		"HOMEBREW_PREFIX",
	}, cfg.Prefix)

	cfg.Cache = env.String(
		ConfigurationEnvPrefix+"_CACHE",
		cfg.Cache)

	cfg.Homebrew.API.Domain = env.OneOfString([]string{
		ConfigurationEnvPrefix + "_HOMEBREW_API_DOMAIN",
		"HOMEBREW_API_DOMAIN",
	}, cfg.Homebrew.API.Domain)

	cfg.Homebrew.API.AutoUpdate.Disabled = env.OneOfOr([]string{
		ConfigurationEnvPrefix + "_HOMEBREW_NO_AUTO_UPDATE",
		"HOMEBREW_NO_AUTO_UPDATE",
	},
		cfg.Homebrew.API.AutoUpdate.Disabled,
		strconv.ParseBool)

	cfg.Homebrew.API.AutoUpdate.Secs = env.OneOfOr([]string{
		ConfigurationEnvPrefix + "_HOMEBREW_API_AUTO_UPDATE_SECS",
		"HOMEBREW_API_AUTO_UPDATE_SECS",
	},
		cfg.Homebrew.API.AutoUpdate.Secs, func(envVal string) (*int, error) {
			val, err := strconv.Atoi(envVal)
			return &val, err
		})

	cfg.Registry.Prefix = env.String(ConfigurationEnvPrefix+"_REGISTRY", cfg.Registry.Prefix)
	cfg.Registry.PlainHTTP = env.Bool(ConfigurationEnvPrefix+"_REGISTRY_PLAIN_HTTP", cfg.Registry.PlainHTTP)
}

// String implements fmt.Stringer.
func (cfg *Configuration) String() string {
	marshalled, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

// LogValuer implements slog.LogValuer.
func (cfg *Configuration) LogValue() slog.Value {
	b, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return slog.StringValue(string(b))
}

// ParseHeaders parses the configured HTTP headers.
func (cfg *RegistryConfig) ParseHeaders() (map[string][]string, error) {
	headers := map[string][]string{}
	for _, h := range cfg.Headers {
		name, value, found := strings.Cut(h, ":")
		if !found || strings.TrimSpace(name) == "" {
			// In conformance to the RFC 2616 specification
			// Reference: https://www.rfc-editor.org/rfc/rfc2616#section-4.2
			return nil, fmt.Errorf("invalid header: %q", h)
		}
		headers[name] = append(headers[name], value)
	}
	return headers, nil
}
