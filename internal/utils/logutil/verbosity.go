package logutil

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// VerbosityOptions configures logger verbosity.
type VerbosityOptions struct {
	Debug   int // number of --debug flags passed
	Quiet   int // number of --quiet flags passed
	Verbose int // number of --verbose flags passed
}

// WithPersistentVerbosityFlags adds the debug, quiet, and verbose flags to the given command.
func WithPersistentVerbosityFlags(cmd *cobra.Command) *VerbosityOptions {
	v := &VerbosityOptions{}
	cmd.PersistentFlags().CountVarP(&v.Debug, "debug", "d", "Display more debugging information")
	cmd.PersistentFlags().CountVarP(&v.Quiet, "quiet", "q", "Make some output more quiet")
	cmd.PersistentFlags().CountVarP(&v.Verbose, "verbose", "v", "Make some output more verbose")
	cmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
	return v
}

// LogLevel produces the desired slog.Level by parsing the "debug", "verbose", and "quiet" flags.
//
// "base" sets the default level, which gets modified by the flags.
func (v *VerbosityOptions) LogLevel(base slog.Level) slog.Level {
	return base + // Starts at default level
		slog.Level(-4*v.Debug) + // Debug flag increases the the log threshold by a full level
		slog.Level(-v.Verbose) + // Verbose flag lowers the log threshold by one numeric step
		slog.Level(v.Quiet) // Quiet flag raises the log threshold by one numeric step
}

// Logs an error when setting up flags.
func FlagErr(name string, err error) {
	if err != nil {
		slog.Warn("flag error", slog.String("flag", name), ErrAttr(err))
	}
}
