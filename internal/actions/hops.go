package actions

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewapi "github.com/act3-ai/hops/internal/brew/api"
	brewformulary "github.com/act3-ai/hops/internal/brew/formulary"
	brewreg "github.com/act3-ai/hops/internal/brew/registry"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/bottle"
	hops "github.com/act3-ai/hops/internal/hops"
	hopsreg "github.com/act3-ai/hops/internal/hops/registry"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/platform"
	"github.com/act3-ai/hops/internal/prefix"
)

// Hops represents the base action
type Hops struct {
	version     string   // version string set by creator
	ConfigFiles []string // sets the config files to be searched
	EnvFiles    []string // load environment variables from these files
	Concurrency int      // sets the maximum threads for any parallel tasks

	// callback functions to override runtime-loaded configuration
	brewOverrides   []func(e *brewenv.Environment)
	configOverrides []func(cfg *hopsv1.Configuration)

	// cache for runtime-loaded objects
	authClient *auth.Client
	cfg        *hopsv1.Configuration
	brewcfg    *brewenv.Environment

	alternateTags map[string]string
	hopsclient    hops.Client
	brewclient    struct {
		formulary brewformulary.PreloadedFormulary
		registry  brewreg.Registry
	}
}

// DefaultConcurrency is the default maximum threads for parallel tasks
const DefaultConcurrency int = 8

// NewHops creates a new Tool with default values
func NewHops(version string) *Hops {
	return &Hops{
		version:     version,
		ConfigFiles: hopsv1.ConfigurationSearchFiles,
		EnvFiles:    brewenv.DefaultEnvironmentFiles(),
		Concurrency: DefaultConcurrency,
	}
}

// Version returns the version (overwritten by main.version if needed)
func (action *Hops) Version() string {
	return action.version
}

// MaxGoroutines produces the maximum number of
// Goroutines that should be started at a time
func (action *Hops) MaxGoroutines() int {
	if action.Concurrency < 1 {
		return 1
	}
	return action.Concurrency
}

// Homebrew produces the default homebrew client
func (action *Hops) Homebrew() *brewenv.Environment {
	if action.brewcfg == nil {
		var err error

		// Load env files
		for _, envfile := range action.EnvFiles {
			err = godotenv.Load(envfile)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				slog.Warn("loading Homebrew environment files", o.ErrAttr(err))
			}
		}

		// Initialize the Homebrew object
		action.brewcfg, err = brewenv.Load()
		if err != nil {
			slog.Debug("loading environment config", o.ErrAttr(err))
		}

		// Override the loaded values
		for _, override := range action.brewOverrides {
			override(action.brewcfg)
		}
	}
	return action.brewcfg
}

// Homebrew produces the default homebrew client
func (action *Hops) Prefix() prefix.Prefix {
	return prefix.Prefix(action.Homebrew().Prefix)
}

// UserAgent produces the tool's user agent string
func (action *Hops) UserAgent() string {
	return "hops/" + action.version
}

// AddHomebrewOverride adds a configuration override function
// The override function will be called when loading
// homebrew's environment configuration
func (action *Hops) AddHomebrewOverride(overrides ...func(e *brewenv.Environment)) {
	action.brewOverrides = append(action.brewOverrides, overrides...)
}

// AddConfigOverride adds a configuration override function
// The override function will be called when loading
// hops' configuration
func (action *Hops) AddConfigOverride(overrides ...func(cfg *hopsv1.Configuration)) {
	action.configOverrides = append(action.configOverrides, overrides...)
}

// Returns the client used for authentication to OCI repositories
func (action *Hops) AuthClient() *auth.Client {
	if action.authClient == nil {
		// prepare authentication using Docker credentials
		credStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
		if err != nil {
			panic(err)
		}

		action.authClient = &auth.Client{
			Client:     retry.NewClient(),
			Cache:      auth.NewCache(),
			Credential: credentials.Credential(credStore), // Use the credentials store
		}
		action.authClient.SetUserAgent(action.UserAgent())

		slog.Debug("using docker config", slog.String("file", credStore.ConfigPath()))
	}

	return action.authClient
}

// Returns the auth headers for HTTP requests
func (action *Hops) authHeaders() http.Header {
	header := http.Header{}
	// Add the GitHub Packages auth header from Homebrew config
	header.Add("Authorization",
		action.Homebrew().GitHubPackagesAuth())
	return header
}

