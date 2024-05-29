package main

import (
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/docs"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils/logutil"

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

	// Logger flags
	logopts := logutil.WithPersistentVerbosityFlags(root) // Add "quiet", "verbose", and "debug" flags
	var logfmt string
	root.PersistentFlags().StringVar(&logfmt, "log-fmt", "text", "Set format for log messages. Options: text, json")
	// var logfile string
	// root.PersistentFlags().StringVar(&logfile, "log-file", "", "Send JSON logs to a file")

	// Restores the original ANSI processing state on Windows
	var restoreWindowsANSI func() error

	// The pre run function logs build info and sets the default output writer
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		// Create [log/slog.Handler] using flag configuration
		level := logopts.LogLevel(slog.LevelInfo) // evaluate log level flags
		handler, err := newTermHandler(
			cmd.ErrOrStderr(),
			level,  // log level flags
			logfmt, // log format flag
			// logfile, // log file flag
		)
		if err != nil {
			return err
		}

		// Set default [log/slog.Logger]
		slog.SetDefault(slog.New(handler))

		// Set termenv default output
		termenv.SetDefaultOutput(termenv.NewOutput(cmd.OutOrStdout()))
		// Enable ANSI processing on Windows color output
		restoreWindowsANSI, err = termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
		if err != nil {
			slog.Error("error enabling ANSI processing", slog.String("error", err.Error()))
		}

		// slog.Log(ctx, logutil.LevelTrace,
		slog.Log(ctx, slog.LevelDebug,
			"Starting logger", slog.String("format", "text"), slog.String("verbosity", level.String()))
		// slog.Log(ctx, logutil.LevelVerbose,
		slog.Log(ctx, slog.LevelDebug,
			"Software", slog.String("version", info.Version)) // Log version info
		slog.Debug("Software details", slog.Attr{ // Log build info
			Key:   "info",
			Value: slog.GroupValue(logutil.VersionAttrs(info)...),
		})

		return nil
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

func newTermHandler(w io.Writer, level slog.Level, format string) (slog.Handler, error) {
	opts := log.Options{
		Level: log.Level(level), // parse verbose/debug flags
	}

	switch format {
	case "text", "":
		opts.Formatter = log.TextFormatter
	case "logfmt":
		opts.Formatter = log.LogfmtFormatter
		opts.ReportTimestamp = true
	case "json":
		opts.Formatter = log.JSONFormatter
		opts.ReportTimestamp = true
	default:
		return nil, errors.New("starting logger: unsupported log format \"" + format + "\"")
	}

	logger := log.NewWithOptions(w, opts)
	logger.SetStyles(o.LogStyles()) // set homebrew-inspired styles
	log.SetDefault(logger)          // set as default charmbracelet/log.Logger

	return logger, nil
}

/*
func logHandler(level slog.Level, format, file string) (slog.Handler, error) {
	var termHandler slog.Handler

	switch format {
	// Text logger using charmbracelet/log
	case "text", "":
		logger := log.NewWithOptions(os.Stdout, log.Options{
			// TimeFormat: time.Kitchen,     // set human-readable time
			Level: log.Level(level), // parse verbose/debug flags
		})
		logger.SetStyles(o.LogStyles()) // set homebrew-inspired styles
		log.SetDefault(logger)          // set as default charmbracelet/log.Logger
		termHandler = logger            // use as termHandler
	// JSON logger using zerolog
	case "json":
		zlogger := zerolog.New(os.Stderr)
		termHandler = slogzerolog.Option{
			Level:  level,
			Logger: &zlogger,
		}.NewZerologHandler()
	default:
		return nil, errors.New("starting logger: unsupported log format \"" + format + "\"")
	}

	// Return terminal handler if no file is specified
	if file == "" {
		return termHandler, nil
	}

	// Create log file
	f, err := os.Create(file)
	if err != nil {
		return nil, fmt.Errorf("creating log file: %w", err)
	}

	// Create file logger
	fileLogger := zerolog.New(f)

	// Create log/slog.Handler
	fileHandler := slogzerolog.Option{
		Level:  logutil.LevelTrace,
		Logger: &fileLogger,
	}.NewZerologHandler()

	// Create fanout logger to duplicate messages to file and terminal handler
	return slogmulti.Fanout(termHandler, fileHandler), nil
}
*/
