// Package gouache provides a sharded cache implementation that distributes
// cache entries across multiple buckets to reduce lock contention and improve
// concurrent performance.
//
// This package implements the gouache.Cache interface using a sharding strategy
// where cache entries are distributed across multiple underlying cache buckets
// based on a hash of the key.
package gouache

import (
	"context"
	"encoding/binary"
	"hash"
	"hash/fnv"

	"github.com/soyacen/gouache"
)

// Ensure that cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*cache)(nil)

// HashFactory is a function type that creates a new hash.Hash instance
// for a given context and key. This allows customization of the hashing
// algorithm used for sharding.
type HashFactory func(ctx context.Context, key string) (hash.Hash, error)

// cache is a sharded cache implementation that distributes entries across
// multiple buckets to improve concurrent access performance.
type cache struct {
	// Options contains configuration options for the cache
	Options *options

	// Buckets is a slice of underlying cache implementations that store
	// the actual cached data. Each entry is assigned to a bucket based
	// on a hash of its key.
	Buckets []gouache.Cache
}

// options holds configuration options for the sharded cache.
type options struct {
	// HashFactory is a function that creates hash instances used for
	// determining which bucket a key should be stored in.
	HashFactory HashFactory
}

// Option is a function that modifies the cache options.
type Option func(*options)

// WithHashFactory returns an Option that sets a custom HashFactory function.
// This allows users to specify a different hashing algorithm for sharding.
//
// Parameters:
//   - hashFactory: A function that creates hash instances for key distribution
//
// Returns:
//   - An Option function that sets the HashFactory
func WithHashFactory(hashFactory HashFactory) Option {
	return func(o *options) {
		o.HashFactory = hashFactory
	}
}

// newOptions creates a new options instance with default values and applies
// the provided options.
//
// Parameters:
//   - opts: Variable number of Option functions to apply
//
// Returns:
//   - A pointer to the configured options instance
func newOptions(opts ...Option) *options {
	options := &options{}
	return options.Apply(opts...).Correct()
}

// Apply applies the provided options to the options instance.
//
// Parameters:
//   - opts: Variable number of Option functions to apply
//
// Returns:
//   - A pointer to the modified options instance
func (o *options) Apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Correct ensures that all options have valid default values.
// If HashFactory is nil, it sets a default FNV-32a hash factory.
//
// Returns:
//   - A pointer to the corrected options instance
func (o *options) Correct() *options {
	if o.HashFactory == nil {
		o.HashFactory = func(ctx context.Context, key string) (hash.Hash, error) {
			return fnv.New32a(), nil
		}
	}
	return o
}

// New creates a new sharded cache instance with the specified buckets and options.
//
// Parameters:
//   - buckets: A slice of gouache.Cache instances to use as buckets
//   - opts: Variable number of Option functions to configure the cache
//
// Returns:
//   - A gouache.Cache implementation that distributes entries across buckets
//
// Panics:
//   - If the buckets slice is empty
func New(buckets []gouache.Cache, opts ...Option) gouache.Cache {
	if len(buckets) == 0 {
		panic("gouache: buckets is empty")
	}
	return &cache{Options: newOptions(opts...), Buckets: buckets}
}

// Get retrieves a value from the cache by its key.
// The key is hashed to determine which bucket contains the value.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached value or nil if not found
//   - An error if the operation fails
func (cache *cache) Get(ctx context.Context, key string) (any, error) {
	bucket, err := cache.bucket(ctx, key)
	if err != nil {
		return nil, err
	}
	return bucket.Get(ctx, key)
}

// Set stores a value in the cache under the specified key.
// The key is hashed to determine which bucket should store the value.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key under which the value will be stored
//   - val: The value to store
//
// Returns:
//   - An error if the operation fails
func (cache *cache) Set(ctx context.Context, key string, val any) error {
	bucket, err := cache.bucket(ctx, key)
	if err != nil {
		return err
	}
	return bucket.Set(ctx, key, val)
}

// Delete removes a value from the cache by its key.
// The key is hashed to determine which bucket contains the value to delete.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - An error if the operation fails
func (cache *cache) Delete(ctx context.Context, key string) error {
	bucket, err := cache.bucket(ctx, key)
	if err != nil {
		return err
	}
	return bucket.Delete(ctx, key)
}

// bucket determines which bucket should handle operations for a given key.
// It uses the configured HashFactory to hash the key and distribute it
// across the available buckets.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to determine the bucket for
//
// Returns:
//   - The gouache.Cache bucket that should handle operations for the key
//   - An error if the hash factory or write operation fails
func (cache *cache) bucket(ctx context.Context, key string) (gouache.Cache, error) {
	// Create a new hash instance using the configured HashFactory
	h, err := cache.Options.HashFactory(ctx, key)
	if err != nil {
		return nil, err
	}

	// Write the key to the hash
	if _, err := h.Write([]byte(key)); err != nil {
		return nil, err
	}

	// Determine the bucket based on the hash size
	switch h.Size() {
	case 4:
		// For 32-bit hashes, use the hash's Sum32 method
		sum32 := h.(hash.Hash32).Sum32()
		return cache.Buckets[sum32%uint32(len(cache.Buckets))], nil
	case 8:
		// For 64-bit hashes, use the hash's Sum64 method
		sum64 := h.(hash.Hash64).Sum64()
		return cache.Buckets[sum64%uint64(len(cache.Buckets))], nil
	default:
		// For other hash sizes, use the raw bytes
		sum := h.Sum(nil)
		// If the hash is less than 4 bytes, use the first bucket
		if len(sum) < 4 {
			return cache.Buckets[0], nil
		}
		// Extract a 32-bit value from the hash and use it to determine the bucket
		sum32 := int(binary.BigEndian.Uint32(sum[:4]))
		return cache.Buckets[sum32%len(cache.Buckets)], nil
	}
}
