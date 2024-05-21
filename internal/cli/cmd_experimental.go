package cli

import (
	"log/slog"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	"github.com/act3-ai/hops/internal/o"
)

// copyCmd creates the command
func copyCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Copy{
		Hops: hops,
		From: "ghcr.io/homebrew/core",
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

	cmd.Flags().StringVar(&action.From, "from", "ghcr.io/homebrew/core", "Source registry prefix for bottles")
	cmd.Flags().BoolVar(&action.FromOCILayout, "from-oci-layout", false, "Set source target as an OCI image layout")
	cmd.Flags().BoolVar(&action.FromPlainHTTP, "from-plain-http", false, "Allow insecure connections to source registry without SSL check")

	cmd.Flags().StringVar(&action.To, "to", "", "Destination registry prefix for bottles")
	cobra.CheckErr(cmd.MarkFlagRequired("to"))
	cmd.Flags().BoolVar(&action.ToOCILayout, "to-oci-layout", false, "Set destination target as an OCI image layout")
	cmd.Flags().BoolVar(&action.ToPlainHTTP, "to-plain-http", false, "Allow insecure connections to destination registry without SSL check")

	cmd.Flags().StringVar(&action.File, "file", "", "Copy formulae listed in a Brewfile")
	if err := cmd.MarkFlagFilename("file"); err != nil {
		slog.Info("flag error", o.ErrAttr(err))
	}

	// Dependency resolution flags
	newWithDependencyFlags(cmd, &action.DependencyOptions)

	return cmd
}

// imagesCmd creates the command
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

	withRegistryFlags(cmd, action.Hops)

	cmd.Flags().StringVar(&action.File, "file", "", "Find images for the formulae listed in a Brewfile")
	if err := cmd.MarkFlagFilename("file"); err != nil {
		slog.Info("flag error", o.ErrAttr(err))
	}
	cmd.Flags().BoolVar(&action.NoResolve, "no-resolve", false, "Do not resolve image tags")
	cmd.Flags().BoolVar(&action.NoVerify, "no-verify", false, "Do not verify tag existence (implies --no-resolve)")
	cmd.MarkFlagsMutuallyExclusive("no-resolve", "no-verify")

	// Dependency resolution flags
	newWithDependencyFlags(cmd, &action.DependencyOptions)

	return cmd
}
