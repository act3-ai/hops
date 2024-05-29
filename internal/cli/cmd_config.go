package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/act3-ai/hops/internal/actions"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	"github.com/act3-ai/hops/internal/cli/doc"
	"github.com/act3-ai/hops/internal/o"
)

// shellenvCmd creates the command.
func shellenvCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.ShellEnv{Hops: hops}
	cmd := &cobra.Command{
		Use: fmt.Sprintf(
			"shellenv [%s]",
			strings.Join(actions.Shells, "|")),
		Short: "Print export statements",
		Long: heredoc.Docf(`
			Print export statements. When run in a shell, this installation of Homebrew will be added to your %s, %s, and %s.
			
			The variables %s, %s and %s are also exported to avoid querying them multiple times.
			To help guarantee idempotence, this command produces no output when Homebrew's %s and %s directories are first and second
			respectively in your %s. Consider adding evaluation of this command's output to your dotfiles (e.g. %s or
			%s on macOS and %s or %s on Linux) with: %s

			The shell can be specified explicitly with a supported shell name parameter. Unknown shells will output POSIX exports.`,
			doc.Code(o.StyleBold(`PATH`)), doc.Code(o.StyleBold(`MANPATH`)), doc.Code(o.StyleBold(`INFOPATH`)),
			doc.Code(o.StyleBold(`HOMEBREW_PREFIX`)), doc.Code(o.StyleBold(`HOMEBREW_CELLAR`)), doc.Code(o.StyleBold(`HOMEBREW_REPOSITORY`)),
			doc.Code(o.StyleBold(`bin`)), doc.Code(o.StyleBold(`sbin`)), doc.Code(o.StyleBold(`PATH`)),
			doc.Code(o.StyleBold(`~/.bash_profile`)),
			doc.Code(o.StyleBold(`~/.zprofile`)),
			doc.Code(o.StyleBold(`~/.bashrc`)),
			doc.Code(o.StyleBold(`~/.zshrc`)),
			doc.CodeBlock("sh", o.StyleBold(`eval "$(brew shellenv)"`)),
		),
		Args:          cobra.MaximumNArgs(1),
		ValidArgs:     actions.Shells,
		Annotations:   map[string]string{},
		SilenceErrors: true, // shellenv command handles error output on its own
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				action.Shell = args[0]
			}
			err := action.Run(cmd.Context())
			if err != nil {
				// This will yell at the user every time the output is evaluated
				cmd.Printf(`echo "ERROR(hops shellenv): %s\n"`, err.Error())
				return err
			}
			return nil
		},
	}

	return cmd
}

// envCmd creates the command.
func envCmd(hops *actions.Hops) *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Show environment config",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			// cmd.Println("# Homebrew configuration:")
			// cmd.Println(hops.Config())
			// cmd.Println("# Hops configuration:")
			cmd.Println(hops.Config())
		},
	}
}

// prefixCmd creates the command.
func prefixCmd(hops *actions.Hops) *cobra.Command {
	return &cobra.Command{
		Use:   "prefix",
		Short: "Show prefix",
		Args:  cobra.NoArgs,
		Run:   func(cmd *cobra.Command, _ []string) { cmd.Println(hops.Prefix()) },
	}
}

// cellarCmd creates the command.
func cellarCmd(hops *actions.Hops) *cobra.Command {
	return &cobra.Command{
		Use:   "cellar",
		Short: "Show Cellar",
		Args:  cobra.NoArgs,
		Run:   func(cmd *cobra.Command, _ []string) { cmd.Println(hops.Prefix().Cellar()) },
	}
}

// cleanupCmd creates the command.
func cleanupCmd(hops *actions.Hops) *cobra.Command {
	action := &actions.Cleanup{Hops: hops}
	cmd := &cobra.Command{
		Use:   "cleanup [formula|cask ...]",
		Short: "Clean up outdated files",
		Long: heredoc.Doc(`
			Remove stale lock files and outdated downloads for all formulae and casks, and
			remove old versions of installed formulae. If arguments are specified, only do
			this for the given formulae and casks. Removes all downloads more than 120 days
			old. This can be adjusted with HOMEBREW_CLEANUP_MAX_AGE_DAYS.`),
		// 		--prune                      Remove all cache files older than specified
		// 											  days. If you want to remove everything, use
		// 											  --prune=all.
		//   -n, --dry-run                    Show what would be removed, but do not
		// 											  actually remove anything.
		//   -s                               Scrub the cache, including downloads for even
		// 											  the latest versions. Note that downloads for
		// 											  any installed formulae or casks will still
		// 											  not be deleted. If you want to delete those
		// 											  too: rm -rf "$(brew --cache)"
		// 		--prune-prefix               Only prune the symlinks and directories from
		// 											  the prefix and remove no other files.
		//   -d, --debug                      Display any debugging information.
		//   -q, --quiet                      Make some output more quiet.
		//   -v, --verbose                    Make some output more verbose.
		//   -h, --help                       Show this message.
		ValidArgsFunction: installedFormulae(hops),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return action.Run(cmd.Context())
		},
	}
	return cmd
}

// configCmd creates the command.
func configCmd(hops *actions.Hops) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management command group",
	}

	cmd.AddCommand(
		configInitCmd(hops),
	)

	return cmd
}

// configInitCmd creates the command.
func configInitCmd(hops *actions.Hops) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write current configuration to config file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := hops.Config()

			content, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}

			cfgfile := hopsv1.ConfigurationFile

			err = os.MkdirAll(filepath.Dir(cfgfile), 0o777)
			if err != nil {
				return fmt.Errorf("creating config file parent directory: %w", err)
			}

			err = os.WriteFile(cfgfile, content, 0o644)
			if err != nil {
				return fmt.Errorf("writing config file: %w", err)
			}

			o.Hai("Wrote " + cfgfile)
			return nil
		},
	}
}
