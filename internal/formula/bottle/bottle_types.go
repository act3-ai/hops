package bottle

import (
	"context"
	"errors"
	"io"

	"github.com/act3-ai/hops/internal/formula"
)

// Bottle registry variations.
type (
	// Registry is a source of Bottles.
	Registry interface {
		// FetchBottle fetches a Bottle from the remote location
		FetchBottle(ctx context.Context, f formula.PlatformFormula) (io.ReadCloser, error)
	}

	// ConcurrentRegistry is a source of Bottles that supports concurrent fetching.
	ConcurrentRegistry interface {
		Registry
		FetchBottles(ctx context.Context, flist []formula.PlatformFormula) ([]io.ReadCloser, error)
	}
)

// Fetch fetches a Bottle from the BottleRegistry.
func Fetch(ctx context.Context, src Registry, f formula.PlatformFormula) (io.ReadCloser, error) {
	return src.FetchBottle(ctx, f)
}

// FetchAll fetches Bottles from the BottleRegistry.
func FetchAll(ctx context.Context, src Registry, formulae []formula.PlatformFormula) ([]io.ReadCloser, error) {
	// Closes all readers and returns a combined error
	closeAll := func(readers []io.ReadCloser) error {
		var err error
		for _, r := range readers {
			if r != nil {
				err = errors.Join(err, r.Close())
			}
		}
		return err
	}

	switch src := src.(type) {
	// Call FetchBottles for BottleRegistry with support for concurrency
	case ConcurrentRegistry:
		readers, err := src.FetchBottles(ctx, formulae)
		if err != nil {
			// As a convenience, close all readers if there was an error
			// (FetchBottles implementation may already do this)
			return nil, errors.Join(err, closeAll(readers))
		}
		return readers, nil
	// Call FetchBottle for all other BottleRegistry kinds
	default:
		readers := make([]io.ReadCloser, 0, len(formulae))
		for _, f := range formulae {
			f, err := src.FetchBottle(ctx, f)
			if err != nil {
				return nil, errors.Join(err, closeAll(readers))
			}
			readers = append(readers, f)
		}
		return readers, nil
	}
}
