package cli

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	"github.com/act3-ai/hops/internal/o"
)

// installCmd creates the command
func installCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Install{Hops: hops}

	cmd := &cobra.Command{
		Use:   "install formula [...]",
		Short: "Install a formula",
		Long: heredoc.Doc(`
			Install a formula. Additional options specific to a formula may be appended to the command.

			Unless HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK is set, brew upgrade or brew
			reinstall will be run for outdated dependents and dependents with broken
			linkage, respectively.

			Unless HOMEBREW_NO_INSTALL_CLEANUP is set, brew cleanup will then be run for
			the installed formulae or, every 30 days, for all formulae.

			Unless HOMEBREW_NO_INSTALL_UPGRADE is set, brew install formula will
			upgrade formula if it is already installed but outdated.
			
			STANDALONE MODE:

			Hops has an alternate mode to fetch all packages and metadata from a single OCI registry.
			The default behavior for standalone mode is to install the version tagged "latest".
			The tag for a formula can be set by using the argument format "<formula>:<tag>".
			`),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: formulaNames(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}

	withRegistryFlags(cmd, action.Hops)

	cmd.Flags().BoolVar(&action.Force, "force", false, "Install formulae without checking for previously installed keg-only or non-migrated versions. When installing casks, overwrite existing files (binaries and symlinks are excluded, unless originally from the same cask)")
	cmd.Flags().BoolVar(&action.DryRun, "dry-run", false, "Show what would be installed, but do not actually install anything")
	cmd.Flags().BoolVar(&action.Overwrite, "overwrite", false, "Delete files that already exist in the prefix while linking")

	// Dependency resolution flags
	withDependencyFlags(cmd, &action.DependencyOptions)
	cmd.Flags().BoolVar(&action.IgnoreDependencies, "ignore-dependencies", false, "Skip installing any dependencies of any kind [TESTING-ONLY]")
	cmd.Flags().BoolVar(&action.OnlyDependencies, "only-dependencies", false, "Install the dependencies with specified options but do not install the formula itself")
	cmd.MarkFlagsMutuallyExclusive("ignore-dependencies", "only-dependencies")

	return cmd
}

// uninstallCmd creates the command
func uninstallCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Uninstall{Hops: hops}

	cmd := &cobra.Command{
		Use:   "uninstall formula [...]",
		Short: "Uninstall a formula",
		Long: heredoc.Doc(`
			`),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args)
		},
	}

	return cmd
}

// updateCmd creates the command
func updateCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Update{Hops: hops}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update formula index",
		Long: heredoc.Doc(`
			Fetch the newest version of all formulae from the Homebrew API and perform any necessary migrations.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return action.Run(cmd.Context())
		},
	}
	return cmd
}

// upgradeCmd creates the command
func upgradeCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Upgrade{Hops: hops}
	cmd := &cobra.Command{
		Use:               "upgrade [formula]...",
		Short:             "Upgrade installed formulae",
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}
	return cmd
}

// linkCmd creates the command
func linkCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Link{Hops: hops}
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("link %s...", o.StyleUnderline("installed_formulae")),
		Aliases: []string{"ln"},
		Short:   "Link an installed formula",
		Long: heredoc.Doc(`
			Symlink all of formula's installed files into Homebrew's prefix. This is done
			automatically when you install formulae but can be useful for manual
			installations.`),
		Args:              cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}

	cmd.Flags().BoolVar(&action.Overwrite, "overwrite", false, "Delete files that already exist in the prefix while linking")
	cmd.Flags().BoolVarP(&action.DryRun, "dry-run", "n", false, "List files which would be linked or deleted by hops link --overwrite without actually linking or deleting any files")
	cmd.Flags().BoolVarP(&action.Force, "force", "f", false, "Allow keg-only formulae to be linked")
	cmd.Flags().BoolVar(&action.Head, "HEAD", false, "Link the HEAD version of the formula if it is installed")

	return cmd
}

// unlinkCmd creates the command
func unlinkCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Unlink{Hops: hops}
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("unlink %s...", o.StyleUnderline("installed_formulae")),
		Aliases: []string{"ln"},
		Short:   "Unlink an installed formula",
		Long: heredoc.Doc(`
			Remove symlinks for formula from Homebrew's prefix. This can be useful for
			temporarily disabling a formula: 
			
			brew unlink formula && commands && brew link formula`),
		Args:              cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), args...)
		},
	}

	cmd.Flags().BoolVarP(&action.DryRun, "dry-run", "n", false, "List files which would be unlinked without actually unlinking or deleting any files")

	return cmd
}
