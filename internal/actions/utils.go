package actions

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	brewreg "github.com/act3-ai/hops/internal/brew/registry"
	hops "github.com/act3-ai/hops/internal/hops"
	hopsreg "github.com/act3-ai/hops/internal/hops/registry"
)

func hopsRegistry(cfg *hopsv1.RegistryConfig, userAgent string) (hopsreg.Registry, error) {
	switch {
	case cfg.Prefix == "":
		return nil, errors.New("no registry configured")
	case cfg.OCILayout:
		slog.Debug("using oci-layout dirs as bottle registry", slog.String("dir", cfg.Prefix))

		return hopsreg.NewLocal(cfg.Prefix), nil
	default:
		// pass attr to authClient
		log := slog.Default().With(slog.String("registry", cfg.Prefix))
		log.Debug("using remote bottle registry")

		client, err := authClient(log, cfg, userAgent)
		if err != nil {
			return nil, err
		}
		reg, err := hopsreg.NewRemote(cfg.Prefix, client, cfg.PlainHTTP)
		if err != nil {
			return nil, err
		}
		return reg, nil
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

// reference: https://github.com/oras-project/oras/blob/main/cmd/oras/internal/option/remote.go#L234
func regTLS(cfg *hopsv1.RegistryConfig) *tls.Config {
	config := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}
	return config
}

// authClient produces an auth client.
//
// reference: https://github.com/oras-project/oras/blob/main/cmd/oras/internal/option/remote.go#L256
func authClient(log *slog.Logger, cfg *hopsv1.RegistryConfig, userAgent string) (*auth.Client, error) {
	var credStore credentials.Store
	switch {
	// Use specified config file
	case cfg.Config != "":
		fileStore, err := credentials.NewStore(cfg.Config, credentials.StoreOptions{})
		if err != nil {
			return nil, fmt.Errorf("loading registry config: %w", err)
		}

		log.Debug("using registry config", slog.String("file", fileStore.ConfigPath()))
		credStore = fileStore
	// Use Docker credentials
	default:
		// prepare authentication using Docker credentials
		dockerStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
		if err != nil {
			return nil, fmt.Errorf("loading docker config: %w", err)
		}

		log.Debug("using docker config", slog.String("file", dockerStore.ConfigPath()))
		credStore = dockerStore
	}

	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	baseTransport.TLSClientConfig = regTLS(cfg) // parse TLS options

	h, err := cfg.ParseHeaders()
	if err != nil {
		return nil, err
	}

	client := &auth.Client{
		Client: &http.Client{
			// http.RoundTripper with a retry using the DefaultPolicy
			// see: https://pkg.go.dev/oras.land/oras-go/v2/registry/remote/retry#Policy
			Transport: retry.NewTransport(baseTransport),
		},
		Cache:      auth.NewCache(),
		Header:     h,
		Credential: credentials.Credential(credStore), // Use the credentials store
	}

	client.SetUserAgent(userAgent)

	return client, nil
}

func brewRegistry(log *slog.Logger, cfg *brewenv.Configuration, maxGoroutines int) brewreg.Registry {
	log.Debug("using Homebrew registry", //nolint:sloglint
		slog.String("HOMEBREW_BOTTLE_DOMAIN", cfg.BottleDomain),
		slog.String("HOMEBREW_ARTIFACT_DOMAIN", cfg.ArtifactDomain))

	return brewreg.NewBottleRegistry(
		cfg.GitHubPackagesHeaders(),
		retry.NewClient(),
		cfg.Cache,
		maxGoroutines,
		cfg.BottleDomain,
		cfg.ArtifactDomain,
	)
}
