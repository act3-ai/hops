package brewreg

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sourcegraph/conc/iter"

	brewfmt "github.com/act3-ai/hops/internal/brew/fmt"
	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/formula/bottle"
	"github.com/act3-ai/hops/internal/utils/logutil"
	"github.com/act3-ai/hops/internal/utils/resputil"
	"github.com/act3-ai/hops/internal/utils/symlink"
)

// Registry defines the capabilities of Homebrew's registry usage.
type Registry interface {
	bottle.ConcurrentRegistry
}

// registry downloads bottles with an HTTP client.
//
// URLs are constructed as:
//
//	[ARTIFACT_DOMAIN/]BOTTLE_ROOT_URL/BOTTLE_NAME:BOTTLE_VERSION@sha256:BOTTLE_SHA256
//
// With HOMEBREW_BOTTLE_DOMAIN:
//
//	[ARTIFACT_DOMAIN/]BOTTLE_DOMAIN/BOTTLE_NAME:BOTTLE_VERSION@sha256:BOTTLE_SHA256
type registry struct {
	headers        http.Header  // for GitHub Packages auth
	HTTP           *http.Client // client for HTTP requests
	cache          string       // cache directory
	maxGoroutines  int
	bottleDomain   string
	artifactDomain string
}

// NewBottleRegistry creates a new BottleRegistry.
func NewBottleRegistry(
	headers http.Header,
	client *http.Client,
	cache string,
	maxGoroutines int,
	bottleDomain,
	artifactDomain string,
) Registry {
	return newRegistry(headers, client, cache, maxGoroutines, bottleDomain, artifactDomain)
}

// newRegistry creates a new registry.
func newRegistry(headers http.Header, client *http.Client, cache string, maxGoroutines int, bottleDomain, artifactDomain string) *registry {
	return &registry{
		headers:        headers,
		HTTP:           client,
		cache:          cache,
		maxGoroutines:  maxGoroutines,
		bottleDomain:   bottleDomain,
		artifactDomain: artifactDomain,
	}
}

// Source provides the source for a bottle.
func (store *registry) Source(f formula.PlatformFormula) (string, error) {
	root, path := bottleURL(f)
	if path == "" { // specifies no bottle
		return "", nil
	}
	if store.bottleDomain != "" {
		// Replace default bottle domain with configured bottle domain
		root = store.bottleDomain
	}

	if store.artifactDomain != "" {
		srcURL, err := url.Parse(root + path)
		if err != nil {
			return "", fmt.Errorf("parsing URL to replace domain with HOMEBREW_ARTIFACT_DOMAIN: %w", err)
		}
		// scheme:opaque?query#fragment
		// scheme://userinfo@host/path?query#fragment
		// This join does not support query or fragment additions
		return store.artifactDomain + "/" + srcURL.Path, nil
	}
	return root + path, nil
}

// bottleURL produces the URL for a bottle.
func bottleURL(f formula.PlatformFormula) (root, path string) {
	btl := f.Bottle()
	if btl == nil {
		return "", ""
	}
	path = "/" + brewfmt.Repo(f.Name()) + "/blobs/sha256:" + btl.Sha256
	return btl.RootURL, path
}

// FetchBottle implements formula.BottleRegistry.
func (store *registry) FetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	return store.fetchBottle(ctx, f)
}

// FetchBottles implements formula.ConcurrentBottleRegistry.
func (store *registry) FetchBottles(ctx context.Context, formulae []formula.PlatformFormula) ([]io.ReadCloser, error) {
	fetchers := iter.Mapper[formula.PlatformFormula, io.ReadCloser]{MaxGoroutines: store.maxGoroutines}
	return fetchers.MapErr(formulae, func(fp *formula.PlatformFormula) (io.ReadCloser, error) {
		return store.fetchBottle(ctx, *fp)
	})
}

// fetchBottle implements formula.BottleRegistry.
func (store *registry) fetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	// Download bottle file
	path, err := store.download(ctx, f)
	if err != nil {
		return nil, err
	}

	// Open downloaded bottle file
	return os.Open(path)
}

// LinkName returns the name of the symlink to the downloaded bottle .tar.gz file for the formula.
//
// Pattern:
//
//	NAME--VERSION
//
// Example:
//
//	cowsay--3.04_1
func linkName(f formula.Formula) string {
	return f.Name() + "--" + formula.PkgVersion(f)
}