// Config returns the Hops CLI configuration
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
			slog.Error("loading config file", slog.String("path", filename), o.ErrAttr(err))
			continue
		}

		slog.Debug("loaded config file", slog.String("path", filename))
		break
	}

	// Set default values for the configuration here
	hopsv1.ConfigurationDefault(action.cfg)

	return action.cfg
}

var errNoRegistryConfig = errors.New("no registry configured")

// registry produces the Hops registry from options
func (action *Hops) registry() (hopsreg.Registry, error) {
	return hopsRegistry(action.AuthClient(), &action.Config().Registry)
}

func hopsRegistry(authClient *auth.Client, cfg *hopsv1.RegistryConfig) (hopsreg.Registry, error) {
	switch {
	case cfg.Prefix == "":
		return nil, errNoRegistryConfig
	case cfg.OCILayout:
		return hopsreg.NewLocal(cfg.Prefix), nil
	default:
		return hopsreg.NewRegistry(
			cfg.Prefix,
			authClient,
			cfg.PlainHTTP,
		)
	}
}

func hopsClient(cache string, alternateTags map[string]string, maxGoroutines int, reg hopsreg.Registry) hops.Client {
	// Create OCI layout cache
	btlcache := hopsreg.NewLocal(cache)

	// Initialize client
	return hops.NewClient(
		reg, btlcache,
		alternateTags, maxGoroutines)
}

// SetAlternateVersions sets alternate tags from a list of arguments,
// and returns the isolated names from the arguments
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

// Formulary produces the configured Formulary
func (action *Hops) Formulary(ctx context.Context) (formula.Formulary, error) {
	switch action.Config().Registry.Prefix {
	// Homebrew-style Formulary
	case "":
		return action.brewFormulary(ctx, false)
	// Hops-style Formulary
	default:
		return action.hopsClient()
	}
}

// BottleRegistry produces the configured Bottle registry
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

// hopsClient initializes the configured formula.Formulary/bottle.Registry
func (action *Hops) hopsClient() (hops.Client, error) {
	if action.hopsclient == nil {
		cache := filepath.Join(action.Config().Cache, "oci")
		slog.Debug("using Hops client", slog.String("registry", action.Config().Registry.Prefix), slog.String("cache", cache))

		// Initialize registry.Registry
		reg, err := action.registry()
		if err != nil {
			return nil, err
		}

		action.hopsclient = hopsClient(
			cache,
			action.alternateTags,
			action.MaxGoroutines(),
			reg)
	}
	return action.hopsclient, nil
}

// brewFormulary initializes the configured formula.Formulary
func (action *Hops) autoUpdate(ctx context.Context) error {
	if action.Config().Registry.Prefix != "" {
		return nil
	}
	_, err := action.brewFormulary(ctx, true)
	return err
}

// brewFormulary initializes the configured formula.Formulary
func (action *Hops) brewFormulary(ctx context.Context, autoUpdate bool) (brewformulary.PreloadedFormulary, error) {
	if action.brewclient.formulary == nil {
		slog.Debug("using Homebrew API formulary", slog.String("HOMEBREW_API_DOMAIN", action.Config().Homebrew.Domain)) //nolint:sloglint

		// Set auto-update config
		var upcfg *hopsv1.AutoUpdateConfig
		if autoUpdate {
			upcfg = &action.Config().Homebrew.AutoUpdate
		}

		// Load the index
		index, err := brewformulary.FetchV1(ctx,
			brewapi.NewClient(action.Config().Homebrew.Domain),
			action.Config().Cache, upcfg)
		if err != nil {
			return nil, err
		}

		action.brewclient.formulary = index
	}
	return action.brewclient.formulary, nil
}

// brewRegistry initializes the configured bottle.Registry
func (action *Hops) brewRegistry() brewreg.Registry {
	if action.brewclient.registry == nil {
		slog.Debug("using Homebrew registry", slog.String("HOMEBREW_ARTIFACT_DOMAIN", action.Homebrew().ArtifactDomain)) //nolint:sloglint
		action.brewclient.registry = brewreg.NewBottleRegistry(
			action.authHeaders(),
			retry.DefaultClient,
			action.Homebrew().Cache,
			action.MaxGoroutines(),
			action.Homebrew().ArtifactDomain,
		)
	}
	return action.brewclient.registry
}

// func parseArgs(args []string) (names, versions []string) {
// 	names = make([]string, len(args))
// 	versions = make([]string, len(args))
// 	for i, arg := range args {
// 		names[i], versions[i] = parseArg(arg)
// 	}
// 	return names, versions
// }

func parseArg(arg string) (name, version string) {
	fields := strings.SplitN(arg, "=", 2)
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
