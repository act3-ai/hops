package cache

import (
	"io"
	"os"
	"path/filepath"
)

// // Cache defines the interface
// type Cache interface {
// 	Exists(string) bool
// 	Fetch(string) (io.ReadCloser, error)
// }

// Dir represents a cache of stuff
type Dir string

// Exists reports if the key exists
func (c Dir) Exists(key string) bool {
	_, err := os.Stat(c.path(key))
	return err == nil
}

// Fetch fetches key from the cache
func (c Dir) Fetch(key string) (io.ReadCloser, error) {
	path := c.path(key)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (c Dir) path(key string) string {
	return filepath.Join(string(c), key)
}

// // FormulaKey formats a formula to a cache key
// func FormulaKey() string {
// }

// Cacheable defines a cacheable object type
type Cacheable interface {
	Key() string                       // produces the cache key for the object
	CacheInit(c Dir) (*os.File, error) // initializes the cached file for the object
}

// Exists checks if the cacheable object exists
func Exists(c Dir, obj Cacheable) bool {
	return c.Exists(obj.Key())
}

// Fetch fetches key from the cache
func Fetch(c Dir, obj Cacheable) (io.ReadCloser, error) {
	path := c.path(obj.Key())

	return os.Open(path)
}

// Init initializes the cache for an object
func Init(c Dir, obj Cacheable) (*os.File, error) {
	return obj.CacheInit(c)
}
