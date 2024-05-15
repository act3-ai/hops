# hops Developer Guide

## Design Patterns

The CLI's implementation is organized as the following layers:

- [`main` Package](./cmd/hops): Entrypoint of the program
- [`cli` Package](./internal/cli): CLI commands defined using the [`cobra`](https://pkg.go.dev/github.com/spf13/cobra) framework
- [`actions` Package](./internal/actions): Functionality of the CLI commands
  - Each command defined in the `cli` package runs an "action" in the `actions` package
- Other Packages: Smaller components of functionality
  - [Internal Packages](./internal): Packages defined for use in this repository
  - [Public Packages](./pkg): Packages defined for use in this repository and others

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

<!-- 
HOPS_REGISTRY=localhost:5001/bottles
HOPS_REGISTRY_PLAIN_HTTP=true
 -->
```sh
# Run a local registry for testing
podman run --name testreg --rm -d -p 5001:5000 registry:2

# Copy the glab Bottle to the local registry
hops copy glab --to localhost:5001/bottles --to-plain-http

# Install glab from the local registry
hops xinstall glab --registry localhost:5001/bottles --plain-http

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

The act3-pt CLI contains a `act3-pt ci release` command that automates this process.

## Code Generation

### Generate CLI Documentation (automatically done in CI/CD pipeline)

A `go:generate` directive in [`gen.go`](./gen.go) runs the following CLI command to regenerate CLI docs:

```bash
NO_COLOR=1 hops gendocs md --only-commands docs/cli/
```
