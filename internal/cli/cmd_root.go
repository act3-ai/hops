package cli

import (
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	"github.com/act3-ai/hops/internal/cli/doc"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils/env"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
)

// NewCLI creates the CLI.
func NewCLI(version string) *cobra.Command {
	// Create hops with scheme initialized
	hops := actions.NewHops(version)

	// cmd represents the base command when called without any subcommands
	cmd := &cobra.Command{
		Use:     "hops",
		Version: version,
		Short:   "Homebrew OCI Package Sipper",
		Long: heredoc.Doc(`
			Hops is an experimental client for installing Homebrew bottles.

		`) + doc.Footer(o.StyleYellow(heredoc.Doc(`
			[CAUTION]: Hops is a experimental software. Package management is a complex task 
			           and there are serious risks with modifying your packages.`))),
	}

	// Run all persistent hooks (not just the first found)
	cobra.EnableTraverseRunHooks = true

	// Concurrency flag
	cmd.PersistentFlags().IntVar(&hops.Concurrency, "concurrency", runtime.NumCPU(), "Concurrency level")

	// Style the error prefix red
	cmd.SetErrPrefix(o.StyleRed(cmd.ErrPrefix()))

	// Add config overrides
	withConfigOverrides(hops, cmd)

	commands.AddGroupedCommands(cmd,
		&cobra.Group{
			ID:    "install",
			Title: "Install and update formulae",
		},
		installCmd(hops),
		uninstallCmd(hops),
		updateCmd(hops),
		upgradeCmd(hops),
	)

	commands.AddGroupedCommands(cmd,
		&cobra.Group{
			ID:    "manage",
			Title: "Manage installed formulae",
		},
		linkCmd(hops),
		unlinkCmd(hops),
		listCmd(hops),
		leavesCmd(hops),
	)

	commands.AddGroupedCommands(cmd,
		&cobra.Group{
			ID:    "explore",
			Title: "Explore available formulae",
		},
		infoCmd(hops),
		depsCmd(hops),
		searchCmd(hops),
	)

	commands.AddGroupedCommands(cmd,
		&cobra.Group{
			ID:    "config",
			Title: "Manage Hops and Homebrew config",
		},
		shellenvCmd(hops),
		envCmd(hops),
		cleanupCmd(hops),
		prefixCmd(hops),
		cellarCmd(hops),
		configCmd(hops),
	)

	commands.AddGroupedCommands(cmd,
		&cobra.Group{
			ID:    "registry",
			Title: "Manage registries of Homebrew Bottles",
		},
		imagesCmd(hops),
		copyCmd(hops),
	)

	return cmd
}

// withConfigOverrides adds config overrides.
func withConfigOverrides(action *actions.Hops, cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVar(&action.ConfigFiles, "config",
		env.PathSlice(hopsv1.ConfigurationEnvName, action.ConfigFiles),
		"Set config file search paths")
	cfgfiles := make([]string, len(hopsv1.UnevaluatedSearchFiles))
	for i := range hopsv1.UnevaluatedSearchFiles {
		cfgfiles[i] = doc.Code(hopsv1.UnevaluatedSearchFiles[i])
	}

	// Set default value so the environment is not queried before output
	cmd.PersistentFlags().Lookup("config").DefValue = strings.Join(cfgfiles, ",")
	// Add environment variable overrides
	action.AddConfigOverride(hopsv1.ConfigurationEnvOverrides)
}
