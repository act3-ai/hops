package iterutil

import (
	"errors"

	"github.com/sourcegraph/conc/iter"
)

// ForEachIdxErr wraps iter.ForEachIdx with an error return.
func ForEachIdxErr[T any](iterator iter.Iterator[T], input []T, f func(int, *T) error) error {
	errs := make([]error, len(input))
	iterator.ForEachIdx(input, func(i int, t *T) {
		errs[i] = f(i, t)
	})
	return errors.Join(errs...)
}

// ForEachErr wraps iter.ForEach with an error return.
func ForEachErr[T any](iterator iter.Iterator[T], input []T, f func(*T) error) error {
	errs := make([]error, len(input))
	iterator.ForEachIdx(input, func(i int, t *T) {
		errs[i] = f(t)
	})
	return errors.Join(errs...)
}
