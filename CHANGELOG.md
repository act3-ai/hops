# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0-beta.1] (2024-05-23)

### Bug Fixes

- Remove docker hub image reference

## [0.1.0-beta.0] (2024-05-23)

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
- **ci**: Run govulncheck
- **dependabot**: Force convential commits for dependabot PRs
- **dependabot**: Assign reviewer to dependabot PRs
- Allow empty commit for changelog

### Miscellaneous Tasks

- **changelog**: Update changelog for tag v0.1.0-beta.0

[0.1.0-beta.1]: https://github.com/act3-ai/hops/compare/v0.1.0-beta.0..v0.1.0-beta.1
[0.1.0-beta.0]: https://github.com/act3-ai/hops/tree/v0.1.0-beta.0

