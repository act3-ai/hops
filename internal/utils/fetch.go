package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
)

// FetchRead abstracts the fetch/read/verify flow
func FetchRead(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor, action func(r io.Reader) error) error {
	rc, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer rc.Close()

	vr := content.NewVerifyReader(rc, desc)

	if err := action(vr); err != nil {
		return err
	}

	return vr.Verify()
}

// FetchAll safely fetches the content described by the descriptor.
// The fetched content is verified against the size and the digest.
func FetchDecode[T any](ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) (*T, error) {
	obj := new(T)
	err := FetchRead(ctx, fetcher, desc, func(r io.Reader) error {
		decoder := json.NewDecoder(r)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(obj); err != nil {
			return fmt.Errorf("decoding JSON failed: %w", err)
		}

		return nil
	})
	return obj, err
}

// ReadAll safely reads the content described by the descriptor.
// The read content is verified against the size and the digest
// using a VerifyReader.
func ReadAll(r io.Reader, desc ocispec.Descriptor) ([]byte, error) {
	if desc.Size < 0 {
		return nil, content.ErrInvalidDescriptorSize
	}
	buf := make([]byte, desc.Size)

	vr := content.NewVerifyReader(r, desc)
	if n, err := io.ReadFull(vr, buf); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("read failed: expected content size of %d, got %d, for digest %s: %w", desc.Size, n, desc.Digest.String(), err)
		}
		return nil, fmt.Errorf("read failed: %w", err)
	}
	if err := vr.Verify(); err != nil {
		return nil, err
	}
	return buf, nil
}
