package iterutil

import (
	"errors"

	"github.com/sourcegraph/conc/iter"
)

// ForEachIdxErr wraps iter.ForEachIdx with an error return
func ForEachIdxErr[T any](iterator iter.Iterator[T], input []T, f func(int, *T) error) error {
	errs := make([]error, len(input))
	iterator.ForEachIdx(input, func(i int, t *T) {
		err := f(i, t)
		if err != nil {
			errs[i] = err
		}
	})
	return errors.Join(errs...)
}
