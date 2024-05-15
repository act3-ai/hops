package actions

import (
	"context"
	"fmt"

	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/pretty"
)

// Info represents the action and its options
type Info struct {
	*Hops

	JSON     string
	Platform platform.Platform
}

// Run runs the action
func (action *Info) Run(ctx context.Context, names ...string) error {
	if action.Platform == "" {
		action.Platform = platform.SystemPlatform()
	}

	index := action.Index()
	err := index.Load(ctx)
	if err != nil {
		return err
	}

	formulae, err := action.FetchAll(o.Noop, index, names...)
	if err != nil {
		return err
	}

	for _, f := range formulae {
		switch action.JSON {
		case "v1":
			fmt.Println(f)
		case "v2":
			fmt.Println("v2 JSON not supported")
		default:
			pretty.Info(f, action.Prefix(), action.Platform)
		}
	}

	return nil
}
