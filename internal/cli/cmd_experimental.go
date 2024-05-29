package cli

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

// copyCmd creates the command.
func copyCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Copy{
		Hops: hops,
		From: hopsv1.RegistryConfig{
			Prefix: "ghcr.io/homebrew/core",
		},
		FromAPIDomain: "https://formulae.brew.sh/api",
	}

	cmd := &cobra.Command{
		Use:   "copy ([formula]... | [--file Brewfile])",
		Short: "Copy and annotate bottles",
		Long: heredoc.Doc(`
			Copy bottles and dependencies from one registry to another. Adds a referring manifest containing all metadata available for the bottle.`),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args)
		},
	}

	// Source registry flags
	cmd.Flags().StringVar(&action.From.Prefix, "from", "ghcr.io/homebrew/core", "Source registry prefix for bottles")
	withRegistryFlags(cmd, "from", "source", &action.From)
	cmd.Flags().StringVar(&action.FromAPIDomain,
		"from-api-domain", "https://formulae.brew.sh/api",
		"Source API domain for metadata")

	// Destination registry flags
	cmd.Flags().StringVar(&action.To.Prefix, "to", "", "Destination registry prefix for bottles")
	logutil.FlagErr("to", cmd.MarkFlagRequired("to"))
	withRegistryFlags(cmd, "to", "destination", &action.To)

	// Formula flags
	cmd.Flags().StringSliceVar(&action.Brewfile, "brewfile", nil, "Copy formulae listed in a Brewfile")
	logutil.FlagErr("brewfile", cmd.MarkFlagFilename("brewfile"))

	// Dependency resolution flags
	withDependencyFlags(cmd, &action.DependencyOptions)

	return cmd
}

// imagesCmd creates the command.
func imagesCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Images{Hops: hops}

	cmd := &cobra.Command{
		Use:   "images ([formula]... | [--file Brewfile])",
		Short: "List formula dependencies",
		Long: heredoc.Doc(`
			Show dependencies as a list of bottle image references. When given multiple formula arguments, combine all images into one list.`),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}

	// Enable registry override flags
	withRegistryConfig(cmd, action.Hops)

	cmd.Flags().StringVar(&action.File, "file", "", "Find images for the formulae listed in a Brewfile")
	logutil.FlagErr("file", cmd.MarkFlagFilename("file"))

	cmd.Flags().BoolVar(&action.NoResolve, "no-resolve", false, "Do not resolve image tags")
	cmd.Flags().BoolVar(&action.NoVerify, "no-verify", false, "Do not verify tag existence (implies --no-resolve)")
	cmd.MarkFlagsMutuallyExclusive("no-resolve", "no-verify")

	// Dependency resolution flags
	withDependencyFlags(cmd, &action.DependencyOptions)

	return cmd
}
