# Changelog

All notable changes to this project will be documented in this file.

## [0.2.1] (2024-06-05)

### Bug Fixes

- **search**: rule was disabling search command ([af693a5](https://github.com/act3-ai/hops/commit/af693a5684d80d4a4cc8233d1a13528daca213f5))

## [0.2.0] (2024-05-29)

### Features

- **install**: respect pour_bottle_only_if conditions ([f906952](https://github.com/act3-ai/hops/commit/f90695247bfded73688f737fbf0210877a1f808d))
- **HOMEBREW_BOTTLE_DOMAIN**: respect the HOMEBREW_BOTTLE_DOMAIN option ([85d625f](https://github.com/act3-ai/hops/commit/85d625fd99eade6744f7b61b8efb8dd64265e5ba))
- **configuration**: unify config file with Homebrew environment config ([ca288cb](https://github.com/act3-ai/hops/commit/ca288cbe246fd53ffff9444ff88e0d8a3c0951eb))
- **configuration**: support custom HTTP headers for registry requests ([132e73e](https://github.com/act3-ai/hops/commit/132e73ef4b8c3646ee99c394015f295ecb8ebd1f))
- **flags**: standardize registry flags ([76133e9](https://github.com/act3-ai/hops/commit/76133e9ba501b2867170f9d6a829c34f61e28110))
- **logging**: support JSON and logfmt logging with --log-fmt flag ([3c3718d](https://github.com/act3-ai/hops/commit/3c3718d14dc05471c034ca08d983c86b4f3ae084))

### Bug Fixes

- **logging**: update verbosity flag handling ([90c6bf4](https://github.com/act3-ai/hops/commit/90c6bf4b258ee8f7c6b5ba96379c8c797fccdb42))
- change formula versioner function signatures ([0e07d93](https://github.com/act3-ai/hops/commit/0e07d935ffa496e397a7d0342e20090593cd8da4))
- **configuration**: expose more environment variables for registry settings ([7b885f3](https://github.com/act3-ai/hops/commit/7b885f33a2db6c0d41cd8498c31d0250d7534bc3))
- **cross-platform bottles**: add workaround to support cross-platform bottles that are not published correctly ([028ab39](https://github.com/act3-ai/hops/commit/028ab392517bc9114b7eeceae87a5de9039df378))
- **changelog**: add commit links ([fc5af57](https://github.com/act3-ai/hops/commit/fc5af5759e437a44766cce46572acef57fc7d76b))

### Dependencies

- **task**: update golangci-lint to 1.58.2 ([c73a8c0](https://github.com/act3-ai/hops/commit/c73a8c0d4acf862bc503f30430ca5a12f8b677f9))
- **github-actions**: update markdownlint ([7d91e7b](https://github.com/act3-ai/hops/commit/7d91e7b77155a83cf85038dbe43e34e82eadf636))

### Miscellaneous Tasks

- get Go version from go.mod ([07ef392](https://github.com/act3-ai/hops/commit/07ef3924dd3ccabac8f4af1522e31b7315826bbb))
- add simple integration tests ([ab83c4e](https://github.com/act3-ai/hops/commit/ab83c4eedfc9d1d4f96b995223f39c7ba1587663))
- **deps**: Bump github.com/charmbracelet/lipgloss ([1866f0b](https://github.com/act3-ai/hops/commit/1866f0ba96d5b3df4f6c048ac0fd564fd2ab254a))

## [0.1.3] (2024-05-24)

### Bug Fixes

- **tap**: use deploy key ([27f0974](https://github.com/act3-ai/hops/commit/27f0974d9adcce91466ad28f6111cf222b482b94))

## [0.1.2] (2024-05-24)

### Bug Fixes

- **dist**: publish homebrew formula ([91f015c](https://github.com/act3-ai/hops/commit/91f015ceb09491b65070763c339f10c3a5585e9a))

### Miscellaneous Tasks

- **ko**: preserve import paths ([9169fd2](https://github.com/act3-ai/hops/commit/9169fd2a4d6b6539aca0694a65273971b15cb9d9))

## [0.1.1] (2024-05-24)

### Miscellaneous Tasks

- **image**: fix CI image tag ([16e7864](https://github.com/act3-ai/hops/commit/16e7864b48244e38debd0614ea6db25e7c61fb31))
- **build image**: give publish permissions ([14cdf96](https://github.com/act3-ai/hops/commit/14cdf964352d685d6caf71e2b7d71062d7bce3b1))

## [0.1.0] (2024-05-24)

### Features

- **actions**: add build workflow (#1) ([41a9a2d](https://github.com/act3-ai/hops/commit/41a9a2d99f5b066e68d9140b53728f7fbf65d0a6))
- **actions**: add release workflow (#3) ([6ea5d26](https://github.com/act3-ai/hops/commit/6ea5d269b12108c85918258394f9e5afed90e926))
- share commands between default and registry mode (#7) ([2e11547](https://github.com/act3-ai/hops/commit/2e115472a3940fa5afe1c7449783c9ff9fb3d482))
- **changelog**: add git-cliff configuration for changelog generation ([1adb4db](https://github.com/act3-ai/hops/commit/1adb4db360495c0bcc4efc2b01a9264eb4061619))

### Bug Fixes

- **fips build workflow**: check out before running local workflow ([3a08b1a](https://github.com/act3-ai/hops/commit/3a08b1a86a0dfd6a151fe49bb3a6e4d55614a846))
- **release config**: separate default and fips builds ([a6a94b2](https://github.com/act3-ai/hops/commit/a6a94b2c66673f4e52c9e50366496780b3423b75))
- **README**: add install instructions ([cdf9712](https://github.com/act3-ai/hops/commit/cdf9712c9d60e3f7661f5b4a33c86b62c8f29de3))
- update goreleaser cfg ([06084a0](https://github.com/act3-ai/hops/commit/06084a0d3b9066768713a3957b80c43a65f8217a))
- push arm docker image with goreleaser ([04e0e75](https://github.com/act3-ai/hops/commit/04e0e75d5073c071a9e56de2e5c352500933570f))
- use default ldflags ([dbcc15d](https://github.com/act3-ai/hops/commit/dbcc15d0cf50c80ffc421e3b6f5a1b4d94f0c3c4))
- **dependabot**: force convential commits for dependabot PRs ([551c956](https://github.com/act3-ai/hops/commit/551c956ef775244422c9ed14f3aa05d52832fa9d))
- **dependabot**: assign reviewer to dependabot PRs ([b77f5f0](https://github.com/act3-ai/hops/commit/b77f5f041a74afa261f604b1f10dabb84cafe4eb))
- allow empty commit for changelog ([48d8325](https://github.com/act3-ai/hops/commit/48d83258acc020644146c945fb39090cafa89001))
- remove docker hub image reference ([ae76c82](https://github.com/act3-ai/hops/commit/ae76c829b7995c36a6e6fe75495c1101645171ff))
- simplify goreleaser config to publish one manifest tag ([351e092](https://github.com/act3-ai/hops/commit/351e092316e936604da16284836671faa852ee22))
- set up docker and qemu in release workflow ([31a84f3](https://github.com/act3-ai/hops/commit/31a84f3a8620490b7926c1949a9b50ea482c1981))
- simplify Goreleaser image build while debugging ([0811013](https://github.com/act3-ai/hops/commit/081101382ec02480add6a0f2755951a6cc7a75f2))
- **release workflow**: change name ([85ea1a2](https://github.com/act3-ai/hops/commit/85ea1a2479f7ff9f1ec5c05732f2f5b3cd732e54))
- **release workflow**: add login step for ghcr.io before running Goreleaser ([ee72a8c](https://github.com/act3-ai/hops/commit/ee72a8c84e19b0d44c00da219816e237d8b0b0df))
- **lint**: enable more linters ([c5c1978](https://github.com/act3-ai/hops/commit/c5c1978174070caed7111a42426fea759ca7c0b6))
- address lint issues ([c26c31c](https://github.com/act3-ai/hops/commit/c26c31ce8a724385594eb824e6e3f6145f077a0e))
- update commit grouping in release notes ([8786598](https://github.com/act3-ai/hops/commit/87865988833c8968f2a6a8239c46b04e0c8d9b11))
- update image labels ([8940a3d](https://github.com/act3-ai/hops/commit/8940a3d122a2cd9d175ac17c0caf77f3b404a788))
- **goreleaser**: remove md5 hash from ko image paths ([01a6faf](https://github.com/act3-ai/hops/commit/01a6fafefa076a5a5ad671c120140393bf572804))
- **goreleaser**: update goreleaser config ([49cea86](https://github.com/act3-ai/hops/commit/49cea86ce22c06acfa6c8032292b1bdd6dce5c56))
- **release**: update release notes generation ([9c19ea1](https://github.com/act3-ai/hops/commit/9c19ea19a47c640bbc65d8aed8fb111a517865a6))
- release notes again ([5adfb1f](https://github.com/act3-ai/hops/commit/5adfb1f1b5cf0b4e778d73bcb30d0e0c70ad40e1))
- update goreleaser config ([48d31af](https://github.com/act3-ai/hops/commit/48d31af9073d9efa0dd5afb1daa353c5b77e7ea7))
- **versioning**: do not trim leading "v" in version tags ([bb45e5a](https://github.com/act3-ai/hops/commit/bb45e5a0142636c8d76388074737382ecedb30e7))
- update cliff config ([8fdd124](https://github.com/act3-ai/hops/commit/8fdd124efaf91aa40f3d121f0fb9c5b1ae8c96fd))

### Documentation

- **CONTRIBUTING.md**: document release process ([7281842](https://github.com/act3-ai/hops/commit/72818423ec8f1bc39c527046fcacffb53b3111ef))

### Miscellaneous Tasks

- run govulncheck ([6d0b478](https://github.com/act3-ai/hops/commit/6d0b4786109686fe4251f28cae5f01889eaec544))
- **changelog**: update changelog for tag v0.1.0-beta.0 ([a718c25](https://github.com/act3-ai/hops/commit/a718c25d65704a8de8e1ff050e90e1ace9cb742f))
- **changelog**: update changelog for tag v0.1.0-beta.1 ([8fa36c9](https://github.com/act3-ai/hops/commit/8fa36c9bb4fab42985d72d8d9a9dafc4e0583252))
- **changelog**: update changelog for tag v0.1.0-beta.2 ([3dbe278](https://github.com/act3-ai/hops/commit/3dbe278e46a215e1beb1a44063181cc4dbbf23b9))
- **repo**: remove unused gitattributes file ([885d06d](https://github.com/act3-ai/hops/commit/885d06dce46d762a94250d00ff48a926dfb9ffe6))
- update version name to avoid collisions ([2ec12f5](https://github.com/act3-ai/hops/commit/2ec12f57b05795b1d6aae31007eab08578f1d0d5))
- update changelog configuration ([362ea65](https://github.com/act3-ai/hops/commit/362ea65dad1787d02ed7033ce12d954af523253b))
- update dependabot commit messages ([55edad6](https://github.com/act3-ai/hops/commit/55edad6ed2f32a9fc94f5231b064cfdbc5bff5d8))
- **Release**: add fetch-depth 0 to checkout ([5f48b4b](https://github.com/act3-ai/hops/commit/5f48b4b5f8fb112eac68a5ac3376dcae20c0ec17))
- **release**: fix flag ([6b71969](https://github.com/act3-ai/hops/commit/6b7196947ace70c8872e678c7070e5f615adf167))
- **docs**: generate docs ([bf3fe7d](https://github.com/act3-ai/hops/commit/bf3fe7de1c3cf7f955d9e94386964ea6d966eca2))

[0.2.1]: https://github.com/act3-ai/hops/compare/v0.2.0..v0.2.1
[0.2.0]: https://github.com/act3-ai/hops/compare/v0.1.3..v0.2.0
[0.1.3]: https://github.com/act3-ai/hops/compare/v0.1.2..v0.1.3
[0.1.2]: https://github.com/act3-ai/hops/compare/v0.1.1..v0.1.2
[0.1.1]: https://github.com/act3-ai/hops/compare/v0.1.0..v0.1.1
[0.1.0]: https://github.com/act3-ai/hops/tree/v0.1.0

