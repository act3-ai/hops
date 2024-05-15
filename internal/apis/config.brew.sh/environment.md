# HOMEBREW ENVIRONMENT

> Source: <https://docs.brew.sh/Manpage#environment>

Note that environment variables must have a value set to be detected. For
example, run `export HOMEBREW_NO_INSECURE_REDIRECT=1` rather than just `export HOMEBREW_NO_INSECURE_REDIRECT`.

`HOMEBREW_*` environment variables can also be set in Homebrew’s environment
files:

- `/etc/homebrew/brew.env` (system-wide)
- `$HOMEBREW_PREFIX/etc/homebrew/brew.env` (prefix-specific)
- `$XDG_CONFIG_HOME/homebrew/brew.env` if `$XDG_CONFIG_HOME` is set or
`$HOME/.homebrew/brew.env` otherwise (user-specific)

User-specific environment files take precedence over prefix-specific files and
prefix-specific files take precedence over system-wide files (unless
`HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY` is set, see below).

Note that these files do not support shell variable expansion e.g. `$HOME` or
command execution e.g. `$(cat file)`.

- `HOMEBREW_API_DOMAIN`: Use this URL as the download mirror for Homebrew JSON API. If metadata files at that URL are temporarily unavailable, the default API domain will be used as a fallback mirror.

    **Default:** `https://formulae.brew.sh/api`.

- `HOMEBREW_ARCH`: Linux only: Pass this value to a type name representing the compiler’s `-march` option.

    **Default:** `native`.

- `HOMEBREW_ARTIFACT_DOMAIN`: Prefix all download URLs, including those for bottles, with this value. For example, `HOMEBREW_ARTIFACT_DOMAIN=http://localhost:8080` will cause a formula with the URL `https://example.com/foo.tar.gz` to instead download from `http://localhost:8080/https://example.com/foo.tar.gz`. Bottle URLs however, have their domain replaced with this prefix. This results in e.g. `https://ghcr.io/v2/homebrew/core/gettext/manifests/0.21` to instead be downloaded from `http://localhost:8080/v2/homebrew/core/gettext/manifests/0.21`

- `HOMEBREW_API_AUTO_UPDATE_SECS`: Check Homebrew’s API for new formulae or cask data every `HOMEBREW_API_AUTO_UPDATE_SECS` seconds. Alternatively, disable API auto-update checks entirely with `HOMEBREW_NO_AUTO_UPDATE`.

    **Default:** `450`.

- `HOMEBREW_AUTO_UPDATE_SECS`: Run `brew update` once every `HOMEBREW_AUTO_UPDATE_SECS` seconds before some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, disable auto-update entirely with `HOMEBREW_NO_AUTO_UPDATE`.

    **Default:** `86400` (24 hours), `3600` (1 hour) if a developer command has been run or `300` (5 minutes) if `HOMEBREW_NO_INSTALL_FROM_API` is set.

- `HOMEBREW_AUTOREMOVE`: If set, calls to `brew cleanup` and `brew uninstall` will automatically remove unused formula dependents and if `HOMEBREW_NO_INSTALL_CLEANUP` is not set, `brew cleanup` will start running `brew autoremove` periodically.

- `HOMEBREW_BAT`: If set, use `bat` for the `brew cat` command.

- `HOMEBREW_BAT_CONFIG_PATH`: Use this as the `bat` configuration file.

    **Default:** `$BAT_CONFIG_PATH`.

- `HOMEBREW_BAT_THEME`: Use this as the `bat` theme for syntax highlighting.

    **Default:** `$BAT_THEME`.

- `HOMEBREW_BOOTSNAP`: If set, use Bootsnap to speed up repeated `brew` calls. A no-op when using Homebrew’s vendored, relocatable Ruby on macOS (as it doesn’t work).

- `HOMEBREW_BOTTLE_DOMAIN`: Use this URL as the download mirror for bottles. If bottles at that URL are temporarily unavailable, the default bottle domain will be used as a fallback mirror. For example, `HOMEBREW_BOTTLE_DOMAIN=http://localhost:8080` will cause all bottles to download from the prefix `http://localhost:8080/`. If bottles are not available at `HOMEBREW_BOTTLE_DOMAIN` they will be downloaded from the default bottle domain.

    **Default:** `https://ghcr.io/v2/homebrew/core`.

