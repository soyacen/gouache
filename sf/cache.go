// Package sf provides a cache implementation that uses singleflight to
// prevent duplicate cache operations for the same key.
//
// This package implements the gouache.Cache interface by wrapping an existing
// cache and using singleflight to ensure only one operation of a given type
// is performed for a given key at a time. This helps reduce the thundering herd
// problem when multiple goroutines request the same missing cache entry.
package sf

import (
	"context"

	"github.com/soyacen/gouache"
	"golang.org/x/sync/singleflight"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is a cache implementation that wraps another cache and uses singleflight
// to prevent duplicate operations for the same key.
//
// This implementation only applies singleflight to Get operations, as these are
// typically the most expensive and prone to the thundering herd problem. Set and
// Delete operations are passed through directly to the underlying cache.
type Cache struct {
	// Cache is the underlying cache implementation that stores the actual data.
	Cache gouache.Cache

	// group is the singleflight group used to deduplicate Get operations.
	group singleflight.Group
}

// Get retrieves a value from the cache by its key.
//
// If multiple goroutines attempt to get the same key simultaneously, only one
// will actually perform the operation while the others wait for the result.
// This helps prevent the thundering herd problem when accessing missing or
// expired cache entries.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	// Use singleflight to ensure only one Get operation for this key runs at a time
	val, err, _ := cache.group.Do(key, func() (any, error) {
		// Delegate to the underlying cache
		return cache.Cache.Get(ctx, key)
	})
	return val, err
}

// Set stores a value in the cache under the specified key.
//
// This operation is passed through directly to the underlying cache without
// any singleflight protection, as Set operations typically don't suffer from
// the thundering herd problem.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key under which the value will be stored
//   - val: The value to store
//
// Returns:
//   - An error if the operation fails
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Delegate directly to the underlying cache
	return cache.Cache.Set(ctx, key, val)
}

// Delete removes a value from the cache by its key.
//
// This operation is passed through directly to the underlying cache without
// any singleflight protection, as Delete operations typically don't suffer from
// the thundering herd problem.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - An error if the operation fails
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Delegate directly to the underlying cache
	return cache.Cache.Delete(ctx, key)
}