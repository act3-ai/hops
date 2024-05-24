# Changelog

All notable changes to this project will be documented in this file.

## [0.1.3] (2024-05-24)

### Bug Fixes

- **tap**: Use deploy key

## [0.1.2] (2024-05-24)

### Bug Fixes

- **dist**: Publish homebrew formula

### Miscellaneous Tasks

- **ko**: Preserve import paths

## [0.1.1] (2024-05-24)

### Miscellaneous Tasks

- **image**: Fix CI image tag
- **build image**: Give publish permissions

## [0.1.0] (2024-05-24)

### Features

- **actions**: Add build workflow (#1)
- **actions**: Add release workflow (#3)
- Share commands between default and registry mode (#7)
- **changelog**: Add git-cliff configuration for changelog generation

### Bug Fixes

- **fips build workflow**: Check out before running local workflow
- **release config**: Separate default and fips builds
- **README**: Add install instructions
- Update goreleaser cfg
- Push arm docker image with goreleaser
- Use default ldflags
- **dependabot**: Force convential commits for dependabot PRs
- **dependabot**: Assign reviewer to dependabot PRs
- Allow empty commit for changelog
- Remove docker hub image reference
- Simplify goreleaser config to publish one manifest tag
- Set up docker and qemu in release workflow
- Simplify Goreleaser image build while debugging
- **release workflow**: Change name
- **release workflow**: Add login step for ghcr.io before running Goreleaser
- **lint**: Enable more linters
- Address lint issues
- Update commit grouping in release notes
- Update image labels
- **goreleaser**: Remove md5 hash from ko image paths
- **goreleaser**: Update goreleaser config
- **release**: Update release notes generation
- Release notes again
- Update goreleaser config
- **versioning**: Do not trim leading "v" in version tags
- Update cliff config

### Documentation

- **CONTRIBUTING.md**: Document release process

### Miscellaneous Tasks

- Run govulncheck
- **changelog**: Update changelog for tag v0.1.0-beta.0
- **changelog**: Update changelog for tag v0.1.0-beta.1
- **changelog**: Update changelog for tag v0.1.0-beta.2
- **repo**: Remove unused gitattributes file
- Update version name to avoid collisions
- Update changelog configuration
- Update dependabot commit messages
- **Release**: Add fetch-depth 0 to checkout
- **release**: Fix flag
- **docs**: Generate docs

[0.1.3]: https://github.com/act3-ai/hops/compare/v0.1.2..v0.1.3
[0.1.2]: https://github.com/act3-ai/hops/compare/v0.1.1..v0.1.2
[0.1.1]: https://github.com/act3-ai/hops/compare/v0.1.0..v0.1.1
[0.1.0]: https://github.com/act3-ai/hops/tree/v0.1.0

