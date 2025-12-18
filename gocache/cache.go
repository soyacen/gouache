// Package gc provides an implementation of the gouache.Cache interface
// using patrickmn/go-cache as the underlying storage mechanism.
//
// This package enables in-memory caching with expiration capabilities by leveraging
// go-cache's thread-safe operations and automatic expiration handling.
package gc

import (
	"context"
	"time"

	"github.com/go-leo/gouache"
	gocache "github.com/patrickmn/go-cache"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is an implementation of gouache.Cache using go-cache as the storage backend.
// It provides methods for storing, retrieving, and deleting cached values with
// support for configurable time-to-live (TTL) settings.
type Cache struct {
	// Cache is the underlying go-cache instance used for storage.
	Cache *gocache.Cache

	// TTL is an optional function to determine the time-to-live duration for a cache entry.
	// If not provided, the default expiration behavior of go-cache is used.
	TTL func(ctx context.Context, key string, val any) (time.Duration, error)
}

// Get retrieves a value from the cache by its key.
// It returns gouache.ErrCacheMiss if the key does not exist or has expired.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails, or gouache.ErrCacheMiss if key doesn't exist
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	// Attempt to get the value from the go-cache
	val, ok := cache.Cache.Get(key)

	// Handle case where entry is not found or has expired
	if !ok {
		return nil, gouache.ErrCacheMiss
	}

	// Return the found value
	return val, nil
}

// Set stores a value in the cache under the specified key with an optional TTL.
// The TTL (time-to-live) can be determined dynamically by the TTL function if provided,
// otherwise uses the default expiration behavior of go-cache.
//
// Parameters:
//   - ctx: Context for the operation, passed to the TTL function if configured
//   - key: The key under which the value will be stored
//   - val: The value to store in the cache
//
// Returns:
//   - An error if the TTL function (if configured) returns an error, otherwise nil
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Initialize TTL to default expiration value
	ttl := gocache.DefaultExpiration

	// Check if a custom TTL function is configured
	if cache.TTL != nil {
		// Use the TTL function to determine expiration duration
		var err error
		ttl, err = cache.TTL(ctx, key, val)
		if err != nil {
			return err
		}
		// Store the value with the computed TTL
		cache.Cache.Set(key, val, ttl)
		return nil
	}

	// Store the value with default expiration
	cache.Cache.Set(key, val, ttl)
	return nil
}

// Delete removes a value from the cache by its key.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - Always returns nil as go-cache.Delete doesn't return errors
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Delegate deletion to the underlying go-cache instance
	cache.Cache.Delete(key)
	return nil
}
