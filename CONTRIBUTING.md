# Hops Contributing Guide

## Design Patterns

The implementation is organized in the following layers:

- [`main` Package](./cmd/hops): Entrypoint of the program
- [`cli` Package](./internal/cli): CLI commands defined using the [`cobra`](https://pkg.go.dev/github.com/spf13/cobra) framework
- [`actions` Package](./internal/actions): Functionality of the CLI commands
  - Each command defined in the `cli` package runs an "action" in the `actions` package
- Other Packages: Smaller components of functionality
  - [Internal Packages](./internal): Packages defined for use in this repository
  <!-- - [Public Packages](./pkg): Packages defined for use in this repository and others -->

## Tooling

The following tools are used for local development of Hops:

- [Go](https://go.dev/): build, test, format
- [Taskfile](https://taskfile.dev/): running tasks
- [Podman](https://podman.io/): building images and running local container registries
- [git-cliff](https://git-cliff.org/): version calculation and changelog generation

## Documentation

### CLI Documentation Generation

A `go:generate` directive in [`gen.go`](./gen.go) runs the following CLI command to regenerate CLI docs:

```bash
NO_COLOR=1 hops gendocs md --only-commands docs/cli/
```

## Testing

### Local Registry

To start a local registry for testing:

```sh
podman run --name testreg --rm -d -p 5001:5000 registry:2
```

Now a local registry is available at the following address:

```plain
localhost:5001
```

The `--plain-http` flag is required to use this local registry.

Example:

```sh
# Run a local registry for testing
podman run --name testreg --rm -d -p 5001:5000 registry:2

# Copy the glab Bottle to the local registry
hops copy glab --to localhost:5001/bottles --to-plain-http

# Install glab from the local registry using flags
hops xinstall glab --registry localhost:5001/bottles --plain-http

# Install glab from the local registry using environment variables
export HOPS_REGISTRY=localhost:5001/bottles
export HOPS_REGISTRY_PLAIN_HTTP=true
hops xinstall glab

# Stop running the local registry
podman stop testreg
```

### Unit Tests

Run the following command from the root directory of the repository. This will run all unit tests

```bash
go test ./...

# or

task test
```

## Releasing

Releases are created with Git tags and distributed with GoReleaser.

Once you have run all tests and code generation.

Release process, first calculate the version of the next release:

```sh
# Calculate version of the next release by checking Git log
git cliff --bumped-version

# Save incremented version to environment variable
# Can use any version, git-cliff version is just a convenience
VERSION="$(git cliff --bumped-version)"

# Update VERSION file
echo "$VERSION" >VERSION

# Update CHANGELOG.md
git cliff --tag "$VERSION" --output CHANGELOG.md

# View release notes for the next version
# Taskfile shortcut: "task next-changelog"
git cliff --tag "$VERSION" --unreleased --strip header

# Save next version's release notes to environment variable
# This is so they can be used in the release commit message
RELEASE_NOTES="$(git cliff --tag "$VERSION" --unreleased --strip header)"

# Stage VERSION and CHANGELOG.md
git add VERSION CHANGELOG.md

# Commit VERSION and CHANGELOG.md
git commit -m "chore(release): $VERSION" -m "$RELEASE_NOTES"

# Create release tag
git tag "$VERSION"
```

Once the release tag has been created, GoReleaser can be run either locally or in a GitHub Action.

Run GoReleaser locally:

```sh
GITHUB_TOKEN="<token>" goreleaser release --clean
```

Run in GitHub Actions:

```sh
# Push any local commits
git push

# Push the new release tag
git push origin "<release tag>"
```

Once the tag has been pushed, the ["Release" workflow](./.github/workflows/release.yml) will run. This workflow runs GoReleaser on the tag.
