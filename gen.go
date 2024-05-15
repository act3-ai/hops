package hops

// Generates CLI documentation in markdown format
// The generated CLI docs are used as source code for the website
//go:generate /usr/bin/env bash -c "NO_COLOR=1 go run ./cmd/hops gendocs md --only-commands docs/cli/"