- `HOMEBREW_BREW_GIT_REMOTE`: Use this URL as the Homebrew/brew `git`(1) remote.

    **Default:** `https://github.com/Homebrew/brew`.

- `HOMEBREW_BROWSER`: Use this as the browser when opening project homepages.

    **Default:** `$BROWSER` or the OS’s default browser.

- `HOMEBREW_CACHE`: Use this directory as the download cache.

    **Default:** macOS: `$HOME/Library/Caches/Homebrew`, Linux: `$XDG_CACHE_HOME/Homebrew` or `$HOME/.cache/Homebrew`.

- `HOMEBREW_CASK_OPTS`: Append these options to all `cask` commands. All `--*dir` options, `--language`, `--require-sha`, `--no-quarantine` and `--no-binaries` are supported. For example, you might add something like the following to your `~/.profile`, `~/.bash_profile`, or `~/.zshenv`:

    `export HOMEBREW_CASK_OPTS="--appdir=~/Applications --fontdir=/Library/Fonts"`

- `HOMEBREW_CLEANUP_MAX_AGE_DAYS`: Cleanup all cached files older than this many days.

    **Default:** `120`.

- `HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS`: If set, `brew install`, `brew upgrade` and `brew reinstall` will cleanup all formulae when this number of days has passed.

    **Default:** `30`.

- `HOMEBREW_COLOR`: If set, force colour output on non-TTY outputs.

- `HOMEBREW_CORE_GIT_REMOTE`: Use this URL as the Homebrew/homebrew-core `git`(1) remote.

    **Default:** `https://github.com/Homebrew/homebrew-core`.

- `HOMEBREW_CURL_PATH`: Linux only: Set this value to a new enough `curl` executable for Homebrew to use.

    **Default:** `curl`.

- `HOMEBREW_CURL_RETRIES`: Pass the given retry count to `--retry` when invoking `curl`(1).

    **Default:** `3`.

- `HOMEBREW_CURL_VERBOSE`: If set, pass `--verbose` when invoking `curl`(1).

- `HOMEBREW_CURLRC`: If set to an absolute path (i.e. beginning with `/`), pass it with `--config` when invoking `curl`(1). If set but **not** a valid path, do not pass `--disable`, which disables the use of `.curlrc`.

- `HOMEBREW_DEBUG`: If set, always assume `--debug` when running commands.

- `HOMEBREW_DEVELOPER`: If set, tweak behaviour to be more relevant for Homebrew developers (active or budding) by e.g. turning warnings into errors.

- `HOMEBREW_DISABLE_LOAD_FORMULA`: If set, refuse to load formulae. This is useful when formulae are not trusted (such as in pull requests).

- `HOMEBREW_DISPLAY`: Use this X11 display when opening a page in a browser, for example with `brew home`. Primarily useful on Linux.

    **Default:** `$DISPLAY`.

- `HOMEBREW_DISPLAY_INSTALL_TIMES`: If set, print install times for each formula at the end of the run.

- `HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN`: Use this base64 encoded username and password for authenticating with a Docker registry proxying GitHub Packages. If `HOMEBREW_DOCKER_REGISTRY_TOKEN` is set, it will be used instead.

- `HOMEBREW_DOCKER_REGISTRY_TOKEN`: Use this bearer token for authenticating with a Docker registry proxying GitHub Packages. Preferred over `HOMEBREW_DOCKER_REGISTRY_BASIC_AUTH_TOKEN`.

- `HOMEBREW_EDITOR`: Use this editor when editing a single formula, or several formulae in the same directory.

    **Note:** `brew edit` will open all of Homebrew as discontinuous files and directories. Visual Studio Code can handle this correctly in project mode, but many editors will do strange things in this case.

    **Default:** `$EDITOR` or `$VISUAL`.

- `HOMEBREW_EVAL_ALL`: If set, `brew` commands evaluate all formulae and casks, executing their arbitrary code, by default without requiring `--eval-all`. Required to cache formula and cask descriptions.

- `HOMEBREW_FAIL_LOG_LINES`: Output this many lines of output on formula `system` failures.

    **Default:** `15`.

- `HOMEBREW_FORBIDDEN_LICENSES`: A space-separated list of licenses. Homebrew will refuse to install a formula if it or any of its dependencies has a license on this list.

