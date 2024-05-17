package actions

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/bottle"
	"github.com/act3-ai/hops/internal/brew"
	brewclient "github.com/act3-ai/hops/internal/brew/client"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
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

// Index returns the index
func (action *Hops) Index() formula.CachedIndex {
	return brewclient.New(
		http.DefaultClient,
		action.Homebrew().Cache,
		action.Config().Homebrew.Domain)
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
			Client:     retry.DefaultClient,
			Cache:      auth.NewCache(),
			Credential: credentials.Credential(credStore), // Use the credentials store
		}
		action.authClient.SetUserAgent(action.UserAgent())

		slog.Debug("using docker config", slog.String("file", credStore.ConfigPath()))
	}

	return action.authClient
}

// Returns the auth headers for HTTP requests
func (action *Hops) AuthHeaders() http.Header {
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

	slog.Debug("using config", slog.String("config", action.cfg.String()))

	return action.cfg
}

// FetchAll finds all formula from the index, using a concurrent iterator if the list of names is large
func (action *Hops) FetchAll(log func(string), index formula.Index, names ...string) ([]*v1.Info, error) {
	find := func(name string) (*v1.Info, error) {
		log("Fetching " + o.StyleGreen(name))
		f := index.Find(name)
		if f == nil {
			return nil, brew.NewErrFormulaNotFound(name)
		}
		return f, nil
	}

	formulae := make([]*v1.Info, len(names))
	var err error
	for i, n := range names {
		formulae[i], err = find(n)
		if err != nil {
			return formulae, err
		}
	}

	return formulae, nil
}

var errNoRegistryConfig = errors.New("no registry configured")

// Registry produces the Hops registry from options
func (action *Hops) Registry() (bottle.Registry, error) {
	switch {
	case action.Config().Registry.Prefix == "":
		return nil, errNoRegistryConfig
	case action.Config().Registry.OCILayout:
		return bottle.NewLocal(action.Config().Registry.Prefix), nil
	default:
		return bottle.NewRegistry(
			action.Config().Registry.Prefix,
			action.AuthClient(),
			action.Config().Registry.PlainHTTP,
		)
	}
}
