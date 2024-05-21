package brewformulary

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

	"github.com/act3-ai/hops/internal/formula"
	"github.com/act3-ai/hops/internal/o"
	"github.com/act3-ai/hops/internal/utils/resputil"
	"github.com/act3-ai/hops/internal/utils/symlink"
)

// BottleStore downloads bottles with an HTTP client.
//
// URLs are constructed as:
//
// [ARTIFACT_DOMAIN/]BOTTLE_ROOT_URL/BOTTLE_NAME:BOTTLE_VERSION@sha256:BOTTLE_SHA256
//
//   - ARTIFACT_DOMAIN: given as argument to NewHTTPStore
//   - BOTTLE_ROOT_URL: f.Bottle["stable"].RootURL
//   - BOTTLE_NAME: f.Name
//   - BOTTLE_VERSION: f.Version()
//
// TODO: Does not support HOMEBREW_ARTIFACT_DOMAIN
type BottleStore struct {
	headers http.Header  // for GitHub Packages auth
	HTTP    *http.Client // client for HTTP requests
	// OCI           remote.Client // client for OCI requests
	cache          string // cache directory
	maxGoroutines  int
	artifactDomain string
}

// NewBottleRegistry creates a new BottleRegistry
func NewBottleRegistry(headers http.Header, client *http.Client, cache string, maxGoroutines int, artifactDomain string) formula.ConcurrentBottleRegistry {
	return NewBottleStore(headers, client, cache, maxGoroutines, artifactDomain)
}

// NewBottleStore creates a new store
func NewBottleStore(headers http.Header, client *http.Client, cache string, maxGoroutines int, artifactDomain string) *BottleStore {
	return &BottleStore{
		headers: headers,
		HTTP:    client,
		// OCI:     regClient,
		cache:          cache,
		maxGoroutines:  maxGoroutines,
		artifactDomain: artifactDomain,
	}
}

// Source provides the source for a bottle
func (store *BottleStore) Source(f formula.PlatformFormula) string {
	src := bottleURL(f)
	if src == "" {
		return ""
	}
	return store.artifactDomain + "/" + src
}

// bottleURL produces the URL for a bottle
func bottleURL(f formula.PlatformFormula) string {
	btl := f.Bottle()
	if btl == nil {
		return ""
	}

	// replace default root URL with configured root
	src := btl.RootURL + "/blobs/sha256:" + btl.Sha256
	return src
}

// FetchBottle implements formula.BottleRegistry.
func (store *BottleStore) FetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	return store.fetchBottle(ctx, f)
}

// FetchBottles implements formula.ConcurrentBottleRegistry.
func (store *BottleStore) FetchBottles(ctx context.Context, formulae []formula.PlatformFormula) ([]io.ReadCloser, error) {
	fetchers := iter.Mapper[formula.PlatformFormula, io.ReadCloser]{MaxGoroutines: store.maxGoroutines}
	return fetchers.MapErr(formulae, func(fp *formula.PlatformFormula) (io.ReadCloser, error) {
		return store.fetchBottle(ctx, *fp)
	})
}

// fetchBottle implements formula.BottleRegistry.
func (store *BottleStore) fetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error) {
	// Download bottle file
	path, err := store.download(ctx, f)
	if err != nil {
		return nil, err
	}
	// Open downloaded bottle file
	return os.Open(path)
}

// LinkName returns the name of the symlink to the downloaded bottle .tar.gz file for the formula
//
// Pattern: NAME--VERSION
//
// Example: cowsay--3.04_1
func linkName(f formula.Formula) string {
	return f.Name() + "--" + formula.PkgVersion(f.Version())
}

// func (store *BottleStore) exists()

// lookupCachedFile
func (store *BottleStore) lookupCachedFile(file, link string) (*os.File, error) {
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
		slog.Warn("Removing unreadable cache file", o.ErrAttr(err))

		err = errors.Join(os.RemoveAll(file))
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

// Download downloads a bottle
func (store *BottleStore) download(ctx context.Context, f formula.PlatformFormula) (string, error) {
	btl := f.Bottle()
	if btl == nil {
		return "", fmt.Errorf("no bottle provided for Formula %s", f.Name())
	}

	source := store.Source(f)

	bottleFileName := formula.BottleFileName(f)

	// 5c7f66c74fe4c17116808bfac4c2729c32062dbec291e2a897d267567c790ea4--cowsay--3.04_1.arm64_sonoma.bottle.tar.gz
	urlsum := sha256.Sum256([]byte(source))
	file := filepath.Join(store.cache, "downloads", fmt.Sprintf("%x--%s", urlsum, bottleFileName))

	// cowsay--3.04_1
	link := filepath.Join(store.cache, linkName(f))

	bottleFile, err := store.lookupCachedFile(file, link)
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

	slog.Debug("starting bottle download", slog.String("ln", link), slog.String("path", file))

	switch u.Scheme {
	case "https", "http":
		err = downloadBottleHTTP(ctx, *store.HTTP, store.headers, source, bottleFile)
		if err != nil {
			return "", fmt.Errorf("[%s] downloading bottle: %w", f.Name(), err)
		}
	// case "oci":
	// 	err := downloadBottleOCI(ctx, store.OCI, strings.TrimPrefix(source, "oci://"), bottleFile)
	// 	if err != nil {
	// 		return fmt.Errorf("[%s] downloading bottle: %w", f.Name(), err)
	// 	}
	default:
		return "", fmt.Errorf("[%s] downloading bottle: unsupported URL scheme %q", f.Name(), u.Scheme)
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

// downloadBottleHTTP downloads a bottle using the given
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
