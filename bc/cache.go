// Package bc provides an implementation of the gouache.Cache interface
// using allegro/bigcache as the underlying storage mechanism.
//
// This package enables high-performance caching capabilities by leveraging
// BigCache's efficient memory management and concurrent access patterns.
package bc

import (
	"context"
	"errors"

	"github.com/allegro/bigcache/v3"
	"github.com/soyacen/gouache"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is an implementation of gouache.Cache using BigCache as the storage backend.
// It provides methods for storing, retrieving, and deleting cached values with
// support for custom serialization and deserialization functions.
type Cache struct {
	// Cache is the underlying Cache instance used for storage.
	Cache *bigcache.BigCache

	// Marshal is an optional function to serialize objects into bytes.
	// If not provided, default type conversions are used for basic types.
	Marshal func(key string, obj any) ([]byte, error)

	// Unmarshal is an optional function to deserialize bytes into objects.
	// If not provided, raw bytes are returned.
	Unmarshal func(key string, data []byte) (any, error)
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
	// Attempt to get the value from BigCache
	data, err := cache.Cache.Get(key)

	// Handle case where entry is not found
	if errors.Is(err, bigcache.ErrEntryNotFound) {
		return nil, gouache.ErrCacheMiss
	}

	// Return other errors as-is
	if err != nil {
		return nil, err
	}

	// If no unmarshal function is defined, return raw data
	if cache.Unmarshal == nil {
		return data, nil
	}

	// Use custom unmarshal function to decode the data
	obj, err := cache.Unmarshal(key, data)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// Set stores a value in the cache under the specified key.
// It handles both raw byte slices and custom objects that require marshaling.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key under which the value will be stored
//   - val: The value to store, either as []byte or any other type requiring marshaling
//
// Returns:
//   - An error if the operation fails, including when Marshal is nil for non-byte values
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Check if the value is already a byte slice
	if data, ok := val.([]byte); ok {
		// Directly store byte slices without marshaling
		return cache.Cache.Set(key, data)
	}

	// For non-byte values, ensure a marshal function is available
	if cache.Marshal == nil {
		return errors.New("gouache: Marshal is nil")
	}

	// Marshal the value into bytes using the custom marshal function
	data, err := cache.Marshal(key, val)
	if err != nil {
		return err
	}

	// Store the marshaled data in BigCache
	return cache.Cache.Set(key, data)
}

// Delete removes a value from the cache by its key.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - An error if the operation fails
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Delegate deletion to the underlying BigCache instance
	return cache.Cache.Delete(key)
}
