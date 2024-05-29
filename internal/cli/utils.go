package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/act3-ai/hops/internal/actions"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/utils/logutil"
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

// withDependencyFlags adds flags for dependency resolution.
func withDependencyFlags(cmd *cobra.Command, opts *formula.DependencyTags) {
	cmd.Flags().BoolVar(&opts.IncludeBuild, "include-build", false, "Include :build dependencies for formula")
	cmd.Flags().BoolVar(&opts.IncludeOptional, "include-optional", false, "Include :optional dependencies for formula")
	cmd.Flags().BoolVar(&opts.IncludeTest, "include-test", false, "Include :test dependencies for formula (non-recursive)")
	cmd.Flags().BoolVar(&opts.SkipRecommended, "skip-recommended", false, "Skip :recommended dependencies for formula")
}

const (
	DistributionSpecReferrersTagV1_1 = "v1.1-referrers-tag" // Referrers tag fallback
	DistributionSpecReferrersAPIV1_1 = "v1.1-referrers-api" // Referrers API
)

// DistributionSpec option struct which implements pflag.Value interface.
//
// Indivates the preference of the implementation of the Referrers API:
//   - "v1.1-referrers-api" for referrers API
//   - "v1.1-referrers-tag" for referrers tag scheme
//   - "" for auto fallback
type DistributionSpec string

// Set validates and sets the flag value from a string argument.
func (ds *DistributionSpec) Set(value string) error {
	switch value {
	case DistributionSpecReferrersTagV1_1,
		DistributionSpecReferrersAPIV1_1:
		*ds = DistributionSpec(value)
	default:
		return fmt.Errorf("unknown distribution specification flag: %s", value)
	}
	return nil
}

// Type returns the string value of the inner flag.
func (*DistributionSpec) Type() string {
	return "string"
}

// Options returns the string of usable options for the flag.
func (*DistributionSpec) Options() string {
	return strings.Join([]string{
		DistributionSpecReferrersTagV1_1,
		DistributionSpecReferrersAPIV1_1,
	}, ", ")
}

// String returns the string representation of the flag.
func (ds *DistributionSpec) String() string {
	return string(*ds)
}

func withRegistryFlags(cmd *cobra.Command, prefix, description string, opts *hopsv1.RegistryConfig) {
	flagPrefix, notePrefix := applyPrefix(prefix, description)

	// // var distSpec DistributionSpec
	// distSpec := DistributionSpec(opts.DistributionSpec)
	// cmd.Flags().Var(&distSpec,
	// 	flagPrefix+"distribution-spec",
	// 	"Set OCI distribution spec version and API option for "+notePrefix+"target. Options: "+distSpec.Options())

	cmd.Flags().StringArrayVar(&opts.Headers,
		flagPrefix+"header", nil,
		"Add custom headers to "+notePrefix+"requests")

	cmd.Flags().BoolVar(&opts.OCILayout,
		flagPrefix+"oci-layout", false,
		"Set "+notePrefix+"target as an OCI image layout")

	cmd.Flags().BoolVar(&opts.PlainHTTP,
		flagPrefix+"plain-http", false,
		"Allow insecure connections to "+notePrefix+"registry without SSL check")

	registryConfigFlagName := flagPrefix + "registry-config"
	cmd.Flags().StringVar(&opts.Config,
		registryConfigFlagName, "",
		"Path of the authentication file for "+notePrefix+"registry")
	if err := cmd.MarkFlagFilename(registryConfigFlagName); err != nil {
		slog.Warn("flag error", slog.String("flag", registryConfigFlagName), logutil.ErrAttr(err))
	}
}

func withRegistryConfig(cmd *cobra.Command, action *actions.Hops) {
	regcfg := &hopsv1.RegistryConfig{}

	withRegistryFlags(cmd, "", "", regcfg)

	action.AddConfigOverride(func(cfg *hopsv1.Configuration) {
		if regcfg.Prefix != "" {
			cfg.Registry.Prefix = regcfg.Prefix
		}
		// if regcfg.DistributionSpec != "" {
		// 	cfg.Registry.DistributionSpec = regcfg.DistributionSpec
		// }
		if regcfg.Headers != nil {
			cfg.Registry.Headers = regcfg.Headers
		}
		if cmd.Flags().Lookup("oci-layout").Changed {
			cfg.Registry.OCILayout = regcfg.OCILayout
		}
		if cmd.Flags().Lookup("plain-http").Changed {
			cfg.Registry.PlainHTTP = regcfg.PlainHTTP
		}
		if regcfg.Config != "" {
			cfg.Registry.Config = regcfg.Config
		}
	})
}

func applyPrefix(prefix, description string) (flagPrefix, notePrefix string) {
	if prefix == "" {
		return "", ""
	}
	return prefix + "-", description + " "
}
