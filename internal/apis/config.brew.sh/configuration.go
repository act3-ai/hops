package brewenv

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/adrg/xdg"

	"github.com/act3-ai/hops/internal/utils/env"
	"github.com/act3-ai/hops/internal/utils/logutil"

	nfenv "github.com/Netflix/go-env"
)

// defaultset is the default values as an envmap
var defaultset = nfenv.EnvSet{
	"HOMEBREW_API_DOMAIN": Default.APIDomain,
	// "HOMEBREW_ARCH":                       Default.Arch,
	"HOMEBREW_API_AUTO_UPDATE_SECS": strconv.Itoa(Default.APIAutoUpdateSecs),
	"HOMEBREW_AUTO_UPDATE_SECS":     strconv.Itoa(Default.AutoUpdateSecs),
	"HOMEBREW_BOTTLE_DOMAIN":        Default.BottleDomain,
	// "HOMEBREW_BREW_GIT_REMOTE":            Default.BrewGitRemote,
	"HOMEBREW_CACHE":                      Default.Cache,
	"HOMEBREW_CLEANUP_MAX_AGE_DAYS":       strconv.Itoa(Default.CleanupMaxAgeDays),
	"HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS": strconv.Itoa(Default.CleanupPeriodicFullDays),
	// "HOMEBREW_CORE_GIT_REMOTE":            Default.CoreGitRemote,
	// "HOMEBREW_CURL_PATH":                  Default.CurlPath,
	// "HOMEBREW_CURL_RETRIES":               strconv.Itoa(Default.CurlRetries),
	// "HOMEBREW_FAIL_LOG_LINES":      strconv.Itoa(Default.FailLogLines),
	// "HOMEBREW_GIT_PATH":            Default.GitPath,
	"HOMEBREW_INSTALL_BADGE": Default.InstallBadge,
	// "HOMEBREW_LIVECHECK_WATCHLIST": Default.LivecheckWatchlist,
	// "HOMEBREW_LOGS":                Default.Logs,
	// "HOMEBREW_MAKE_JOBS":           strconv.Itoa(Default.MakeJobs),
	// "HOMEBREW_PIP_INDEX_URL":       Default.PIPIndexURL,
	// "HOMEBREW_SSH_CONFIG_PATH":     Default.SSHConfigPath,
	// "HOMEBREW_SVN":                 Default.SVN,
	// "HOMEBREW_TEMP":                Default.Temp,
}

// DefaultEnvironmentFiles returns the default files to load the environment config from
func DefaultEnvironmentFiles() []string {
	usercfgdir := env.String("XDG_CONFIG_HOME", filepath.Join(xdg.Home, ".homebrew"))

	// PrefixEnvFile is the location of the environment file in the prefix
	prefix := filepath.Join("etc", "homebrew", "brew.env")
	user := filepath.Join(usercfgdir, "brew.env")
	system := filepath.Join("/etc", "homebrew", "brew.env")

	if env.NotEmpty("HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY") {
		return []string{
			prefix,
			user,
			system,
		}
	}
	return []string{
		system,
		prefix,
		user,
	}
}

// Load loads the environment config from the OS environment
func Load() (*Environment, error) {
	// Limit to the envs we care about
	for k, v := range defaultset {
		if _, ok := os.LookupEnv(k); !ok {
			err := os.Setenv(k, v)
			if err != nil {
				return Default, err
			}
		}
	}

	e := &Environment{}
	_, err := nfenv.UnmarshalFromEnviron(e)
	if err != nil {
		return Default, fmt.Errorf("loading environment: %w", err)
	}

	// TODO: move these to the default EnvSet as functions
	if e.Developer {
		e.AutoUpdateSecs = env.Int("HOMEBREW_AUTO_UPDATE_SECS", 3600)
	}
	// if e.NoInstallFromAPI {
	// 	e.AutoUpdateSecs = env.Int("HOMEBREW_AUTO_UPDATE_SECS", 300)
	// }

	if e.DockerRegistryToken != "" {
		e.DockerRegistryBasicAuthToken = e.DockerRegistryToken
	}

	// unimplemented:
	// HOMEBREW_FORCE_BREWED_CA_CERTIFICATES
	// If set, always use a Homebrew-installed `ca-certificates` rather than the system version. Automatically set if the system version is too old.
	// HOMEBREW_FORCE_BREWED_CURL
	// If set, always use a Homebrew-installed `curl`(1) rather than the system version. Automatically set if the system version of `curl` is too old.
	// HOMEBREW_FORCE_BREWED_GIT
	// If set, always use a Homebrew-installed `git`(1) rather than the system version. Automatically set if the system version of `git` is too old.

	// if e.Developer {
	// 	// Default value flips if developer mode is on
	// 	e.SorbetRuntime = env.NotEmpty("HOMEBREW_SORBET_RUNTIME")
	// }

	return e, nil
}

