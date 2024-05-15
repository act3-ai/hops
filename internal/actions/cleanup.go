package actions

import (
	"context"
	"fmt"
	"os"
)

// Cleanup represents the action and its options
type Cleanup struct {
	*Hops
}

// Run runs the action
func (action *Cleanup) Run(_ context.Context) error {
	broken, err := action.Prefix().BrokenLinks()
	if err != nil {
		return err
	}

	for _, bl := range broken {
		fmt.Println("Removing " + bl)
		err = os.Remove(bl)
		if err != nil {
			return err
		}
	}

	// TODO: remove outdated, unlinked kegs

	fmt.Printf("Pruned %d symbolic links and %d directories from %s\n", len(broken), 0, action.Prefix())

	return nil
}
