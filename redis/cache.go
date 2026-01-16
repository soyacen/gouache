// Package redis provides an implementation of the gouache.Cache interface
// using Redis as the underlying storage mechanism.
//
// This package enables distributed caching capabilities by leveraging
// Redis's persistence, replication, and clustering features.
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/soyacen/gouache"
	"github.com/redis/go-redis/v9"
)

// Ensure that Cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*Cache)(nil)

// Cache is an implementation of gouache.Cache using Redis as the storage backend.
// It provides methods for storing, retrieving, and deleting cached values with
// support for custom serialization/deserialization and configurable TTL.
type Cache struct {
	// Cache is the underlying Redis client instance used for storage operations.
	Cache redis.Cmdable

	// TTL is an optional function to determine the time-to-live duration for a cache entry.
	// If not provided, entries will not expire by default.
	TTL func(ctx context.Context, key string, val any) (time.Duration, error)

	// Marshal is an optional function to serialize objects into strings.
	// If not provided, default type conversions are used for basic types.
	Marshal func(key string, obj any) (string, error)

	// Unmarshal is an optional function to deserialize strings into objects.
	// If not provided, raw strings are returned.
	Unmarshal func(key string, data string) (any, error)
}

// Get retrieves a value from the Redis cache by its key.
// It returns gouache.ErrCacheMiss if the key does not exist.
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails, or gouache.ErrCacheMiss if key doesn't exist
func (cache *Cache) Get(ctx context.Context, key string) (any, error) {
	// Attempt to get the value from Redis
	data, err := cache.Cache.Get(ctx, key).Result()

	// Handle case where entry is not found
	if errors.Is(err, redis.Nil) {
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

// Set stores a value in the Redis cache under the specified key.
// It handles both raw strings and custom objects that require marshaling.
// TTL can be determined dynamically by the TTL function if provided.
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - key: The key under which the value will be stored
//   - val: The value to store, either as string or any other type requiring marshaling
//
// Returns:
//   - An error if the operation fails, including when Marshal is nil for non-string values
func (cache *Cache) Set(ctx context.Context, key string, val any) error {
	// Initialize TTL to zero (no expiration)
	ttl := time.Duration(0)

	// Check if a custom TTL function is configured
	if cache.TTL != nil {
		var err error
		// Use the TTL function to determine expiration duration
		ttl, err = cache.TTL(ctx, key, val)
		if err != nil {
			return err
		}
	}

	// Check if the value is already a string
	if data, ok := val.(string); ok {
		// Directly store strings without marshaling
		return cache.Cache.Set(ctx, key, data, ttl).Err()
	}

	// For non-string values, ensure a marshal function is available
	if cache.Marshal == nil {
		return errors.New("gouache: Marshal is nil")
	}

	// Marshal the value into string using the custom marshal function
	data, err := cache.Marshal(key, val)
	if err != nil {
		return err
	}

	// Store the marshaled data in Redis
	return cache.Cache.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from the Redis cache by its key.
//
// Parameters:
//   - ctx: Context for the Redis operation
//   - key: The key of the value to delete
//
// Returns:
//   - An error if the operation fails
func (cache *Cache) Delete(ctx context.Context, key string) error {
	// Delegate deletion to the underlying Redis client instance
	return cache.Cache.Del(ctx, key).Err()
}
