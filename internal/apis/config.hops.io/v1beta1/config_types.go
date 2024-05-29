package v1beta1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
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
	Cache string `json:"cache,omitempty" yaml:"cache,omitempty" env:"CACHE"`

	// Configuration shared from Homebrew.
	Homebrew brewenv.Configuration `json:"homebrew,omitempty" yaml:"homebrew,omitempty" envPrefix:"HOMEBREW_"`

	// Registry configures a Hops-compatible registry for Bottles.
	Registry RegistryConfig `json:"registry,omitempty" yaml:"registry,omitempty" envPrefix:"REGISTRY_"`
}

// RegistryConfig configures a Hops-compatible registry for Bottles.
type RegistryConfig struct {
	// Prefix is the prefix for all Bottle repositories
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" env:",inline"`

	// DistributionSpec sets OCI distribution spec version and API option for target. options: v1.1-referrers-api, v1.1-referrers-tag
	// DistributionSpec string `json:"distributionSpec,omitempty" yaml:"distributionSpec,omitempty" env:"DISTRIBUTION_SPEC"`

	// Headers adds custom headers to requests
	Headers []string `json:"headers,omitempty" yaml:"headers,omitempty" env:"HEADERS"`

	// Insecure	allows connections to SSL registry without certs
	Insecure bool `json:"insecure,omitempty" yaml:"insecure,omitempty" env:"INSECURE"`

	// OCILayout sets the registry as an OCI image layout
	OCILayout bool `json:"ociLayout,omitempty" yaml:"ociLayout,omitempty" env:"OCI_LAYOUT"`

	// PlainHTTP allows insecure connections to registry without SSL check
	PlainHTTP bool `json:"plainHTTP,omitempty" yaml:"plainHTTP,omitempty" env:"PLAIN_HTTP"`

	// Config sets the path of the authentication file for the registry
	Config string `json:"config,omitempty" yaml:"config,omitempty" env:"CONFIG"`
}

// ConfigurationDefault defaults the object's fields.
func ConfigurationDefault(cfg *Configuration) {
	if cfg.Prefix == "" {
		cfg.Prefix = prefix.Default().String()
	}

	if cfg.Cache == "" {
		cfg.Cache = filepath.Join(xdg.CacheHome, "hops")
	}

	// Default Homebrew fields
	brewenv.ConfigurationDefault(&cfg.Homebrew)
}

// ConfigurationEnvOverrides overrides the configuration with environment variables.
func ConfigurationEnvOverrides(cfg *Configuration) {
	cfg.Prefix = env.OneOfString([]string{
		ConfigurationEnvPrefix + "_PREFIX",
		"HOMEBREW_PREFIX",
	}, cfg.Prefix)

	cfg.Cache = env.String(ConfigurationEnvPrefix+"_CACHE", cfg.Cache)

	// Override registry fields
	RegistryConfigEnvOverrides(ConfigurationEnvPrefix+"_REGISTRY", &cfg.Registry)

	// Override Homebrew fields
	brewenv.ConfigurationEnvOverrides(&cfg.Homebrew)
}

// RegistryConfigEnvOverrides overrides the configuration with environment variables.
func RegistryConfigEnvOverrides(envPrefix string, cfg *RegistryConfig) {
	cfg.Prefix = env.String(envPrefix, cfg.Prefix)
	cfg.Headers = env.StringSlice(envPrefix+"_HEADERS", cfg.Headers, ",")
	cfg.Insecure = env.Bool(envPrefix+"_INSECURE", cfg.Insecure)
	cfg.OCILayout = env.Bool(envPrefix+"_OCI_LAYOUT", cfg.OCILayout)
	cfg.PlainHTTP = env.Bool(envPrefix+"_PLAIN_HTTP", cfg.PlainHTTP)
	cfg.Config = env.String(envPrefix+"_CONFIG", cfg.Config)
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