// lookupCachedFile.
func lookupCachedFile(file, link string) (*os.File, error) {
	// Create parent directories (also will create the cache directory if it does not exist)
	err := os.MkdirAll(filepath.Dir(file), 0o775)
	if err != nil {
		return nil, fmt.Errorf("creating download file: %w", err)
	}

	// Create the symlink into the downloads directory
	err = symlink.Relative(file, link, &symlink.Options{Overwrite: true})
	if err != nil {
		return nil, fmt.Errorf("creating cache symlink: %w", err)
	}

	_, err = os.Stat(link)
	if err == nil {
		// Return here if the file is already downloaded
		return nil, nil
	}

	// For any unreadable files, remove the files and redownload
	if !errors.Is(err, os.ErrNotExist) {
		// Implies an unreadable file in the cache
		// Remove files and redownload
		slog.Warn("Removing unreadable cache file", logutil.ErrAttr(err))

		err = os.RemoveAll(file)
		if err != nil {
			return nil, fmt.Errorf("removing unreadable file in cache: %w", err)
		}
	}

	// Open the symlink path, creating the cache file
	bottleFile, err := os.Create(link)
	if err != nil {
		return nil, fmt.Errorf("creating download file: %w", err)
	}

	return bottleFile, nil
}

// Download downloads a bottle.
func (store *registry) download(ctx context.Context, f formula.PlatformFormula) (string, error) {
	btl := f.Bottle()
	if btl == nil {
		return "", fmt.Errorf("no bottle provided for Formula %s", f.Name())
	}

	source, err := store.Source(f)
	if err != nil {
		return "", err
	}

	bottleFileName := formula.BottleFileName(f)

	// 5c7f66c74fe4c17116808bfac4c2729c32062dbec291e2a897d267567c790ea4--cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
	urlsum := sha256.Sum256([]byte(source))
	file := filepath.Join(store.cache, "downloads", fmt.Sprintf("%x--%s", urlsum, bottleFileName))

	// cowsay--3.04_1
	link := filepath.Join(store.cache, linkName(f))

	bottleFile, err := lookupCachedFile(file, link)
	if err == nil && bottleFile == nil {
		// Return here if the file is already downloaded
		// this should also validate the existing file's checksum
		slog.Debug("Already downloaded: " + bottleFileName)
		return "", nil
	} else if err != nil {
		return "", err
	}
	defer bottleFile.Close()

	u, err := url.Parse(source)
	if err != nil {
		return "", fmt.Errorf("[%s] parsing bottle source: %w", f.Name(), err)
	}

	slog.Debug("starting bottle download",
		slog.String("url", source),
		slog.String("ln", link),
		slog.String("path", file))

	deleteOnFailure := func() error {
		return errors.Join(
			os.RemoveAll(file),
			os.RemoveAll(link),
		)
	}

	switch u.Scheme {
	case "https", "http":
		slog.Debug("Downloading bottle")
		err = downloadBottleHTTP(ctx, *store.HTTP, store.headers, source, bottleFile)
		if err != nil {
			return "", errors.Join(
				fmt.Errorf("[%s] downloading bottle: %w", f.Name(), err),
				deleteOnFailure(),
			)
		}
	// case "oci":
	// 	err := downloadBottleOCI(ctx, store.OCI, strings.TrimPrefix(source, "oci://"), bottleFile)
	// 	if err != nil {
	// 		return fmt.Errorf("[%s] downloading bottle: %w", f.Name(), err)
	// 	}
	default:
		return "", errors.Join(
			fmt.Errorf("[%s] downloading bottle: unsupported URL scheme %q", f.Name(), u.Scheme),
			deleteOnFailure(),
		)
	}

	slog.Debug("Downloaded " + bottleFileName)

	return link, nil
}

/*
// downloadBottleOCI downloads a bottle using the oras client
func downloadBottleOCI(ctx context.Context, client remote.Client, ref string, w io.Writer) error {
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return err
	}
	repo.Client = client

	bd, err := repo.Blobs().Resolve(ctx, ref)
	if err != nil {
		return fmt.Errorf("downloading bottle: could not resolve blob: %w", err)
	}

	slog.Debug("downloading bottle blob", slog.String("digest", bd.Digest.String()))
	r, err := repo.Blobs().Fetch(ctx, bd)
	if err != nil {
		return fmt.Errorf("oras blob fetch failed: %w", err)
	}
	defer r.Close()

	// Use VerifyReader to verify the digest as the blob is read
	vr := content.NewVerifyReader(r, bd)

	_, err = io.Copy(w, vr)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	err = vr.Verify()
	if err != nil {
		return fmt.Errorf("digest mismatch: %w", err)
	}

	return nil
}
*/

// downloadBottleHTTP downloads a bottle using the given.
func downloadBottleHTTP(ctx context.Context, client http.Client, header http.Header, ref string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	if err != nil {
		return fmt.Errorf("preparing request: %w", err)
	}

	// Set the headers
	req.Header = header

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-success status and handle
	if !resputil.HTTPSuccess(resp) {
		return resputil.HandleHTTPError(resp)
	}

	// Copy the response body to the provided writer
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
