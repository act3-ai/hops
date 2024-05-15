package v1beta1

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/xdg"

	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils"
	"github.com/act3-ai/hops/internal/utils/env"
)

var (
	// configurationFileParts is the parts used to assemble
	// the config file name and default paths
	configurationFileParts = []string{"hops", "config.yaml"}

	// ConfigurationMatchFiles is a list of patterns used to match a config file for schema validation
	ConfigurationMatchFiles = utils.ConfigValidatePath(configurationFileParts...)

	// ConfigurationFile is the default config file location
	ConfigurationFile = utils.DefaultConfigPath(configurationFileParts...)

	// ConfigurationSearchFiles is the possible config file locations in descending priority order
	ConfigurationSearchFiles = utils.ConfigSearchPath(configurationFileParts...)

	// UnevaluatedConfigurationSearchFiles are used for display
	UnevaluatedSearchFiles = []string{
		strings.Join(configurationFileParts, "-"),
		filepath.Join("$XDG_CONFIG_HOME", filepath.Join(configurationFileParts...)),
		filepath.Join("/", "etc", filepath.Join(configurationFileParts...)),
	}

	// ConfigurationEnvPrefix is the prefix for configuration environment variables
	ConfigurationEnvPrefix = "HOPS"

	// ConfigurationEnvName is the environment variable name that overrides the search paths
	ConfigurationEnvName = "HOPS_CONFIG"
)

// Configuration represents the Hops CLI's configuration file.
type Configuration struct {
	// Cache sets the path used for caches
	//
	// Default: $XDG_CACHE_HOME/hops
	Cache string `json:"cache,omitempty" yaml:"cache,omitempty"`

	// Configures Hops' usage of Homebrew's sources.
	Homebrew HomebrewAPIConfig `json:"homebrew,omitempty" yaml:"homebrew,omitempty"`

	// Registry sets the registry used for bottles
	Registry RegistryConfig `json:"registry,omitempty" yaml:"registry,omitempty"`
}

type RegistryConfig struct {
	// Prefix sets a Hops-compatible registry for bottles
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`

	// CAFile sets the server certificate authority file for the remote registry
	// CAFile string `json:"caFile,omitempty" yaml:"caFile,omitempty"`

	// DistributionSpec sets OCI distribution spec version and API option for target. options: v1.1-referrers-api, v1.1-referrers-tag
	// DistributionSpec string `json:"distributionSpec,omitempty" yaml:"distributionSpec,omitempty"`

	// Headers adds custom headers to requests
	// Headers []string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Insecure	allows connections to SSL registry without certs
	// Insecure bool `json:"insecure,omitempty" yaml:"insecure,omitempty"`

	// OCILayout sets the registry as an OCI image layout
	OCILayout bool `json:"ociLayout,omitempty" yaml:"ociLayout,omitempty"`

	// // Username sets the registry username
	// Username string
	// // Password sets the registry password
	// Password string

	// PlainHTTP allows insecure connections to registry without SSL check
	PlainHTTP bool `json:"plainHTTP,omitempty" yaml:"plainHTTP,omitempty"`

	// RegistryConfig sets the path of the authentication file for the registry
	// RegistryConfig string `json:"registryConfig,omitempty" yaml:"registryConfig,omitempty"`

	// Resolve sets customized DNS for registry, formatted in host:port:address[:address_port]
	// Resolve string `json:"resolve,omitempty" yaml:"resolve,omitempty"`

	/*
		oras flags:
		      --ca-file string                             server certificate authority file for the remote registry
		  -H, --header stringArray                         add custom headers to requests
		      --insecure                                   allow connections to SSL registry without certs
		      --oci-layout                                 set target as an OCI image layout
		  -p, --password string                            registry password or identity token
		      --plain-http                                 allow insecure connections to registry without SSL check
		      --registry-config path                       path of the authentication file for registry
		      --resolve host:port:address[:address_port]   customized DNS for registry, formatted in host:port:address[:address_port]
		  -u, --username string                            registry username
	*/
}

// AutoUpdateConfig configures auto-updates for a formula index.
type AutoUpdateConfig struct {
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Secs     *int `json:"secs,omitempty" yaml:"secs,omitempty"`
}

const DefaultAutoUpdateSecs = 86400

// HomebrewAPIConfig configures Hops' usage of the Homebrew API.
type HomebrewAPIConfig struct {
	// // Disables the Homebrew API. If enabled, the Homebrew API is queried for available formulae. Formulae defined in the indexes configured for Hops will supersede Homebrew's API.
	// //
	// // Default: false
	// Disabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`

	// Domain to use for the Homebrew API. Overrides Homebrew's HOMEBREW_API_DOMAIN value.
	//
	// Default: https://formulae.brew.sh/api
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty"`

	// Configure auto-update behavior
	AutoUpdate AutoUpdateConfig `json:"autoUpdate,omitempty" yaml:"autoUpdate,omitempty"`
}

