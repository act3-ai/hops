package brewenv

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"

	"github.com/act3-ai/hops/internal/apis/apiutil"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/utils/env"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

var (
	// configurationFileParts is the parts used to assemble
	// the config file name and default paths.
	configurationFileName = "brew.env"

	// ConfigurationMatchFiles is a list of patterns used to match a config file for schema validation.
	ConfigurationMatchFiles = apiutil.ConfigDocumentedPath(configurationFileName)

	// ConfigurationFile is the default config file location.
	ConfigurationFile = apiutil.DefaultConfigPath(configurationFileName)

	// UnevaluatedConfigurationSearchFiles are used for display.
	UnevaluatedSearchFiles = []string{
		PrefixEnvFile(),
		filepath.Join("$XDG_CONFIG_HOME", configurationFileName),
		filepath.Join(xdg.Home, ".homebrew"),
		SystemEnvFile(),
	}
)

// Configuration represents [Homebrew's environment configuration].
// Note that environment variables must have a value set to be detected. For
// example, run "export HOMEBREW_NO_INSECURE_REDIRECT=1" rather than just "export HOMEBREW_NO_INSECURE_REDIRECT".
//
// `HOMEBREW_*` environment variables can also be set in Homebrew’s environment
// files:
//
//   - /etc/homebrew/brew.env (system-wide)
//   - $HOMEBREW_PREFIX/etc/homebrew/brew.env` (prefix-specific)
//   - $XDG_CONFIG_HOME/homebrew/brew.env if $XDG_CONFIG_HOME is set or
//     $HOME/.homebrew/brew.env otherwise (user-specific)
//
// User-specific environment files take precedence over prefix-specific files and
// prefix-specific files take precedence over system-wide files (unless
// `HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY` is set, see below).
//
// Note that these files do not support shell variable expansion e.g. `$HOME` or
// command execution e.g. `$(cat file)`.
//
// [Homebrew's environment configuration]: https://docs.brew.sh/Manpage#environment
type Configuration struct {
	// Use this directory as the download cache.
	//
	// Defaults:
	//  - macOS: `$XDG_CACHE_HOME/Homebrew` or `$HOME/Library/Caches/Homebrew`
	//  - Linux: `$XDG_CACHE_HOME/Homebrew` or `$HOME/.cache/Homebrew`.
	Cache string `json:"cache,omitempty" yaml:"cache,omitempty" env:"CACHE"`

	// Prefix all download URLs, including those for bottles, with this value. For example, `HOMEBREW_ARTIFACT_DOMAIN=http://localhost:8080` will cause a formula with the URL `https://example.com/foo.tar.gz` to instead download from `http://localhost:8080/https://example.com/foo.tar.gz`. Bottle URLs however, have their domain replaced with this prefix. This results in e.g. `https://ghcr.io/v2/homebrew/core/gettext/manifests/0.21` to instead be downloaded from `http://localhost:8080/v2/homebrew/core/gettext/manifests/0.21`
	ArtifactDomain string `json:"artifactDomain,omitempty" yaml:"artifactDomain,omitempty" env:"ARTIFACT_DOMAIN"`

	// Configures Hops' usage of the Homebrew API.
	API APIConfig `json:"api,omitempty" yaml:"api,omitempty" envPrefix:"API_"`

	// If set, do not automatically update before running some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, run this less often by setting `HOMEBREW_AUTO_UPDATE_SECS` to a value higher than the default.
	// NoAutoUpdate bool `json:"HOMEBREW_NO_AUTO_UPDATE" env:"HOMEBREW_NO_AUTO_UPDATE"`

	// Run `brew update` once every `HOMEBREW_AUTO_UPDATE_SECS` seconds before some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, disable auto-update entirely with `HOMEBREW_NO_AUTO_UPDATE`.
	//
	// Default: `86400` (24 hours), `3600` (1 hour) if a developer command has been run or `300` (5 minutes) if `HOMEBREW_NO_INSTALL_FROM_API` is set.
	// AutoUpdateSecs int `json:"autoUpdateSecs,omitempty" yaml:"autoUpdateSecs,omitempty" env:"HOMEBREW_AUTO_UPDATE_SECS"`

	// Use this URL as the download mirror for bottles. If bottles at that URL are temporarily unavailable, the default bottle domain will be used as a fallback mirror. For example, setting BottleDomain to `HOMEBREW_BOTTLE_DOMAIN=http://localhost:8080` will cause all bottles to download from the prefix `http://localhost:8080/`.
	//
	// Default: `https://ghcr.io/v2/homebrew/core`.
	BottleDomain string `json:"bottleDomain,omitempty" yaml:"bottleDomain,omitempty" env:"BOTTLE_DOMAIN"`

	// Docker registry configuration
	DockerRegistry DockerRegistryConfig `json:"registry,omitempty" yaml:"registry,omitempty" envPrefix:"DOCKER_REGISTRY_"`
}