// Environment represents [Homebrew's environment configuration]
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
type Environment struct {
	// HOMEBREW_PREFIX
	Prefix string `json:"HOMEBREW_PREFIX" env:"HOMEBREW_PREFIX"`

	// HOMEBREW_API_DOMAIN
	// Use this URL as the download mirror for Homebrew JSON API. If metadata files at that URL are temporarily unavailable, the default API domain will be used as a fallback mirror.
	//
	// Default: `https://formulae.brew.sh/api`.
	APIDomain string `json:"HOMEBREW_API_DOMAIN" env:"HOMEBREW_API_DOMAIN"`

	// HOMEBREW_ARCH
	// Linux only: Pass this value to a type name representing the compiler’s `-march` option.
	//
	//	Default: `native`.
	// Arch string `json:"HOMEBREW_ARCH" env:"HOMEBREW_ARCH"`

	// HOMEBREW_ARTIFACT_DOMAIN
	// Prefix all download URLs, including those for bottles, with this value. For example, `HOMEBREW_ARTIFACT_DOMAIN=http://localhost:8080` will cause a formula with the URL `https://example.com/foo.tar.gz` to instead download from `http://localhost:8080/https://example.com/foo.tar.gz`. Bottle URLs however, have their domain replaced with this prefix. This results in e.g. `https://ghcr.io/v2/homebrew/core/gettext/manifests/0.21` to instead be downloaded from `http://localhost:8080/v2/homebrew/core/gettext/manifests/0.21`
	ArtifactDomain string `json:"HOMEBREW_ARTIFACT_DOMAIN" env:"HOMEBREW_ARTIFACT_DOMAIN"`

	// HOMEBREW_API_AUTO_UPDATE_SECS
	// Check Homebrew’s API for new formulae or cask data every `HOMEBREW_API_AUTO_UPDATE_SECS` seconds. Alternatively, disable API auto-update checks entirely with `HOMEBREW_NO_AUTO_UPDATE`.
	//
	// Default: `450`.
	APIAutoUpdateSecs int `json:"HOMEBREW_API_AUTO_UPDATE_SECS" env:"HOMEBREW_API_AUTO_UPDATE_SECS"`

	// HOMEBREW_AUTO_UPDATE_SECS
	// Run `brew update` once every `HOMEBREW_AUTO_UPDATE_SECS` seconds before some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, disable auto-update entirely with `HOMEBREW_NO_AUTO_UPDATE`.
	//
	// Default: `86400` (24 hours), `3600` (1 hour) if a developer command has been run or `300` (5 minutes) if `HOMEBREW_NO_INSTALL_FROM_API` is set.
	AutoUpdateSecs int `json:"HOMEBREW_AUTO_UPDATE_SECS" env:"HOMEBREW_AUTO_UPDATE_SECS"`

	// HOMEBREW_AUTOREMOVE
	// If set, calls to `brew cleanup` and `brew uninstall` will automatically remove unused formula dependents and if `HOMEBREW_NO_INSTALL_CLEANUP` is not set, `brew cleanup` will start running `brew autoremove` periodically.
	// Autoremove bool `json:"HOMEBREW_AUTOREMOVE" env:"HOMEBREW_AUTOREMOVE"`

	// HOMEBREW_BAT
	// If set, use `bat` for the `brew cat` command.
	// Bat bool `json:"HOMEBREW_BAT" env:"HOMEBREW_BAT"`

	// HOMEBREW_BAT_CONFIG_PATH
	// Use this as the `bat` configuration file.
	//
	// Default: `$BAT_CONFIG_PATH`.
	// BatConfigPath string `json:"HOMEBREW_BAT_CONFIG_PATH" env:"HOMEBREW_BAT_CONFIG_PATH,BAT_CONFIG_PATH"`

	// HOMEBREW_BAT_THEME
	// Use this as the `bat` theme for syntax highlighting.
	//
	// Default: `$BAT_THEME`.
	// BatTheme string `json:"HOMEBREW_BAT_THEME" env:"HOMEBREW_BAT_THEME,BAT_THEME"`

	// HOMEBREW_BOOTSNAP
	// If set, use Bootsnap to speed up repeated `brew` calls. A no-op when using Homebrew’s vendored, relocatable Ruby on macOS (as it doesn’t work).
	// Bootsnap bool `json:"HOMEBREW_BOOTSNAP" env:"HOMEBREW_BOOTSNAP"`

	// HOMEBREW_BOTTLE_DOMAIN
	// Use this URL as the download mirror for bottles. If bottles at that URL are temporarily unavailable, the default bottle domain will be used as a fallback mirror. For example, `HOMEBREW_BOTTLE_DOMAIN=http://localhost:8080` will cause all bottles to download from the prefix `http://localhost:8080/`. If bottles are not available at `HOMEBREW_BOTTLE_DOMAIN` they will be downloaded from the default bottle domain.
	//
	// Default: `https://ghcr.io/v2/homebrew/core`.
	BottleDomain string `json:"HOMEBREW_BOTTLE_DOMAIN" env:"HOMEBREW_BOTTLE_DOMAIN"`

	// HOMEBREW_BREW_GIT_REMOTE
	// Use this URL as the Homebrew/brew `git`(1) remote.
	//
	// Default: `https://github.com/Homebrew/brew`.
	// BrewGitRemote string `json:"HOMEBREW_BREW_GIT_REMOTE" env:"HOMEBREW_BREW_GIT_REMOTE"`

	// HOMEBREW_BROWSER
	// Use this as the browser when opening project homepages.
	//
	// Default: `$BROWSER` or the OS’s default browser.
	// Browser string `json:"HOMEBREW_BROWSER" env:"HOMEBREW_BROWSER,BROWSER"`

	// HOMEBREW_CACHE
	// Use this directory as the download cache.
	//
	// Defaults:
	//  - macOS: `$XDG_CACHE_HOME/Homebrew` or `$HOME/Library/Caches/Homebrew`
	//  - Linux: `$XDG_CACHE_HOME/Homebrew` or `$HOME/.cache/Homebrew`.
	Cache string `json:"HOMEBREW_CACHE" env:"HOMEBREW_CACHE"`

	// HOMEBREW_CASK_OPTS
	// Append these options to all `cask` commands. All `--*dir` options, `--language`, `--require-sha`, `--no-quarantine` and `--no-binaries` are supported. For example, you might add something like the following to your `~/.profile`, `~/.bash_profile`, or `~/.zshenv`:
	//
	//     `export HOMEBREW_CASK_OPTS="--appdir=~/Applications --fontdir=/Library/Fonts"`
	// CaskOpts string `json:"HOMEBREW_CASK_OPTS" env:"HOMEBREW_CASK_OPTS"`

	// HOMEBREW_CLEANUP_MAX_AGE_DAYS
	// Cleanup all cached files older than this many days.
	//
	// Default: `120`.
	CleanupMaxAgeDays int `json:"HOMEBREW_CLEANUP_MAX_AGE_DAYS" env:"HOMEBREW_CLEANUP_MAX_AGE_DAYS"`

	// HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS
	// If set, `brew install`, `brew upgrade` and `brew reinstall` will cleanup all formulae when this number of days has passed.
	//
	// Default: `30`.
	CleanupPeriodicFullDays int `json:"HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS" env:"HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS"`

	// HOMEBREW_COLOR
	// If set, force colour output on non-TTY outputs.
	// Color bool `json:"HOMEBREW_COLOR" env:"HOMEBREW_COLOR"`

	// HOMEBREW_CORE_GIT_REMOTE
	// Use this URL as the Homebrew/homebrew-core `git`(1) remote.
	//
	// Default: `https://github.com/Homebrew/homebrew-core`.
	// CoreGitRemote string `json:"HOMEBREW_CORE_GIT_REMOTE" env:"HOMEBREW_CORE_GIT_REMOTE"`

	// HOMEBREW_CURL_PATH
	// Linux only: Set this value to a new enough `curl` executable for Homebrew to use.
	//
	// Default: `curl`.
	// CurlPath string `json:"HOMEBREW_CURL_PATH" env:"HOMEBREW_CURL_PATH"`

	// HOMEBREW_CURL_RETRIES
	// Pass the given retry count to `--retry` when invoking `curl`(1).
	//
	// Default: `3`.
	// CurlRetries int `json:"HOMEBREW_CURL_RETRIES" env:"HOMEBREW_CURL_RETRIES"`

	// HOMEBREW_CURL_VERBOSE
	// If set, pass `--verbose` when invoking `curl`(1).
	// CurlVerbose bool `json:"HOMEBREW_CURL_VERBOSE" env:"HOMEBREW_CURL_VERBOSE"`

	// HOMEBREW_CURLRC
	// If set to an absolute path (i.e. beginning with `/`), pass it with `--config` when invoking `curl`(1). If set but **not** a valid path, do not pass `--disable`, which disables the use of `.curlrc`.
	// Curlrc string `json:"HOMEBREW_CURLRC" env:"HOMEBREW_CURLRC"`

	// HOMEBREW_DEBUG
	// If set, always assume `--debug` when running commands.
	Debug bool `json:"HOMEBREW_DEBUG" env:"HOMEBREW_DEBUG"`

	// HOMEBREW_DEVELOPER
	// If set, tweak behaviour to be more relevant for Homebrew developers (active or budding) by e.g. turning warnings into errors.
	Developer bool `json:"HOMEBREW_DEVELOPER" env:"HOMEBREW_DEVELOPER"`

	// HOMEBREW_DISABLE_LOAD_FORMULA
	// If set, refuse to load formulae. This is useful when formulae are not trusted (such as in pull requests).
	// DisableLoadFormula bool `json:"HOMEBREW_DISABLE_LOAD_FORMULA" env:"HOMEBREW_DISABLE_LOAD_FORMULA"`

	// HOMEBREW_DISPLAY
	// Use this X11 display when opening a page in a browser, for example with `brew home`. Primarily useful on Linux.
	//
	// Default: `$DISPLAY`.
	// Display string `json:"HOMEBREW_DISPLAY" env:"HOMEBREW_DISPLAY,DISPLAY"`

	// HOMEBREW_DISPLAY_INSTALL_TIMES
	// If set, print install times for each formula at the end of the run.
	// DisplayInstallTimes bool `json:"HOMEBREW_DISPLAY_INSTALL_TIMES" env:"HOMEBREW_DISPLAY_INSTALL_TIMES"`

	// HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN
	// Use this base64 encoded username and password for authenticating with a Docker registry proxying GitHub Packages. If `HOMEBREW_DOCKER_REGISTRY_TOKEN` is set, it will be used instead.
	DockerRegistryBasicAuthToken string `json:"HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN" env:"HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN"`

	// HOMEBREW_DOCKER_REGISTRY_TOKEN
	// Use this bearer token for authenticating with a Docker registry proxying GitHub Packages. Preferred over `HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN`.
	DockerRegistryToken string `json:"HOMEBREW_DOCKER_REGISTRY_TOKEN" env:"HOMEBREW_DOCKER_REGISTRY_TOKEN"`

	// HOMEBREW_EDITOR
	// Use this editor when editing a single formula, or several formulae in the same directory.
	//
	// **Note:** `brew edit` will open all of Homebrew as discontinuous files and directories. Visual Studio Code can handle this correctly in project mode, but many editors will do strange things in this case.
	//
	// Default: `$EDITOR` or `$VISUAL`.
	// Editor string `json:"HOMEBREW_EDITOR" env:"HOMEBREW_EDITOR,EDITOR,VISUAL"`

	// HOMEBREW_EVAL_ALL
	// If set, `brew` commands evaluate all formulae and casks, executing their arbitrary code, by default without requiring `--eval-all`. Required to cache formula and cask descriptions.
	// HomebrewEvalAll bool `json:"HOMEBREW_EVAL_ALL" env:"HOMEBREW_EVAL_ALL"`

	// HOMEBREW_FAIL_LOG_LINES
	// Output this many lines of output on formula `system` failures.
	//
	// Default: `15`.
	// FailLogLines int `json:"HOMEBREW_FAIL_LOG_LINES" env:"HOMEBREW_FAIL_LOG_LINES"`

	// HOMEBREW_FORBIDDEN_LICENSES
	// A space-separated list of licenses. Homebrew will refuse to install a formula if it or any of its dependencies has a license on this list.
	// ForbiddenLicenses env.SpaceSlice `json:"HOMEBREW_FORBIDDEN_LICENSES" env:"HOMEBREW_FORBIDDEN_LICENSES"`

	// HOMEBREW_FORCE_BREWED_CA_CERTIFICATES
	// If set, always use a Homebrew-installed `ca-certificates` rather than the system version. Automatically set if the system version is too old.
	// ForceBrewedCACertificates bool `json:"HOMEBREW_FORCE_BREWED_CA_CERTIFICATES" env:"HOMEBREW_FORCE_BREWED_CA_CERTIFICATES"`

	// HOMEBREW_FORCE_BREWED_CURL
	// If set, always use a Homebrew-installed `curl`(1) rather than the system version. Automatically set if the system version of `curl` is too old.
	// ForceBrewedCurl bool `json:"HOMEBREW_FORCE_BREWED_CURL" env:"HOMEBREW_FORCE_BREWED_CURL"`

	// HOMEBREW_FORCE_BREWED_GIT
	// If set, always use a Homebrew-installed `git`(1) rather than the system version. Automatically set if the system version of `git` is too old.
	// ForceBrewedGit bool `json:"HOMEBREW_FORCE_BREWED_GIT" env:"HOMEBREW_FORCE_BREWED_GIT"`

	// HOMEBREW_FORCE_VENDOR_RUBY
	// If set, always use Homebrew’s vendored, relocatable Ruby version even if the system version of Ruby is new enough.
	// ForceVendorRuby bool `json:"HOMEBREW_FORCE_VENDOR_RUBY" env:"HOMEBREW_FORCE_VENDOR_RUBY"`

	// HOMEBREW_GIT_EMAIL
	// Set the Git author and committer email to this value.
	// GitEmail string `json:"HOMEBREW_GIT_EMAIL" env:"HOMEBREW_GIT_EMAIL"`

	// HOMEBREW_GIT_NAME
	// Set the Git author and committer name to this value.
	// GitName string `json:"HOMEBREW_GIT_NAME" env:"HOMEBREW_GIT_NAME"`

	// HOMEBREW_GIT_PATH
	// Linux only: Set this value to a new enough `git` executable for Homebrew to use.
	//
	// Default: `git`.
	// GitPath string `json:"HOMEBREW_GIT_PATH" env:"HOMEBREW_GIT_PATH"`

	// HOMEBREW_GITHUB_API_TOKEN
	// Use this personal access token for the GitHub API, for features such as `brew search`. You can create one at <https://github.com/settings/tokens>. If set, GitHub will allow you a greater number of API requests. For more information, see: <https://docs.github.com/en/rest/overview/rate-limits-for-the-rest-api>
	//
	//     **Note:** Homebrew doesn’t require permissions for any of the scopes, but some developer commands may require additional permissions.
	// GitHubAPIToken string `json:"HOMEBREW_GITHUB_API_TOKEN" env:"HOMEBREW_GITHUB_API_TOKEN"`

	// HOMEBREW_GITHUB_PACKAGES_TOKEN
	// Use this GitHub personal access token when accessing the GitHub Packages Registry (where bottles may be stored).
	// GitHubPackagesToken string `json:"HOMEBREW_GITHUB_PACKAGES_TOKEN" env:"HOMEBREW_GITHUB_PACKAGES_TOKEN"`

	// HOMEBREW_GITHUB_PACKAGES_USER
	// Use this username when accessing the GitHub Packages Registry (where bottles may be stored).
	// GitHubPackagesUser string `json:"HOMEBREW_GITHUB_PACKAGES_USER" env:"HOMEBREW_GITHUB_PACKAGES_USER"`

	// HOMEBREW_INSTALL_BADGE
	// Print this text before the installation summary of each successful build.
	//
	// Default: The “Beer Mug” emoji.
	InstallBadge string `json:"HOMEBREW_INSTALL_BADGE" env:"HOMEBREW_INSTALL_BADGE"`

	// HOMEBREW_LIVECHECK_WATCHLIST
	// Consult this file for the list of formulae to check by default when no formula argument is passed to `brew livecheck`.
	//
	// Default: `$XDG_CONFIG_HOME/homebrew/livecheck_watchlist.txt` if `$XDG_CONFIG_HOME` is set or `$HOME/.homebrew/livecheck_watchlist.txt` otherwise.
	// LivecheckWatchlist string `json:"HOMEBREW_LIVECHECK_WATCHLIST" env:"HOMEBREW_LIVECHECK_WATCHLIST"`

	// HOMEBREW_LOGS
	// Use this directory to store log files.
	//
	// Default: macOS: `$HOME/Library/Logs/Homebrew`, Linux: `$XDG_CACHE_HOME/Homebrew/Logs` or `$HOME/.cache/Homebrew/Logs`.
	// Logs string `json:"HOMEBREW_LOGS" env:"HOMEBREW_LOGS"`

	// HOMEBREW_MAKE_JOBS
	// Use this value as the number of parallel jobs to run when building with `make`(1).
	//
	// Default: The number of available CPU cores.
	// MakeJobs int `json:"HOMEBREW_MAKE_JOBS" env:"HOMEBREW_MAKE_JOBS"`

	// HOMEBREW_NO_ANALYTICS
	// If set, do not send analytics. Google Analytics were destroyed. For more information, see: <https://docs.brew.sh/Analytics>
	// NoAnalytics bool `json:"HOMEBREW_NO_ANALYTICS" env:"HOMEBREW_NO_ANALYTICS"`

	// HOMEBREW_NO_AUTO_UPDATE
	// If set, do not automatically update before running some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, run this less often by setting `HOMEBREW_AUTO_UPDATE_SECS` to a value higher than the default.
	// NoAutoUpdate bool `json:"HOMEBREW_NO_AUTO_UPDATE" env:"HOMEBREW_NO_AUTO_UPDATE"`

	// HOMEBREW_NO_BOOTSNAP
	// If set, do not use Bootsnap to speed up repeated `brew` calls.
	// NoBootsnap bool `json:"HOMEBREW_NO_BOOTSNAP" env:"HOMEBREW_NO_BOOTSNAP"`

	// HOMEBREW_NO_CLEANUP_FORMULAE
	// A comma-separated list of formulae. Homebrew will refuse to clean up or autoremove a formula if it appears on this list.
	// NoCleanupFormulae env.CommaSlice `json:"HOMEBREW_NO_CLEANUP_FORMULAE" env:"HOMEBREW_NO_CLEANUP_FORMULAE"`

	// HOMEBREW_NO_COLOR
	// If set, do not print text with colour added.
	//
	// Default: `$NO_COLOR`.
	// NoColor bool `json:"HOMEBREW_NO_COLOR" env:"HOMEBREW_NO_COLOR,NO_COLOR"`

	// HOMEBREW_NO_EMOJI
	// If set, do not print `HOMEBREW_INSTALL_BADGE` on a successful build.
	NoEmoji bool `json:"HOMEBREW_NO_EMOJI" env:"HOMEBREW_NO_EMOJI"`

	// HOMEBREW_NO_ENV_HINTS
	// If set, do not print any hints about changing Homebrew’s behaviour with environment variables.
	// NoEnvHints bool `json:"HOMEBREW_NO_ENV_HINTS" env:"HOMEBREW_NO_ENV_HINTS"`

	// HOMEBREW_NO_GITHUB_API
	// If set, do not use the GitHub API, e.g. for searches or fetching relevant issues after a failed install.
	// NoGitHubAPI bool `json:"HOMEBREW_NO_GITHUB_API" env:"HOMEBREW_NO_GITHUB_API"`

	// HOMEBREW_NO_INSECURE_REDIRECT
	// If set, forbid redirects from secure HTTPS to insecure HTTP.
	//
	//     **Note:** while ensuring your downloads are fully secure, this is likely to cause from-source SourceForge, some GNU &amp; GNOME-hosted formulae to fail to download.
	// NoInsecureRedirect bool `json:"HOMEBREW_NO_INSECURE_REDIRECT" env:"HOMEBREW_NO_INSECURE_REDIRECT"`

	// HOMEBREW_NO_INSTALL_CLEANUP
	// If set, `brew install`, `brew upgrade` and `brew reinstall` will never automatically cleanup installed/upgraded/reinstalled formulae or all formulae every `HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS` days. Alternatively, `HOMEBREW_NO_CLEANUP_FORMULAE` allows specifying specific formulae to not clean up.
	// NoInstallCleanup bool `json:"HOMEBREW_NO_INSTALL_CLEANUP" env:"HOMEBREW_NO_INSTALL_CLEANUP"`

	// HOMEBREW_NO_INSTALL_FROM_API
	// If set, do not install formulae and casks in homebrew/core and homebrew/cask taps using Homebrew’s API and instead use (large, slow) local checkouts of these repositories.
	// NoInstallFromAPI bool `json:"HOMEBREW_NO_INSTALL_FROM_API" env:"HOMEBREW_NO_INSTALL_FROM_API"`

	// HOMEBREW_NO_INSTALL_UPGRADE
	// If set, `brew install` **`formula|cask`** will not upgrade **`formula|cask`** if it is installed but outdated.
	NoInstallUpgrade bool `json:"HOMEBREW_NO_INSTALL_UPGRADE" env:"HOMEBREW_NO_INSTALL_UPGRADE"`

	// HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK
	// If set, do not check for broken linkage of dependents or outdated dependents after installing, upgrading or reinstalling formulae. This will result in fewer dependents (and their dependencies) being upgraded or reinstalled but may result in more breakage from running `brew install` **`formula`** or `brew upgrade` **`formula`**.
	NoInstalledDependentsCheck bool `json:"HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK" env:"HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK"`

	// HOMEBREW_NO_UPDATE_REPORT_NEW
	// If set, `brew update` will not show the list of newly added formulae/casks.
	NoUpdateReportNew bool `json:"HOMEBREW_NO_UPDATE_REPORT_NEW" env:"HOMEBREW_NO_UPDATE_REPORT_NEW"`

	// HOMEBREW_PIP_INDEX_URL
	// If set, `brew install` **`formula`** will use this URL to download PyPI package resources.
	//
	// Default: `https://pypi.org/simple`.
	// PIPIndexURL string `json:"HOMEBREW_PIP_INDEX_URL" env:"HOMEBREW_PIP_INDEX_URL"`

	// HOMEBREW_PRY
	// If set, use Pry for the `brew irb` command.
	// Pry bool `json:"HOMEBREW_PRY" env:"HOMEBREW_PRY"`

	// HOMEBREW_UPGRADE_GREEDY
	// If set, pass `--greedy` to all cask upgrade commands.
	// UpgradeGreedy bool `json:"HOMEBREW_UPGRADE_GREEDY" env:"HOMEBREW_UPGRADE_GREEDY"`

	// HOMEBREW_SIMULATE_MACOS_ON_LINUX
	// If set, running Homebrew on Linux will simulate certain macOS code paths. This is useful when auditing macOS formulae while on Linux.
	// SimulateMacOSOnLinux bool `json:"HOMEBREW_SIMULATE_MACOS_ON_LINUX" env:"HOMEBREW_SIMULATE_MACOS_ON_LINUX"`

	// HOMEBREW_SKIP_OR_LATER_BOTTLES
	// If set along with `HOMEBREW_DEVELOPER`, do not use bottles from older versions of macOS. This is useful in development on new macOS versions.
	// SkipOrLaterBottles bool `json:"HOMEBREW_SKIP_OR_LATER_BOTTLES" env:"HOMEBREW_SKIP_OR_LATER_BOTTLES"`

	// HOMEBREW_SORBET_RUNTIME
	// If set, enable runtime typechecking using Sorbet. Set by default for `HOMEBREW_DEVELOPER` or when running some developer commands.
	// SorbetRuntime bool `json:"HOMEBREW_SORBET_RUNTIME" env:"HOMEBREW_SORBET_RUNTIME"`

	// HOMEBREW_SSH_CONFIG_PATH
	// If set, Homebrew will use the given config file instead of `~/.ssh/config` when fetching Git repositories over SSH.
	//
	// Default: `$HOME/.ssh/config`
	// SSHConfigPath string `json:"HOMEBREW_SSH_CONFIG_PATH" env:"HOMEBREW_SSH_CONFIG_PATH"`

	// HOMEBREW_SVN
	// Use this as the `svn`(1) binary.
	//
	// Default: A Homebrew-built Subversion (if installed), or the system-provided binary.
	// SVN string `json:"HOMEBREW_SVN" env:"HOMEBREW_SVN"`

	// HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY
	// If set in Homebrew’s system-wide environment file (`/etc/homebrew/brew.env`), the system-wide environment file will be loaded last to override any prefix or user settings.
	SystemEnvTakesPriority bool `json:"HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY" env:"HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY"`

	// HOMEBREW_SUDO_THROUGH_SUDO_USER
	// If set, Homebrew will use the `SUDO_USER` environment variable to define the user to `sudo`(8) through when running `sudo`(8).
	// SudoThroughSudoUser bool `json:"HOMEBREW_SUDO_THROUGH_SUDO_USER" env:"HOMEBREW_SUDO_THROUGH_SUDO_USER"`

	// HOMEBREW_TEMP
	// Use this path as the temporary directory for building packages. Changing this may be needed if your system temporary directory and Homebrew prefix are on different volumes, as macOS has trouble moving symlinks across volumes when the target does not yet exist. This issue typically occurs when using FileVault or custom SSD configurations.
	//
	// Default: macOS: `/private/tmp`, Linux: `/tmp`.
	// Temp string `json:"HOMEBREW_TEMP" env:"HOMEBREW_TEMP"`

	// HOMEBREW_UPDATE_TO_TAG
	// If set, always use the latest stable tag (even if developer commands have been run).
	// UpdateToTag bool `json:"HOMEBREW_UPDATE_TO_TAG" env:"HOMEBREW_UPDATE_TO_TAG"`

	// HOMEBREW_VERBOSE
	// If set, always assume `--verbose` when running commands.
	Verbose bool `json:"HOMEBREW_VERBOSE" env:"HOMEBREW_VERBOSE"`

	// HOMEBREW_VERBOSE_USING_DOTS
	// If set, verbose output will print a `.` no more than once a minute. This can be useful to avoid long-running Homebrew commands being killed due to no output.
	// VerboseUsingDots bool `json:"HOMEBREW_VERBOSE_USING_DOTS" env:"HOMEBREW_VERBOSE_USING_DOTS"`

	// SUDO_ASKPASS
	// If set, pass the `-A` option when calling `sudo`(8).
	// SudoAskpass bool `json:"SUDO_ASKPASS" env:"SUDO_ASKPASS"`

	// all_proxy
	// Use this SOCKS5 proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
	// AllProxy string `json:"all_proxy" env:"all_proxy"`

	// ftp_proxy
	// Use this FTP proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
	// FTPProxy string `json:"ftp_proxy" env:"ftp_proxy"`

	// http_proxy
	// Use this HTTP proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
	// HTTPProxy string `json:"http_proxy" env:"http_proxy"`

	// https_proxy
	// Use this HTTPS proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
	// HTTPSProxy string `json:"https_proxy" env:"https_proxy"`

	// no_proxy
	// A comma-separated list of hostnames and domain names excluded from proxying by `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
	// NoProxy env.CommaSlice `json:"no_proxy" env:"no_proxy"`
}

