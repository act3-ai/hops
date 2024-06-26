---
title: hops install
description: Install a formula
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# hops install

Install a formula

## Synopsis

Install a formula. Additional options specific to a formula may be appended to the command.

Unless HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK is set, brew upgrade or brew
reinstall will be run for outdated dependents and dependents with broken
linkage, respectively.

Unless HOMEBREW_NO_INSTALL_CLEANUP is set, brew cleanup will then be run for
the installed formulae or, every 30 days, for all formulae.

Unless HOMEBREW_NO_INSTALL_UPGRADE is set, brew install formula will
upgrade formula if it is already installed but outdated.

STANDALONE MODE:

Hops has an alternate mode to fetch all packages and metadata from a single OCI registry.
The default behavior for standalone mode is to install the version tagged "latest".
The tag for a formula can be set by using the argument format "<formula>:<tag>".


## Usage

```plaintext
hops install formula [...] [flags]
```

## Options

```plaintext
      --dry-run                  Show what would be installed, but do not actually install anything
      --force                    Install formulae without checking for previously installed keg-only or non-migrated versions. When installing casks, overwrite existing files (binaries and symlinks are excluded, unless originally from the same cask)
      --header stringArray       Add custom headers to requests
  -h, --help                     help for install
      --ignore-dependencies      Skip installing any dependencies of any kind [TESTING-ONLY]
      --include-build            Include :build dependencies for formula
      --include-optional         Include :optional dependencies for formula
      --include-test             Include :test dependencies for formula (non-recursive)
      --oci-layout               Set target as an OCI image layout
      --only-dependencies        Install the dependencies with specified options but do not install the formula itself
      --overwrite                Delete files that already exist in the prefix while linking
      --plain-http               Allow insecure connections to registry without SSL check
      --registry string          Registry prefix for bottles (overrides config)
      --registry-config string   Path of the authentication file for registry
      --skip-recommended         Skip :recommended dependencies for formula
```

## Options inherited from parent commands

```plaintext
      --concurrency int   Concurrency level (default 8)
      --config strings    Set config file search paths (default `hops-config.yaml`,`$XDG_CONFIG_HOME/hops/config.yaml`,`/etc/hops/config.yaml`)
  -d, --debug count       Display more debugging information
      --log-fmt string    Set format for log messages. Options: text, json (default "text")
  -q, --quiet count       Make some output more quiet
  -v, --verbose count     Make some output more verbose
```