// DockerRegistryConfig configures docker registry usage.
type DockerRegistryConfig struct {
	// Use this base64 encoded username and password for authenticating with a Docker registry proxying GitHub Packages. If Token is set, it will be used instead.
	BasicAuthToken string `json:"basicAuthToken,omitempty" yaml:"basicAuthToken,omitempty" env:"BASIC_AUTH_TOKEN"`

	// Use this bearer token for authenticating with a Docker registry proxying GitHub Packages. Preferred over BasicAuthToken.
	Token string `json:"token,omitempty" yaml:"token,omitempty" env:"TOKEN"`
}

// GitHubPackagesAuth derives the GitHub Packages auth from the env config.
//
// From: https://github.com/Homebrew/brew/blob/master/Library/Homebrew/brew.sh
func (e *Configuration) GitHubPackagesAuth() string {
	switch {
	case e.DockerRegistry.Token != "":
		return "Bearer " + e.DockerRegistry.Token
	case e.DockerRegistry.BasicAuthToken != "":
		return "Bearer " + e.DockerRegistry.BasicAuthToken
	default:
		return "Bearer QQ=="
	}
}

// APIConfig configures Hops' usage of the Homebrew API.
type APIConfig struct {
	// Domain to use for the Homebrew API. Overrides Homebrew's HOMEBREW_API_DOMAIN value.
	//
	// Default: https://formulae.brew.sh/api
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty" env:"DOMAIN"`

	// Configure auto-update behavior
	AutoUpdate AutoUpdateConfig `json:"autoUpdate,omitempty" yaml:"autoUpdate,omitempty" envPrefix:"AUTO_UPDATE_"`
}

// AutoUpdateConfig configures auto-updates for a formula index.
type AutoUpdateConfig struct {
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty" env:"DISABLED"`
	Secs     *int `json:"secs,omitempty" yaml:"secs,omitempty" env:"SECS"`
}