func (e *Environment) String() string {
	b, err := nfenv.Marshal(e)
	if err != nil {
		slog.Error("marshalling environment config: %w", logutil.ErrAttr(err))
		panic(err)
		// return ""
	}

	vals := nfenv.EnvSetToEnviron(b)
	slices.Sort(vals)

	return strings.Join(vals, "\n")
}

// unused:
// func defaultLivecheckWatchlist() string {
// 	if _, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
// 		return filepath.Join(xdg.ConfigHome, "homebrew", "livecheck_watchlist.txt")
// 	}
// 	return filepath.Join(xdg.Home, ".homebrew", "livecheck_watchlist.txt")
// }

// unused:
// func defaultLogs() string {
// 	_, ok := os.LookupEnv("XDG_CACHE_HOME")
// 	if runtime.GOOS == "darwin" && !ok {
// 		return filepath.Join(xdg.Home, "Library", "Logs", "Homebrew")
// 	}
// 	return filepath.Join(xdg.CacheHome, "Homebrew", "logs")
// }

// unused:
// func defaultTemp() string {
// 	if runtime.GOOS == "darwin" {
// 		return "/private/tmp"
// 	}
// 	return "/tmp"
// }

// GitHubPackagesAuth derives the GitHub Packages auth from the env config
//
// From: https://github.com/Homebrew/brew/blob/master/Library/Homebrew/brew.sh
//
//	if [[ -n "${HOMEBREW_DOCKER_REGISTRY_TOKEN}" ]]
//	then
//	  export HOMEBREW_GITHUB_PACKAGES_AUTH="Bearer ${HOMEBREW_DOCKER_REGISTRY_TOKEN}"
//	elif [[ -n "${HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN}" ]]
//	then
//	  export HOMEBREW_GITHUB_PACKAGES_AUTH="Basic ${HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN}"
//	else
//	  export HOMEBREW_GITHUB_PACKAGES_AUTH="Bearer QQ=="
//	fi
func (e *Environment) GitHubPackagesAuth() string {
	if e.DockerRegistryToken != "" {
		return "Bearer " + e.DockerRegistryToken
	}
	if e.DockerRegistryBasicAuthToken != "" {
		return "Bearer " + e.DockerRegistryBasicAuthToken
	}
	return "Bearer QQ=="
}