- `HOMEBREW_FORCE_BREWED_CA_CERTIFICATES`: If set, always use a Homebrew-installed `ca-certificates` rather than the system version. Automatically set if the system version is too old.

- `HOMEBREW_FORCE_BREWED_CURL`: If set, always use a Homebrew-installed `curl`(1) rather than the system version. Automatically set if the system version of `curl` is too old.

- `HOMEBREW_FORCE_BREWED_GIT`: If set, always use a Homebrew-installed `git`(1) rather than the system version. Automatically set if the system version of `git` is too old.

- `HOMEBREW_FORCE_VENDOR_RUBY`: If set, always use Homebrew’s vendored, relocatable Ruby version even if the system version of Ruby is new enough.

- `HOMEBREW_GIT_EMAIL`: Set the Git author and committer email to this value.

- `HOMEBREW_GIT_NAME`: Set the Git author and committer name to this value.

- `HOMEBREW_GIT_PATH`: Linux only: Set this value to a new enough `git` executable for Homebrew to use.

    **Default:** `git`.

- `HOMEBREW_GITHUB_API_TOKEN`: Use this personal access token for the GitHub API, for features such as `brew search`. You can create one at <https://github.com/settings/tokens>. If set, GitHub will allow you a greater number of API requests. For more information, see: <https://docs.github.com/en/rest/overview/rate-limits-for-the-rest-api>

    **Note:** Homebrew doesn’t require permissions for any of the scopes, but some developer commands may require additional permissions.

- `HOMEBREW_GITHUB_PACKAGES_TOKEN`: Use this GitHub personal access token when accessing the GitHub Packages Registry (where bottles may be stored).

- `HOMEBREW_GITHUB_PACKAGES_USER`: Use this username when accessing the GitHub Packages Registry (where bottles may be stored).

- `HOMEBREW_INSTALL_BADGE`: Print this text before the installation summary of each successful build.

    **Default:** The “Beer Mug” emoji.

- `HOMEBREW_LIVECHECK_WATCHLIST`: Consult this file for the list of formulae to check by default when no formula argument is passed to `brew livecheck`.

    **Default:** `$XDG_CONFIG_HOME/homebrew/livecheck_watchlist.txt` if `$XDG_CONFIG_HOME` is set or `$HOME/.homebrew/livecheck_watchlist.txt` otherwise.

- `HOMEBREW_LOGS`: Use this directory to store log files.

    **Default:** macOS: `$HOME/Library/Logs/Homebrew`, Linux: `$XDG_CACHE_HOME/Homebrew/Logs` or `$HOME/.cache/Homebrew/Logs`.

- `HOMEBREW_MAKE_JOBS`: Use this value as the number of parallel jobs to run when building with `make`(1).

    **Default:** The number of available CPU cores.

- `HOMEBREW_NO_ANALYTICS`: If set, do not send analytics. Google Analytics were destroyed. For more information, see: <https://docs.brew.sh/Analytics>

- `HOMEBREW_NO_AUTO_UPDATE`: If set, do not automatically update before running some commands, e.g. `brew install`, `brew upgrade` and `brew tap`. Alternatively, run this less often by setting `HOMEBREW_AUTO_UPDATE_SECS` to a value higher than the default.

- `HOMEBREW_NO_BOOTSNAP`: If set, do not use Bootsnap to speed up repeated `brew` calls.

- `HOMEBREW_NO_CLEANUP_FORMULAE`: A comma-separated list of formulae. Homebrew will refuse to clean up or autoremove a formula if it appears on this list.

- `HOMEBREW_NO_COLOR`: If set, do not print text with colour added.

    **Default:** `$NO_COLOR`.

- `HOMEBREW_NO_EMOJI`: If set, do not print `HOMEBREW_INSTALL_BADGE` on a successful build.

- `HOMEBREW_NO_ENV_HINTS`: If set, do not print any hints about changing Homebrew’s behaviour with environment variables.

- `HOMEBREW_NO_GITHUB_API`: If set, do not use the GitHub API, e.g. for searches or fetching relevant issues after a failed install.

- `HOMEBREW_NO_INSECURE_REDIRECT`: If set, forbid redirects from secure HTTPS to insecure HTTP.

    **Note:** while ensuring your downloads are fully secure, this is likely to cause from-source SourceForge, some GNU &amp; GNOME-hosted formulae to fail to download.

