package symlink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CreateOption represents an option
type CreateOption func(oldname, newname string) error

// Overwrite overwrites existing files
var Overwrite CreateOption = func(_, newname string) error {
	// Remove current file
	err := os.Remove(newname)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// MkdirAll creates the parent directory
var MkdirAll CreateOption = func(_, newname string) error {
	// Create parent directory
	err := os.MkdirAll(filepath.Dir(newname), 0o775)
	if err != nil {
		return err
	}
	return nil
}

// Options contains options for creating symlinks
type Options struct {
	Name        string // Name to prefix dry-run messages with
	MkdirParent bool   // Create parent directory of symlink if it does not exist
	Overwrite   bool   // Delete files that already exist in the prefix while linking
	DryRun      bool   // List files which would be linked or overwritten without actually linking or deleting any files
}

// Relative creates a relative symlink at newname to the location specified by oldname
func Relative(oldname, newname string, opts *Options) error {
	// Evaluate relative path
	relativeOldname, err := filepath.Rel(filepath.Dir(newname), oldname)
	if err != nil {
		return fmt.Errorf("calculating relative path to link from %s to %s: %w", oldname, newname, err)
	}

	prefix := ""
	if opts.Name != "" {
		prefix = "[" + opts.Name + "] "
	}
	dryrun := func(msg string) {
		fmt.Println(prefix + msg)
	}

	_, err = os.Lstat(newname)
	switch {
	// Symlink does not exist
	case errors.Is(err, os.ErrNotExist):
		if opts.DryRun {
			dryrun(fmt.Sprintf("ln -s %s %s", newname, oldname))
			return nil
		}
	// Unspecified lstat error
	case err != nil:
		return fmt.Errorf("checking for existing symlink %s: %w", newname, err)
	// Symlink does exist and should be overwritten
	case opts.Overwrite:
		if opts.DryRun {
			dryrun(fmt.Sprintf("rm %s && ln -s %s %s (overwrite)", newname, oldname, newname))
			return nil
		}

		// Remove current file at target
		err := os.Remove(newname)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	// Symlink exists but should not be overwritten
	default:
		if opts.DryRun {
			dryrun(fmt.Sprintf("ln -s %s %s (skipped, would overwrite %s)", oldname, newname, newname))
			return nil
		}
		return nil
	}

	if opts.MkdirParent {
		// Create parent directory
		err := os.MkdirAll(filepath.Dir(newname), 0o775)
		if err != nil {
			return fmt.Errorf("creating parent directory for symlink %s: %w", newname, err)
		}
	}

	// Create symlink with the relative path
	return os.Symlink(relativeOldname, newname)
}
