package actions

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewapi "github.com/act3-ai/hops/internal/brew/api"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	brewreg "github.com/act3-ai/hops/internal/brew/registry"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/bottle"
	hops "github.com/act3-ai/hops/internal/hops"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
	"github.com/act3-ai/hops/internal/utils/logutil"
)

// Hops represents the base action.
type Hops struct {
	version     string   // version string set by creator
	ConfigFiles []string // sets the config files to be searched
	EnvFiles    []string // load environment variables from these files
	Concurrency int      // sets the maximum threads for any parallel tasks

	// callback functions to override runtime-loaded configuration
	configOverrides []func(cfg *hopsv1.Configuration)

	// cache for runtime-loaded objects
	cfg           *hopsv1.Configuration
	alternateTags map[string]string
	hopsclient    hops.Client
	brewformulary brewformulary.PreloadedFormulary
	brewregistry  brewreg.Registry
}

// DefaultConcurrency is the default maximum threads for parallel tasks.
const DefaultConcurrency int = 8

// NewHops creates a new Tool with default values.
func NewHops(version string) *Hops {
	return &Hops{
		version:     version,
		ConfigFiles: hopsv1.ConfigurationSearchFiles,
		EnvFiles:    brewenv.DefaultEnvironmentFiles(),
		Concurrency: DefaultConcurrency,
	}
}

// Version returns the version (overwritten by main.version if needed).
func (action *Hops) Version() string {
	return action.version
}

// MaxGoroutines produces the maximum number of
// Goroutines that should be started at a time.
func (action *Hops) MaxGoroutines() int {
	if action.Concurrency < 1 {
		return 1
	}
	return action.Concurrency
}

// Prefix produces the configured prefix.
func (action *Hops) Prefix() prefix.Prefix {
	return prefix.Prefix(action.Config().Prefix)
}

// UserAgent produces the tool's user agent string.
func (action *Hops) UserAgent() string {
	return "hops/" + action.version
}

// AddConfigOverride adds a configuration override function.
// The override function will be called when loading
// hops' configuration.
func (action *Hops) AddConfigOverride(overrides ...func(cfg *hopsv1.Configuration)) {
	action.configOverrides = append(action.configOverrides, overrides...)
}

// Config returns the Hops CLI configuration.
func (action *Hops) Config() *hopsv1.Configuration {
	if action.cfg != nil {
		return action.cfg
	}

	action.cfg = &hopsv1.Configuration{}

	// Set override functions to be called before returning
	defer func() {
		for _, override := range action.configOverrides {
			override(action.cfg)
		}

		slog.Debug("using config", slog.String("config", action.cfg.String()))
	}()

	// Load first config file found
	for _, filename := range action.ConfigFiles {
		content, err := os.ReadFile(filename)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			slog.Warn("skipping unreadable config file", slog.String("path", filename), slog.Any("reason", err))
			continue
		}

		// Regardless of if the bytes are of any external version,
		// it will be read successfully and converted into the internal version
		if err := yaml.Unmarshal(content, action.cfg); err != nil {
			// err = fmt.Errorf("loading config file %s: %w", filename, err)
			slog.Error("loading config file", slog.String("path", filename), logutil.ErrAttr(err))
			continue
		}

		slog.Debug("loaded config file", slog.String("path", filename))
		break
	}

	// Set default values for the configuration here
	hopsv1.ConfigurationDefault(action.cfg)

	return action.cfg
}

// SetAlternateVersions sets alternate tags from a list of arguments,
// and returns the isolated names from the arguments.
func (action *Hops) SetAlternateTags(args []string) (names []string) {
	action.alternateTags = map[string]string{}
	names = make([]string, 0, len(args))
	for _, arg := range args {
		name, version := parseArg(arg)
		action.alternateTags[name] = version
		names = append(names, name)
	}
	return names
}

// Formulary produces the configured Formulary.
func (action *Hops) Formulary(ctx context.Context) (formula.Formulary, error) {
	switch action.Config().Registry.Prefix {
	// Homebrew-style Formulary
	case "":
		return action.brewFormulary(ctx)
	// Hops-style Formulary
	default:
		return action.hopsClient()
	}
}

// BottleRegistry produces the configured Bottle registry.
func (action *Hops) BottleRegistry() (bottle.Registry, error) {
	switch action.Config().Registry.Prefix {
	// Homebrew-style Registry
	case "":
		return action.brewRegistry(), nil
	// Hops-style Registry
	default:
		return action.hopsClient()
	}
}

// hopsClient initializes the configured formula.Formulary/bottle.Registry.
func (action *Hops) hopsClient() (hops.Client, error) {
	if action.hopsclient == nil {
		// Initialize registry.Registry
		reg, err := hopsRegistry(&action.Config().Registry, action.UserAgent())
		if err != nil {
			return nil, err
		}

		action.hopsclient = hopsClient(
			filepath.Join(action.Config().Cache, "oci"),
			action.alternateTags,
			action.MaxGoroutines(),
			reg)
	}
	return action.hopsclient, nil
}

// brewFormulary initializes the configured formula.Formulary.
func (action *Hops) brewFormulary(ctx context.Context) (brewformulary.PreloadedFormulary, error) {
	if action.brewformulary == nil {
		// Load the index
		index, err := brewFormulary(ctx,
			action.Config().Homebrew.API.Domain,
			&action.Config().Homebrew.API.AutoUpdate,
			action.Config().Cache)
		if err != nil {
			return nil, err
		}

		action.brewformulary = index
	}
	return action.brewformulary, nil
}

func brewFormulary(ctx context.Context, domain string, cfg *brewenv.AutoUpdateConfig, cache string) (brewformulary.PreloadedFormulary, error) {
	slog.Debug("using Homebrew API formulary", //nolint:sloglint
		slog.String("HOMEBREW_API_DOMAIN", domain))
	// Load the index
	return brewformulary.FetchV1(ctx,
		brewapi.NewClient(domain),
		cache, cfg)
}

// brewRegistry initializes the configured bottle.Registry.
func (action *Hops) brewRegistry() brewreg.Registry {
	if action.brewregistry == nil {
		action.brewregistry = brewRegistry(slog.Default(), &action.Config().Homebrew, action.MaxGoroutines())
	}
	return action.brewregistry
}

func parseArg(arg string) (name, version string) {
	fields := strings.SplitN(arg, ":", 2)
	if len(fields) == 1 {
		return fields[0], ""
	}
	return fields[0], fields[1]
}

func (action *Hops) fetchFromArgs(ctx context.Context, args []string, plat platform.Platform) ([]formula.PlatformFormula, error) {
	names := action.SetAlternateTags(args)
	store, err := action.Formulary(ctx)
	if err != nil {
		return nil, err
	}
	return formula.FetchAllPlatform(ctx, store, names, plat)
}
