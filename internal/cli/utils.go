package cli

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	"github.com/act3-ai/hops/internal/formula"
)

// AutocompleteFormulae returns an autocompletion function that suggests formula names.
func AutocompleteFormulae(ctx context.Context, action *actions.Hops) func(s string) []string {
	return func(_ string) []string {
		index, err := action.Formulary(ctx)
		if err != nil {
			cobra.CompErrorln("loading completions: " + err.Error())
			return []string{}
		}

		switch index := index.(type) {
		case brewformulary.PreloadedFormulary:
			return index.ListNames()
		default:
			cobra.CompErrorln("completions not available for standalone regsistry mode")
			return []string{}
		}
	}
}

// CompleteInstalledFormulae returns an autocompletion function that suggests formula names.
func CompleteInstalledFormulae(_ context.Context, action *actions.Hops) func(s string) []string {
	return func(_ string) []string {
		entries, err := os.ReadDir(action.Prefix().Cellar())
		if errors.Is(err, os.ErrNotExist) {
			return []string{}
		} else if err != nil {
			cobra.CompErrorln("loading completions: checking cellar: " + err.Error())
			return []string{}
		}

		hits := []string{}
		for _, d := range entries {
			if d.IsDir() {
				hits = append(hits, d.Name())
			}
		}

		return hits
	}
}

// formulaNames produces the autocompletion function for formula names.
func formulaNames(hops *actions.Hops) func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return AutocompleteFormulae(cmd.Context(), hops)(toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// installedFormulae produces the autocompletion function for installed formula names.
func installedFormulae(hops *actions.Hops) func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return CompleteInstalledFormulae(cmd.Context(), hops)(toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// withRegistryFlags adds a flag to override the registry field in config.
func withRegistryFlags(cmd *cobra.Command, action *actions.Hops) {
	var registry string
	var plainHTTP bool
	cmd.Flags().StringVarP(&registry, "registry", "r", "", "Registry prefix for bottles")
	cmd.Flags().BoolVar(&plainHTTP, "plain-http", false, "Allow insecure connections to registry without SSL check")
	action.AddConfigOverride(func(cfg *hopsv1.Configuration) {
		if registry != "" {
			cfg.Registry.Prefix = registry
		}
		if cmd.Flags().Lookup("plain-http").Changed {
			cfg.Registry.PlainHTTP = plainHTTP
		}
	})
}

// withDependencyFlags adds flags for dependency resolution.
func withDependencyFlags(cmd *cobra.Command, opts *formula.DependencyTags) {
	cmd.Flags().BoolVar(&opts.IncludeBuild, "include-build", false, "Include :build dependencies for formula")
	cmd.Flags().BoolVar(&opts.IncludeOptional, "include-optional", false, "Include :optional dependencies for formula")
	cmd.Flags().BoolVar(&opts.IncludeTest, "include-test", false, "Include :test dependencies for formula (non-recursive)")
	cmd.Flags().BoolVar(&opts.SkipRecommended, "skip-recommended", false, "Skip :recommended dependencies for formula")
}
