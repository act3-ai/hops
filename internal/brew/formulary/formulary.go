package brewformulary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	brewenv "github.com/act3-ai/hops/internal/apis/config.brew.sh"
	api "github.com/act3-ai/hops/internal/apis/formulae.brew.sh"
	brewv1 "github.com/act3-ai/hops/internal/apis/formulae.brew.sh/v1"
	brewapi "github.com/act3-ai/hops/internal/brew/api"
)

// FetchV1 fetches the v1 index either from the cache or from the API according to its existence and the auto-update configuration.
func FetchV1(ctx context.Context, apiclient *brewapi.Client, dir string, autoUpdate *brewenv.AutoUpdateConfig) (*V1Cache, error) {
	_, err := os.Stat(formulaeFile(dir))
	switch {
	// File does not exist or is unreadable
	case err != nil:
		return fetchV1(ctx, apiclient, dir)
	// File exists and auto-update is disabled
	case autoUpdate == nil:
		return LoadV1(dir)
	// File exists but requires updating
	case autoUpdate.ShouldAutoUpdate(formulaeFile(dir)):
		return fetchV1(ctx, apiclient, dir)
	// File exists and does not need updated
	default:
		return LoadV1(dir)
	}
}

func formulaeFile(dir string) string {
	return filepath.Join(dir, "api", "formula.json")
}

func namesFile(dir string) string {
	return filepath.Join(dir, "api", api.CachedFormulaNamesFile)
}

func aliasesFile(dir string) string {
	return filepath.Join(dir, "api", api.CachedFormulaAliasesFile)
}

// readWriteJSON reads from r while writing to a file at path and simultaneously decoding JSON into type T.
func readWriteJSON[T any](path string, r io.Reader) (*T, error) {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return nil, fmt.Errorf("creating dir: %w", err)
	}

	// Create the index file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	// Create decoder that reads from a TeeReader
	// The TeeReader writes to the file as it reads from the given reader
	decoder := json.NewDecoder(io.TeeReader(r, f))
	// decoder.DisallowUnknownFields()

	obj := new(T)
	if err := decoder.Decode(obj); err != nil {
		return nil, fmt.Errorf("decoding JSON failed: %w", err)
	}

	return obj, nil
}

// LoadV1 loads the v1 index from a cache directory.
func LoadV1(dir string) (*V1Cache, error) {
	file := formulaeFile(dir)
	// Check for existing index file
	f, err := os.Open(file)
	switch {
	// No cached file
	case errors.Is(err, os.ErrNotExist):
		return nil, errors.New("loading index from cache: index is not cached")
	// Unreadable file
	case err != nil:
		return nil, fmt.Errorf("reading cached file %s: %w", file, err)
	// Readable cached file
	default:
		defer f.Close()

		data := brewv1.Index{}

		jd := json.NewDecoder(f)

		err := jd.Decode(&data)
		if err != nil {
			return nil, fmt.Errorf("parsing cached file: %w", err)
		}

		return cacheV1(data), nil
	}
}

func fetchV1(ctx context.Context, apiclient *brewapi.Client, dir string) (*V1Cache, error) {
	// Fetch full v1 list
	r, err := apiclient.FetchFormulaeV1(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Parse JSON and cache the response
	data, err := readWriteJSON[[]*brewv1.Info](formulaeFile(dir), r)
	if err != nil {
		return nil, err
	}

	// Load in cached form
	index := cacheV1(*data)

	// Update cache
	err = writeAPICache(index, dir)
	if err != nil {
		return index, err
	}

	return index, nil
}