// ConfigurationDefault defaults the object's fields
func ConfigurationDefault(cfg *Configuration) {
	if cfg.Cache == "" {
		cfg.Cache = filepath.Join(xdg.CacheHome, "hops")
	}

	if cfg.Homebrew.AutoUpdate.Secs == nil {
		cfg.Homebrew.AutoUpdate.Secs = new(int)
		*(cfg.Homebrew.AutoUpdate.Secs) = DefaultAutoUpdateSecs
	}

	if cfg.Homebrew.Domain == "" {
		cfg.Homebrew.Domain = "https://formulae.brew.sh/api"
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

// ConfigurationEnvOverrides overrides the configuration with environment variables
func ConfigurationEnvOverrides(cfg *Configuration) {
	cfg.Cache = env.String(
		ConfigurationEnvPrefix+"_CACHE",
		cfg.Cache)

	// cfg.Homebrew.Disabled = env.Bool(
	// 	ConfigurationEnvPrefix+"_API_DISABLED",
	// 	cfg.Homebrew.Disabled)

	cfg.Homebrew.Domain = env.OneOfString([]string{
		ConfigurationEnvPrefix + "_API_DOMAIN",
		"HOMEBREW_API_DOMAIN",
	}, cfg.Homebrew.Domain)

	cfg.Homebrew.AutoUpdate.Disabled = env.OneOfOr([]string{
		ConfigurationEnvPrefix + "_API_AUTOUPDATE_DISABLED",
		"HOMEBREW_NO_AUTO_UPDATE",
	},
		cfg.Homebrew.AutoUpdate.Disabled,
		strconv.ParseBool)

	cfg.Homebrew.AutoUpdate.Secs = env.OneOfOr([]string{
		ConfigurationEnvPrefix + "_API_AUTOUPDATE_DISABLED",
		"HOMEBREW_NO_AUTO_UPDATE",
	},
		cfg.Homebrew.AutoUpdate.Secs, func(envVal string) (*int, error) {
			val, err := strconv.Atoi(envVal)
			return &val, err
		})

	cfg.Registry.Prefix = env.String(ConfigurationEnvPrefix+"_REGISTRY", cfg.Registry.Prefix)
	cfg.Registry.PlainHTTP = env.Bool(ConfigurationEnvPrefix+"_REGISTRY_PLAIN_HTTP", cfg.Registry.PlainHTTP)
}

// String implements fmt.Stringer
func (cfg *Configuration) String() string {
	marshalled, err := json.MarshalIndent(cfg, "", "   ")
	if err != nil {
		panic(err)
	}
	return string(marshalled)
}

// // LogValuer implements slog.LogValuer
// func (cfg *Configuration) LogValue() slog.Value {
// 	b, err := json.Marshal(cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return slog.StringValue(string(b))
// }

// ShouldAutoUpdate returns true if an auto update should be run
func (au *AutoUpdateConfig) ShouldAutoUpdate(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Means file is unreadable, log the error
			slog.Info("checking cached index", o.ErrAttr(err), slog.String("path", file))
		}
		// Return true if file does not exist or was unreadable
		return true
	}

	// Return false if auto-updating is disabled
	// This is only obeyed once we know we have a formula index to use
	if au.Disabled {
		return false
	}

	// Return true if the index has not been updated in a
	// longer period of time than the user has configured
	return au.Secs != nil &&
		int(time.Since(info.ModTime()).Seconds()) >= *au.Secs
}

// func (cfg *RegistryConfig) ParseHeaders() (map[string][]string, error) {
// 	headers := map[string][]string{}
// 	for _, h := range cfg.Headers {
// 		name, value, found := strings.Cut(h, ":")
// 		if !found || strings.TrimSpace(name) == "" {
// 			// In conformance to the RFC 2616 specification
// 			// Reference: https://www.rfc-editor.org/rfc/rfc2616#section-4.2
// 			return nil, fmt.Errorf("invalid header: %q", h)
// 		}
// 		headers[name] = append(headers[name], value)
// 	}
// 	return headers, nil
// }