- `HOMEBREW_NO_INSTALL_CLEANUP`: If set, `brew install`, `brew upgrade` and `brew reinstall` will never automatically cleanup installed/upgraded/reinstalled formulae or all formulae every `HOMEBREW_CLEANUP_PERIODIC_FULL_DAYS` days. Alternatively, `HOMEBREW_NO_CLEANUP_FORMULAE` allows specifying specific formulae to not clean up.

- `HOMEBREW_NO_INSTALL_FROM_API`: If set, do not install formulae and casks in homebrew/core and homebrew/cask taps using Homebrew’s API and instead use (large, slow) local checkouts of these repositories.

- `HOMEBREW_NO_INSTALL_UPGRADE`: If set, `brew install` **`formula|cask`** will not upgrade **`formula|cask`** if it is installed but outdated.

- `HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK`: If set, do not check for broken linkage of dependents or outdated dependents after installing, upgrading or reinstalling formulae. This will result in fewer dependents (and their dependencies) being upgraded or reinstalled but may result in more breakage from running `brew install` **`formula`** or `brew upgrade` **`formula`**.

- `HOMEBREW_NO_UPDATE_REPORT_NEW`: If set, `brew update` will not show the list of newly added formulae/casks.

- `HOMEBREW_PIP_INDEX_URL`: If set, `brew install` **`formula`** will use this URL to download PyPI package resources.

    **Default:** `https://pypi.org/simple`.

- `HOMEBREW_PRY`: If set, use Pry for the `brew irb` command.

- `HOMEBREW_UPGRADE_GREEDY`: If set, pass `--greedy` to all cask upgrade commands.

- `HOMEBREW_SIMULATE_MACOS_ON_LINUX`: If set, running Homebrew on Linux will simulate certain macOS code paths. This is useful when auditing macOS formulae while on Linux.

- `HOMEBREW_SKIP_OR_LATER_BOTTLES`: If set along with `HOMEBREW_DEVELOPER`, do not use bottles from older versions of macOS. This is useful in development on new macOS versions.

- `HOMEBREW_SORBET_RUNTIME`: If set, enable runtime typechecking using Sorbet. Set by default for `HOMEBREW_DEVELOPER` or when running some developer commands.

- `HOMEBREW_SSH_CONFIG_PATH`: If set, Homebrew will use the given config file instead of `~/.ssh/config` when fetching Git repositories over SSH.

    **Default:** `$HOME/.ssh/config`

- `HOMEBREW_SVN`: Use this as the `svn`(1) binary.

    **Default:** A Homebrew-built Subversion (if installed), or the system-provided binary.

- `HOMEBREW_SYSTEM_ENV_TAKES_PRIORITY`: If set in Homebrew’s system-wide environment file (`/etc/homebrew/brew.env`), the system-wide environment file will be loaded last to override any prefix or user settings.

- `HOMEBREW_SUDO_THROUGH_SUDO_USER`: If set, Homebrew will use the `SUDO_USER` environment variable to define the user to `sudo`(8) through when running `sudo`(8).

- `HOMEBREW_TEMP`: Use this path as the temporary directory for building packages. Changing this may be needed if your system temporary directory and Homebrew prefix are on different volumes, as macOS has trouble moving symlinks across volumes when the target does not yet exist. This issue typically occurs when using FileVault or custom SSD configurations.

    **Default:** macOS: `/private/tmp`, Linux: `/tmp`.

- `HOMEBREW_UPDATE_TO_TAG`: If set, always use the latest stable tag (even if developer commands have been run).

- `HOMEBREW_VERBOSE`: If set, always assume `--verbose` when running commands.

- `HOMEBREW_VERBOSE_USING_DOTS`: If set, verbose output will print a `.` no more than once a minute. This can be useful to avoid long-running Homebrew commands being killed due to no output.

- `SUDO_ASKPASS`: If set, pass the `-A` option when calling `sudo`(8).

- `all_proxy`: Use this SOCKS5 proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.

- `ftp_proxy`: Use this FTP proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.

- `http_proxy`: Use this HTTP proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.

- `https_proxy`: Use this HTTPS proxy for `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.

- `no_proxy`: A comma-separated list of hostnames and domain names excluded from proxying by `curl`(1), `git`(1) and `svn`(1) when downloading through Homebrew.