// ShouldAutoUpdate returns true if an auto update should be run.
func (au *AutoUpdateConfig) ShouldAutoUpdate(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// Means file is unreadable, log the error
			slog.Info("checking cache age", logutil.ErrAttr(err), slog.String("path", file))
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

// UserEnvFile produces the system env file location.
func UserEnvFile() string {
	usercfgdir := env.String("XDG_CONFIG_HOME", filepath.Join(xdg.Home, ".homebrew"))
	return filepath.Join(usercfgdir, configurationFileName)
}

// PrefixEnvFile produces the prefix env file location.
func PrefixEnvFile() string {
	return filepath.Join(prefix.Default().String(), "etc", "homebrew", configurationFileName)
}

// SystemEnvFile produces the system env file location.
func SystemEnvFile() string {
	return filepath.Join("/", "etc", "homebrew", configurationFileName)
}

// SystemEnvTakesPriority reports whether the HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY variable is set.
// The setting indicates that the system env file should take priority over the prefix and user env files.
func SystemEnvTakesPriority() bool {
	return env.NotEmpty("HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY")
}

// DefaultEnvironmentFiles returns the default env files in increasing priority order.
func DefaultEnvFiles() []string {
	if SystemEnvTakesPriority() {
		return []string{
			PrefixEnvFile(),
			UserEnvFile(),
			SystemEnvFile(),
		}
	}
	return []string{
		SystemEnvFile(),
		PrefixEnvFile(),
		UserEnvFile(),
	}
}

// ConfigurationDefault defaults the object's fields.
func ConfigurationDefault(cfg *Configuration) {
	if cfg.Cache == "" {
		cfg.Cache = filepath.Join(xdg.CacheHome, "Homebrew")
	}

	if cfg.API.AutoUpdate.Secs == nil {
		cfg.API.AutoUpdate.Secs = new(int)
		*(cfg.API.AutoUpdate.Secs) = DefaultAutoUpdateSecs
	}

	if cfg.API.Domain == "" {
		cfg.API.Domain = "https://formulae.brew.sh/api"
	}
}

// // LoadEnvFiles loads Homebrew environment files so their variables are available to override configuration.
// func LoadEnvFiles() {
// 	var err error
// 	for _, envfile := range DefaultEnvFiles() {
// 		err = godotenv.Load(envfile)
// 		if err != nil && !errors.Is(err, os.ErrNotExist) {
// 			slog.Warn("loading environment file",
// 				slog.String("file", envfile),
// 				logutil.ErrAttr(err))
// 		}
// 	}
// }

// func ReadEnvFiles(envfiles ...string) map[string]string {
// 	vals, err := godotenv.Read(envfiles...)
// 	if err != nil && !errors.Is(err, os.ErrNotExist) {
// 		slog.Warn("loading environment file",
// 			slog.String("files", strings.Join(envfiles, ",")),
// 			logutil.ErrAttr(err))
// 	}
// 	return vals
// }

// func EnvValuesOverride(values map[string]string) func(cfg *NewEnvironment) {
// 	return func(cfg *NewEnvironment) {
// 		if values == nil {
// 			return
// 		}

// 		if cfg.API.Domain == "" {
// 			if v, ok := values["HOMEBREW_API_DOMAIN"]; ok {
// 				cfg.API.Domain = v
// 			}
// 		}

// 		if !cfg.API.AutoUpdate.Disabled {
// 			if v, ok := values["HOMEBREW_NO_AUTO_UPDATE"]; ok && v != "" {
// 				cfg.API.AutoUpdate.Disabled, _ = strconv.ParseBool(v)
// 			}
// 		}

// 		cfg.API.Domain = env.OneOfString([]string{
// 			ConfigurationEnvPrefix + "_API_DOMAIN",
// 			"HOMEBREW_API_DOMAIN",
// 		}, cfg.API.Domain)

// 		cfg.API.AutoUpdate.Disabled = env.OneOfOr([]string{
// 			ConfigurationEnvPrefix + "_API_AUTO_UPDATE_DISABLED",
// 			"HOMEBREW_NO_AUTO_UPDATE",
// 		},
// 			cfg.API.AutoUpdate.Disabled,
// 			strconv.ParseBool)

// 		if cfg.API.AutoUpdate.Secs == nil {
// 			if v, ok := values["HOMEBREW_API_AUTO_UPDATE_SECS"]; ok && v != "" {
// 				val, err := strconv.Atoi(v)
// 				if err != nil {
// 					slog.Warn("parsing HOMEBREW_API_AUTO_UPDATE_SECS", slog.String("value", v), logutil.ErrAttr(err))
// 				} else {
// 					cfg.API.AutoUpdate.Secs = &val
// 				}
// 			}
// 		}

// 		cfg.API.AutoUpdate.Secs = env.OneOfOr([]string{
// 			ConfigurationEnvPrefix + "_API_AUTO_UPDATE_SECS",
// 			"HOMEBREW_API_AUTO_UPDATE_SECS",
// 		},
// 			cfg.API.AutoUpdate.Secs, func(envVal string) (*int, error) {
// 				val, err := strconv.Atoi(envVal)
// 				return &val, err
// 			})
// 	}
// }
