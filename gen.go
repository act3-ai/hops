package hops

// Generates CLI documentation in markdown format.
// The generated CLI docs are used as source code for the website.
//go:generate /usr/bin/env bash -c "NO_COLOR=1 go run ./cmd/hops gendocs md --only-commands docs/cli/"

// Generates performant JSON Marshal/Unmarshal functions for structs that avoid reflection.
//go:generate go run github.com/mailru/easyjson/easyjson@master -all -pkg -- ./internal/apis/formulae.brew.sh/v1
//go:generate go run github.com/mailru/easyjson/easyjson@master -all -pkg -- ./internal/apis/formulae.brew.sh/v2
//go:generate go run github.com/mailru/easyjson/easyjson@master -all -pkg -- ./internal/apis/formulae.brew.sh/v3
//go:generate go run github.com/mailru/easyjson/easyjson@master -all -pkg -- ./internal/apis/formulae.brew.sh/common
