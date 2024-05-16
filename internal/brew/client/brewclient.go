package brewclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	hopsv1 "github.com/act3-ai/hops/internal/apis/config.hops.io/v1beta1"
	api "github.com/act3-ai/hops/internal/apis/formulae.brew.sh"
	v1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	"github.com/act3-ai/hops/internal/utils/resputil"
)

// Client represents the Homebrew API index
type Client struct {
	*http.Client        // for API requests
	domain       string // HOMEBREW_API_DOMAIN
	file         string // cached location of the index
	namefile     string // cached location of the names
	aliasfile    string // cached location of formula aliases
	*APIIndex           // uses the default implementations
}

// New creates a new Index for a Homebrew API source
func New(client *http.Client, cache, apiDomain string) *Client {
	return &Client{
		Client:    client,
		domain:    apiDomain,
		file:      filepath.Join(cache, "api", "formula.json"),
		namefile:  filepath.Join(cache, "api", api.CachedFormulaNamesFile),
		aliasfile: filepath.Join(cache, "api", api.CachedFormulaAliasesFile),
	}
}

// Load implements Loader
func (hi *Client) Load(ctx context.Context) error {
	parseIndex := func(r io.Reader) error {
		data := v1.Index{}

		jd := json.NewDecoder(r)
		jd.DisallowUnknownFields()

		err := jd.Decode(&data)
		// err := jd.(b, &data)
		if err != nil {
			return fmt.Errorf("parsing formula index: %w", err)
		}

		hi.APIIndex = NewAPIIndex(data)

		return nil
	}

	// Check for existing index file
	f, err := os.Open(hi.file)
	if err == nil {
		// If file exists, parse it and return existing file
		return parseIndex(f)
	} else if !errors.Is(err, os.ErrNotExist) {
		// If file was unreadable, return the error
		return fmt.Errorf("checking index file %s: %w", hi.file, err)
	}

	// Create parent directory
	err = os.MkdirAll(filepath.Dir(hi.file), 0o775)
	if err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	// Open the symlink path, creating the bottle download file
	download, err := os.OpenFile(hi.file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("creating formula index: %w", err)
	}
	defer download.Close()

	// Fetch the index
	r, err := hi.fetchAll(ctx)
	if err != nil {
		return fmt.Errorf("downloading formula index: %w", err)
	}

	// Read the response, use tee reader to write to cache file while decoding
	err = parseIndex(io.TeeReader(r, download))
	if err != nil {
		return err
	}

	// Write names and aliases to basic text files for lighter reads
	err = hi.cacheFormulaNames()
	if err != nil {
		return fmt.Errorf("caching index: %w", err)
	}

	return nil
}

// fetchAll
func (hi *Client) fetchAll(ctx context.Context) (io.ReadCloser, error) {
	// https://formulae.brew.sh/docs/api/#list-metadata-for-all-homebrewcore-formulae-or-homebrewcask-casks
	// GET https://formulae.brew.sh/api/formula.json
	slog.Info("Downloading " + hi.domain + "/formula.json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		hi.domain+"/formula.json",
		nil)
	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	resp, err := hi.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for a non-success status and handle
	if !resputil.HTTPSuccess(resp) {
		return resp.Body, resputil.HandleHTTPError(resp)
	}

	return resp.Body, nil
}

// CachedLocation produces the cached location of the index file
// The function does not check if the path actually exists or not
func (hi *Client) Source() string {
	return hi.domain
}

// IsCached reports whether a cached version of the index exists
func (hi *Client) IsCached() bool {
	_, err1 := os.Stat(hi.file)
	_, err2 := os.Stat(hi.namefile)
	_, err3 := os.Stat(hi.aliasfile)
	return errors.Join(err1, err2, err3) == nil
}

// ShouldAutoUpdate reports whether an auto update should be performed or not
func (hi *Client) ShouldAutoUpdate(opts *hopsv1.AutoUpdateConfig) bool {
	return opts.ShouldAutoUpdate(hi.file)
}

// Reset resets the cached index
func (hi *Client) Reset(_ *hopsv1.AutoUpdateConfig) error {
	// Remove cached index file
	// The load functions only download a fresh index if there is no existing index
	if err := os.RemoveAll(hi.file); err != nil {
		return fmt.Errorf("removing existing index: %w", err)
	}
	return nil
}

// cacheFormulaNames caches the formula names in a separate file
func (hi *Client) cacheFormulaNames() error {
	if hi.APIIndex == nil {
		slog.Debug("skipped caching formula names, empty formula index")
		return nil
	}

	// Create parent directory
	err := os.MkdirAll(filepath.Dir(hi.namefile), 0o775)
	if err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	names := hi.APIIndex.ListNames()
	err = api.WriteFormulaNames(names, hi.namefile)
	if err != nil {
		return err
	}

	aliases := hi.APIIndex.Aliases()
	err = api.WriteFormulaAliases(aliases, hi.aliasfile)
	if err != nil {
		return err
	}

	return nil
}