// AddGitHubPackagesAuthHeader adds the GitHub Packages auth header to the request
func (e *Environment) AddGitHubPackagesAuthHeader(req http.Request) {
	req.Header.Set("Authorization", e.GitHubPackagesAuth())
}

// Default is the default values
var Default = &Environment{
	APIDomain: "https://formulae.brew.sh/api",
	// Arch:                         "native",
	ArtifactDomain:    "",
	APIAutoUpdateSecs: 450,
	AutoUpdateSecs:    86400,
	// Autoremove:                   false,
	// Bat:                          false,
	// BatConfigPath:                "",
	// BatTheme:                     "",
	// Bootsnap:                     false,
	BottleDomain: "https://ghcr.io/v2/homebrew/core",
	// BrewGitRemote:                "https://github.com/Homebrew/brew",
	// Browser:                      "",
	Cache: filepath.Join(xdg.CacheHome, "Homebrew"),
	// CaskOpts:                     "",
	CleanupMaxAgeDays:       120,
	CleanupPeriodicFullDays: 30,
	// Color:                        false,
	// CoreGitRemote:                "https://github.com/Homebrew/homebrew-core",
	// CurlPath:                     "curl",
	// CurlRetries:                  3,
	// CurlVerbose:                  false,
	// Curlrc:                       "",
	Debug:     false,
	Developer: false,
	// DisableLoadFormula:           false,
	// Display:                      "",
	// DisplayInstallTimes:          false,
	DockerRegistryBasicAuthToken: "",
	DockerRegistryToken:          "",
	// Editor:                       "",
	// HomebrewEvalAll:              false,
	// FailLogLines:                 15,
	// ForbiddenLicenses:          []string{},
	// ForceBrewedCACertificates:  false,
	// ForceBrewedCurl:            false,
	// ForceBrewedGit:             false,
	// ForceVendorRuby:            false,
	// GitEmail:                   "",
	// GitName:                    "",
	// GitPath:                    "git",
	// GitHubAPIToken:             "",
	// GitHubPackagesToken:        "",
	// GitHubPackagesUser:         "",
	InstallBadge: "🌼", // env.Or("HOMEBREW_INSTALL_BADGE", "🍺",
	// LivecheckWatchlist:         defaultLivecheckWatchlist(),
	// Logs:                       defaultLogs(),
	// MakeJobs:                   runtime.NumCPU(),
	// NoAnalytics:                false,
	// NoAutoUpdate:               false,
	// NoBootsnap:                 false,
	// NoCleanupFormulae:          []string{},
	// NoColor:                    false,
	NoEmoji: false,
	// NoEnvHints:                 false,
	// NoGitHubAPI:                false,
	// NoInsecureRedirect:         false,
	// NoInstallCleanup:           false,
	// NoInstallFromAPI:           false,
	NoInstallUpgrade:           false,
	NoInstalledDependentsCheck: false,
	NoUpdateReportNew:          false,
	// PIPIndexURL:                "https://pypi.org/simple",
	// Pry:                        false,
	// UpgradeGreedy:              false,
	// SimulateMacOSOnLinux:       false,
	// SkipOrLaterBottles:         false,
	// SorbetRuntime:              false,
	// SSHConfigPath:              filepath.Join(xdg.Home, ".ssh", "config"),
	// SVN:                        "svn",
	SystemEnvTakesPriority: false,
	// SudoThroughSudoUser:    false,
	// Temp:                   defaultTemp(),
	// UpdateToTag:            false,
	Verbose: false,
	// VerboseUsingDots:       false,
	// SudoAskpass:            false,
	// AllProxy:               "",
	// FTPProxy:               "",
	// HTTPProxy:              "",
	// HTTPSProxy:             "",
	// NoProxy:                []string{},
}
