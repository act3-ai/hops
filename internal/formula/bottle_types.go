package formula

import (
	"context"
	"errors"
	"io"
)

// Bottle registry variations
type (
	// bottleRegistry is a source of Bottles
	BottleRegistry interface {
		// FetchBottle fetches a Bottle from the remote location
		FetchBottle(ctx context.Context, f PlatformFormula) (io.ReadCloser, error)
	}

	// concurrentBottleRegistry is a source of Bottles that supports concurrent fetching
	ConcurrentBottleRegistry interface {
		BottleRegistry
		FetchBottles(ctx context.Context, flist []PlatformFormula) ([]io.ReadCloser, error)
	}
)

// FetchBottle fetches a Bottle from the BottleRegistry
func FetchBottle(ctx context.Context, src BottleRegistry, f PlatformFormula) (io.ReadCloser, error) {
	return src.FetchBottle(ctx, f)
}

// FetchAll fetches Formulae from the Formulary
func FetchBottles(ctx context.Context, src BottleRegistry, formulae []PlatformFormula) ([]io.ReadCloser, error) {
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
	case ConcurrentBottleRegistry:
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
