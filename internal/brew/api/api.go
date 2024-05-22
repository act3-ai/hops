package brewapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/act3-ai/hops/internal/utils/resputil"
)

const (
	v1formulae = "formula.json"
	v2formulae = "formula.jws.json"
	v2casks    = "cask.jws.json"
	v3core     = "internal/v3/homebrew-core.jws.json"
)

// type NameFetcher[T any] interface {
// 	Fetch(name string) (T, error)
// }

// type AllFetcher[T any] interface {
// 	FetchAll(name string) (T, error)
// }

// type NameLister interface {
// 	Names() ([]string, error)
// }

// type NameFetcherLister[T any] interface {
// 	NameFetcher[T]
// 	NameLister
// }

// type V1 interface {
// 	Formulae() NameFetcher[*brewv1.Info]
// }

// type V2 interface {
// 	AllFetcher[*brewv2.Tap]
// 	Formulae() AllFetcher[[]*brewv2.Formula]
// 	Casks() AllFetcher[[]*v2.Cask]
// }

// type V3 interface {
// 	Tap(tap string) AllFetcher[*brewv3.Tap]
// }

// type V3Formulae interface {
// 	NameFetcher[*brewv3.Formula]
// }

// Client makes API requests
type Client struct {
	HTTPClient *http.Client
	APIDomain  string
}

// NewClient initializes a default client
func NewClient(apiDomain string) *Client {
	return &Client{
		HTTPClient: retry.NewClient(),
		APIDomain:  apiDomain,
	}
}

// FetchFormulaV1 fetches the v1 API data for the named formula
// If name does not exist, this function will fail
func (client Client) FetchFormulaV1(ctx context.Context, name string) (io.ReadCloser, error) {
	return client.fetch(ctx, "formula/"+url.PathEscape(name)+".json")
}

// FetchFormulaeV1 fetches the v1 API data for all existing formulae
func (client Client) FetchFormulaeV1(ctx context.Context) (io.ReadCloser, error) {
	// https://formulae.brew.sh/docs/api/#list-metadata-for-all-homebrewcore-formulae-or-homebrewcask-casks
	// GET https://formulae.brew.sh/api/formula.json
	return client.fetch(ctx, v1formulae)
}

// FetchFormulaeV2 fetches the v2 API data for all existing formulae
func (client Client) FetchFormulaeV2(ctx context.Context) (io.ReadCloser, error) {
	return client.fetch(ctx, v2formulae)
}

// FetchCasksV2 fetches the v2 API data for all existing casks
func (client Client) FetchCasksV2(ctx context.Context) (io.ReadCloser, error) {
	return client.fetch(ctx, v2casks)
}

// FetchCoreV3 fetches the v3 API data for the homebrew/core tap
func (client Client) FetchCoreV3(ctx context.Context) (io.ReadCloser, error) {
	return client.fetch(ctx, v3core)
}

func (client Client) fetch(ctx context.Context, endpoint string) (io.ReadCloser, error) {
	target := client.APIDomain + "/" + endpoint

	slog.Debug("Fetching", slog.String("target", target))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		target,
		nil)
	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for a non-success status and handle
	if !resputil.HTTPSuccess(resp) {
		return resp.Body, resputil.HandleHTTPError(resp)
	}

	return resp.Body, nil
}
