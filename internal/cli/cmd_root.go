package cli

import (
	"log/slog"
	"runtime"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	"github.com/act3-ai/hops/internal/cli/doc"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils/env"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
)

// NewCLI creates the base hops command
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

	// Add "verbose" and "debug" flags
	withVerbosityOverrides(hops, cmd)
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
			ID:    "experimental",
			Title: "Experimental commands for working with Homebrew Bottles",
		},
		imagesCmd(hops),
		xinstallCmd(hops),
		copyCmd(hops),
	)

	return cmd
}

// withVerbosityOverrides adds the debug, quiet, and verbose flags to the given command
func withVerbosityOverrides(action *actions.Hops, cmd *cobra.Command) {
	var debug, quiet, verbose int
	cmd.PersistentFlags().CountVarP(&debug, "debug", "d", "Display any debugging information")
	cmd.PersistentFlags().CountVarP(&quiet, "quiet", "q", "Make some output more quiet")
	cmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Make some output more verbose")
	action.AddHomebrewOverride(func(e *brewenv.Environment) {
		e.Debug = e.Debug || debug > 0
		e.Verbose = e.Verbose || verbose > 0
	})
}

// withConfigOverrides adds config overrides
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

// LogLevel produces the desired slog.Level by parsing the "debug", "verbose", and "quiet" flags.
//
// "base" sets the default level, which gets modified by the flags.
func LogLevel(cmd *cobra.Command, base slog.Level) slog.Level {
	level := base // start with default level

	// Debug flag increases the the log threshold by a full level
	debug, err := strconv.Atoi(cmd.Flags().Lookup("debug").Value.String())
	if err != nil {
		panic(err)
	}
	level -= 4 * slog.Level(debug)

	// Verbose flag lowers the log threshold by one numeric step
	verbose, err := strconv.Atoi(cmd.Flags().Lookup("verbose").Value.String())
	if err != nil {
		panic(err)
	}
	level -= 1 * slog.Level(verbose)

	// Verbose flag raises the log threshold by one numeric step
	quiet, err := strconv.Atoi(cmd.Flags().Lookup("quiet").Value.String())
	if err != nil {
		panic(err)
	}
	level += 1 * slog.Level(quiet)

	return level
}
