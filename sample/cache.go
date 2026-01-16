// Package sample provides a simple in-memory cache implementation.
//
// This package implements the gouache.Cache interface using Go's sync.Map
// for concurrent-safe operations without external dependencies.
package sample

import (
	"context"
	"sync"

	"github.com/soyacen/gouache"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is a simple in-memory cache implementation using sync.Map.
// It provides thread-safe operations for storing, retrieving, and deleting cached values.
type Cache struct {
	// cache is the underlying sync.Map used for storage.
	// sync.Map provides concurrent-safe operations without external dependencies.
	cache sync.Map
}

// Get retrieves a value from the cache by its key.
// It returns gouache.ErrCacheMiss if the key does not exist.
//
// Parameters:
//   - ctx: Context for the operation (not used in this implementation)
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails, or gouache.ErrCacheMiss if key doesn't exist
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	// Attempt to load the value from sync.Map
	val, ok := cache.cache.Load(key)

	// Return cache miss error if key doesn't exist
	if !ok {
		return nil, gouache.ErrCacheMiss
	}

	// Return the found value
	return val, nil
}

// Set stores a value in the cache under the specified key.
//
// Parameters:
//   - ctx: Context for the operation (not used in this implementation)
//   - key: The key under which the value will be stored
//   - val: The value to store
//
// Returns:
//   - Always returns nil as sync.Map.Store doesn't return errors
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Store the value in sync.Map
	cache.cache.Store(key, val)

	// sync.Map.Store doesn't return errors, so always return nil
	return nil
}

// Delete removes a value from the cache by its key.
//
// Parameters:
//   - ctx: Context for the operation (not used in this implementation)
//   - key: The key of the value to delete
//
// Returns:
//   - Always returns nil as sync.Map.Delete doesn't return errors
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Delete the value from sync.Map
	cache.cache.Delete(key)

	// sync.Map.Delete doesn't return errors, so always return nil
	return nil
}
