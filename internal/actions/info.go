package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/pretty"
)

// Info represents the action and its options.
type Info struct {
	*Hops

	JSON     string
	Platform platform.Platform
}

// Run runs the action.
func (action *Info) Run(ctx context.Context, args ...string) error {
	// Default the platform option if action.JSON was not set
	if action.Platform == "" {
		action.Platform = platform.SystemPlatform()
	}

	names := action.SetAlternateTags(args)

	formulary, err := action.Formulary(ctx)
	if err != nil {
		return err
	}

	switch {
	case action.JSON == "v1":
		return jsonV1(ctx, formulary, names)
	case action.JSON == "v2":
		slog.Error("v2 JSON not supported")
		return nil
	default:
		return action.pretty(ctx, formulary, names)
	}
}

func jsonV1(ctx context.Context, fmlry formula.Formulary, names []string) error {
	formulae, err := formula.FetchAll(ctx, fmlry, names)
	if err != nil {
		return err
	}

	for _, f := range formulae {
		switch f := f.(type) {
		case *formula.V1:
			content, err := json.Marshal(f.SourceV1())
			if err != nil {
				return err
			}
			fmt.Println(string(content))
		default:
			return errors.New("could not get v1 API data for formula " + f.Name())
		}
	}

	return nil
}

func (action *Info) pretty(ctx context.Context, fmlry formula.Formulary, names []string) error {
	formulae, err := formula.FetchAllPlatform(ctx, fmlry, names, action.Platform)
	if err != nil {
		return err
	}

	for _, f := range formulae {
		switch f := f.(type) {
		case formula.PlatformFormula:
			pretty.Info(f, action.Prefix())
		default:
			return errors.New("missing metadata for formula " + f.Name())
		}
	}

	return nil
}
