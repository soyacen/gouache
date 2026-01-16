// Package lru provides an implementation of the gouache.Cache interface
// using hashicorp/golang-lru as the underlying storage mechanism.
//
// This package enables LRU (Least Recently Used) caching capabilities by leveraging
// the golang-lru library's efficient implementation of the LRU algorithm.
package lru

import (
	"context"

	"github.com/soyacen/gouache"
	lrucache "github.com/hashicorp/golang-lru"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is an implementation of gouache.Cache using LRU cache as the storage backend.
// It provides methods for storing, retrieving, and deleting cached values with
// LRU eviction policy when the cache reaches its capacity.
type Cache struct {
	// Cache is the underlying LRU cache instance used for storage.
	Cache *lrucache.Cache
}

// Get retrieves a value from the cache by its key.
// It returns gouache.ErrCacheMiss if the key does not exist.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails, or gouache.ErrCacheMiss if key doesn't exist
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	// Attempt to get the value from the LRU cache
	val, ok := cache.Cache.Get(key)

	// Handle case where entry is not found
	if !ok {
		return nil, gouache.ErrCacheMiss
	}

	// Return the found value
	return val, nil
}

// Set stores a value in the cache with the given key.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to store the value under
//   - val: The value to store
//
// Returns:
//   - Always returns nil as LRU cache Add operation is always successful
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Add the value to the LRU cache
	_ = cache.Cache.Add(key, val)
	return nil
}

// Delete removes a value from the cache by its key.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - Always returns nil as LRU cache Remove operation doesn't return errors
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Remove the value from the LRU cache
	_ = cache.Cache.Remove(key)
	return nil
}
