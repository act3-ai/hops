package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	"github.com/act3-ai/hops/internal/o"
)

// infoCmd creates the command.
func infoCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Info{Hops: hops}
	cmd := &cobra.Command{
		Use:               "info [formula]...",
		Short:             "View formula metadata",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}
	cmd.Flags().StringVar(&action.JSON, "json", "", "Print a JSON representation")
	cmd.Flags().Lookup("json").NoOptDefVal = "v1"

	// Platform selector (not reflected in Homebrew)
	platflag := cmd.Flags().VarPF(&action.Platform, "platform", "p", "View dependencies on platform")
	platflag.DefValue = "system"

	return cmd
}

// searchCmd creates the command.
func searchCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Search{Hops: hops}
	text := o.StyleUnderline("text")
	regex := o.StyleUnderline("regex")
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("search [%s|/%s/]...", text, regex),
		Short: "Search available formulae",
		Long: heredoc.Docf(`
			Perform a substring search of cask tokens and formula names for %s. If %s is flanked by slashes, it is interpreted as a regular expression.`,
			text, text),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}
	cmd.Flags().BoolVar(&action.Desc, "desc", false, "Search for formulae with a description matching "+text)
	return cmd
}

// depsCmd creates the command.
func depsCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Deps{Hops: hops}

	var tree bool

	cmd := &cobra.Command{
		Use:               "deps [formula]...",
		Short:             "View formula dependencies",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			switch {
			case tree:
				return action.Tree(ctx, args...)
			default:
				return action.Run(ctx, args...)
			}
		},
	}

	// Mode switch flags
	cmd.Flags().BoolVar(&tree, "tree", false, "Show dependencies as a tree. When given multiple formula arguments, show individual trees for each formula.")

	// Platform selector (not reflected in Homebrew)
	platflag := cmd.Flags().VarPF(&action.Platform, "platform", "p", "View dependencies on platform")
	platflag.DefValue = "system"

	// Dependency resolution flags
	withDependencyFlags(cmd, &action.DependencyOptions)

	return cmd
}

// leavesCmd creates the command.
func leavesCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Leaves{Hops: hops}

	cmd := &cobra.Command{
		Use:   "leaves [formula]...",
		Short: "List installed formulae that are not dependencies of another installed formula",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return action.Run(cmd.Context())
		},
	}

	// cmd.Flags().BoolVarP(&action.InstalledOnRequest, "installed-on-request", "r", false, "Only list leaves that were manually installed")
	// cmd.Flags().BoolVarP(&action.InstalledAsDependency, "installed-as-dependency", "p", false, "Only list leaves that were installed as dependencies")
	// cmd.MarkFlagsMutuallyExclusive("installed-on-request", "installed-as-dependency")

	return cmd
}

// listCmd creates the command.
func listCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.List{Hops: hops}

	cmd := &cobra.Command{
		Use:               "list [installed_formula|installed_cask ...]",
		Aliases:           []string{"ls", "formulae"},
		Short:             "List installed formulae",
		Long:              "List all installed formulae. If formula is provided, summarise the paths within its current keg.",
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}

	// cmd.Flags().BoolVar(&action.FullName, "full-name", false, "Print formulae with fully-qualified names. Unless --full-name, --versions or --pinned are passed, other options (i.e. -1, -l, -r and -t) are passed to ls(1) which produces the actual output")
	cmd.Flags().BoolVar(&action.Versions, "versions", false, "Show the version number for installed formulae, or only the specified formulae if formula are provided")
	cmd.Flags().BoolVar(&action.Multiple, "multiple", false, "Only show formulae with multiple versions installed")
	// cmd.Flags().BoolVar(&action.Pinned, "pinned", false, "List only pinned formulae, or only the specified (pinned) formulae if formula are provided. See also pin, unpin")
	// cmd.Flags().BoolVar(&action.OnePerLine, "1", false, "Force output to be one entry per line. This is the default when output is not to a terminal")
	// cmd.MarkFlagsMutuallyExclusive("versions", "1")
	// cmd.Flags().BoolVarP(&action.Long, "long", "l", false, "List formulae and/or casks in long format. Has no effect when a formula or cask name is passed as an argument")
	// cmd.MarkFlagsMutuallyExclusive("versions", "long")

	return cmd
}
