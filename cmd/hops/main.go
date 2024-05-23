package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/docs"
	"github.com/act3-ai/hops/internal/o"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	vv "gitlab.com/act3-ai/asce/go-common/pkg/version"

	"github.com/act3-ai/hops/internal/cli"
)

// Retrieves build info.
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

func main() {
	info := getVersionInfo()         // Load the version info from the build
	root := cli.NewCLI(info.Version) // Create the root command
	root.SilenceUsage = true         // Silence usage when root is called

	// Layout of embedded documentation to surface in the help command
	// and generate in the gendocs command
	embeddedDocs := docs.Embedded(root)

	docsCmd := commands.NewInfoCmd(embeddedDocs)
	docsCmd.Use = "docs"

	// Add common commands
	root.AddCommand(
		commands.NewVersionCmd(info),
		commands.NewGendocsCmd(embeddedDocs),
		docsCmd,
	)

	// Restores the original ANSI processing state on Windows
	var restoreWindowsANSI func() error

	// Store persistent pre run function to avoid overwriting it
	persistentPreRun := root.PersistentPreRun

	// The pre run function logs build info and sets the default output writer
	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		level := cli.LogLevel(cmd, slog.LevelInfo) // parse log level flags
		// cmd.Println("log level: ", level)
		logger := log.NewWithOptions(cmd.OutOrStdout(), log.Options{
			TimeFormat: time.Kitchen,     // set human-readable time
			Level:      log.Level(level), // parse verbose/debug flags
		})
		logger.SetStyles(o.LogStyles())          // set homebrew-inspired styles
		log.SetDefault(logger)                   // set as default charmbracelet/log.Logger
		slog.SetDefault(slog.New(log.Default())) // set as the default slog.Logger

		// Set termenv default output
		termenv.SetDefaultOutput(termenv.NewOutput(cmd.OutOrStdout()))
		// Enable ANSI processing on Windows color output
		var err error
		restoreWindowsANSI, err = termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
		if err != nil {
			slog.Error("error enabling ANSI processing", slog.String("error", err.Error()))
		}

		slog.Debug("Software", slog.String("version", info.Version)) // Log version info
		slog.Debug("Software details", slog.Any("info", info))       // Log build info

		if persistentPreRun != nil {
			persistentPreRun(cmd, args)
		}
	}

	// The post run function restores the terminal
	root.PersistentPostRun = func(_ *cobra.Command, _ []string) {
		// Restore original ANSI processing state on Windows
		if err := restoreWindowsANSI(); err != nil {
			slog.Error("error restoring ANSI processing state", slog.String("error", err.Error()))
		}
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
