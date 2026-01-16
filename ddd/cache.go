// Package ddd (Delay Double Delete) provides a cache implementation that uses
// the delay double delete pattern to maintain cache consistency with a database.
//
// This package implements the gouache.Cache interface by wrapping a cache and
// database, ensuring cache consistency through a two-phase deletion process:
// 1. Immediate deletion from cache and database
// 2. Delayed secondary deletion from cache to handle race conditions
package ddd

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/soyacen/gouache"
)

// Ensure that cache implements the gouache.Cache interface at compile time.
var _ gouache.Cache = (*cache)(nil)

// Gopher is a function type that executes a given function asynchronously.
// It's used to run delayed operations in the background.
type Gopher func(f func()) error

// options holds configuration options for the delay double delete cache.
type options struct {
	// DelayDuration is the time to wait before performing the second cache deletion.
	DelayDuration time.Duration

	// DeleteTimeout is the timeout for the delayed delete operation.
	DeleteTimeout time.Duration

	// ErrorHandler is called when an error occurs during the delayed delete operation.
	ErrorHandler func(error)

	// Gopher is responsible for executing functions asynchronously.
	Gopher Gopher
}

// Option is a function that modifies the cache options.
type Option func(*options)

// WithDelayDuration returns an Option that sets the delay duration before
// the second cache deletion.
//
// Parameters:
//   - dur: The duration to wait before performing the second deletion
//
// Returns:
//   - An Option function that sets the DelayDuration
func WithDelayDuration(dur time.Duration) Option {
	return func(o *options) {
		o.DelayDuration = dur
	}
}

// WithDeleteTimeout returns an Option that sets the timeout for the delayed
// delete operation.
//
// Parameters:
//   - dur: The timeout duration for the delayed delete operation
//
// Returns:
//   - An Option function that sets the DeleteTimeout
func WithDeleteTimeout(dur time.Duration) Option {
	return func(o *options) {
		o.DeleteTimeout = dur
	}
}

// WithErrorHandler returns an Option that sets a custom error handler for
// errors that occur during the delayed delete operation.
//
// Parameters:
//   - f: A function to handle errors
//
// Returns:
//   - An Option function that sets the ErrorHandler
func WithErrorHandler(f func(error)) Option {
	return func(o *options) {
		o.ErrorHandler = f
	}
}

// WithGopher returns an Option that sets a custom Gopher function for
// executing delayed operations.
//
// Parameters:
//   - gopher: A function that executes other functions asynchronously
//
// Returns:
//   - An Option function that sets the Gopher
func WithGopher(gopher Gopher) Option {
	return func(o *options) {
		o.Gopher = gopher
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
//
// Returns:
//   - A pointer to the corrected options instance
func (o *options) Correct() *options {
	// Set default delay duration to 500ms if not specified or invalid
	if o.DelayDuration <= 0 {
		o.DelayDuration = 500 * time.Millisecond
	}

	// Set default delete timeout to 500s if not specified or invalid
	if o.DeleteTimeout <= 0 {
		o.DeleteTimeout = 500 * time.Second
	}

	// Set default error handler if not specified
	if o.ErrorHandler == nil {
		o.ErrorHandler = func(err error) {
			slog.Error("ddd.Cache.Get", slog.String("err", err.Error()))
		}
	}

	// Set default Gopher if not specified
	if o.Gopher == nil {
		o.Gopher = func(f func()) error {
			go f()
			return nil
		}
	}
	return o
}

// cache is a cache implementation that uses the delay double delete pattern
// to maintain consistency between cache and database.
type cache struct {
	// Options contains configuration options for the cache
	Options *options

	// Cache is the underlying cache implementation
	Cache gouache.Cache

	// Database is the underlying database implementation
	Database gouache.Database
}

// New creates a new delay double delete cache instance with the specified
// cache, database, and options.
//
// Parameters:
//   - c: The underlying cache implementation
//   - d: The underlying database implementation
//   - opts: Variable number of Option functions to configure the cache
//
// Returns:
//   - A gouache.Cache implementation that uses the delay double delete pattern
func New(c gouache.Cache, d gouache.Database, opts ...Option) gouache.Cache {
	return &cache{Options: newOptions(opts...), Cache: c, Database: d}
}

// Get retrieves a value from the cache by its key. If the value is not found
// in the cache, it attempts to retrieve it from the database and populate
// the cache with the result.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key to retrieve the value for
//
// Returns:
//   - The cached or database value or nil if not found
//   - An error if the operation fails
func (cache *cache) Get(ctx context.Context, key string) (any, error) {
	// Try to get the value from cache first
	val, err := cache.Cache.Get(ctx, key)

	// If cache miss, try to get from database
	if errors.Is(err, gouache.ErrCacheMiss) {
		// Get value from database
		val, err := cache.Database.Select(ctx, key)
		if err != nil {
			return nil, err
		}

		// Populate cache with database value
		return val, cache.Cache.Set(ctx, key, val)
	}

	// Return cache value or error
	return val, err
}

// Set stores a value in both the cache and database. It first deletes the
// existing cache entry, then upserts the value in the database, and finally
// schedules a delayed deletion of the cache entry to handle race conditions.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key under which the value will be stored
//   - val: The value to store
//
// Returns:
//   - An error if the operation fails
func (cache *cache) Set(ctx context.Context, key string, val any) error {
	// Delete existing cache entry
	if err := cache.Cache.Delete(ctx, key); err != nil {
		return err
	}

	// Upsert value in database
	if err := cache.Database.Upsert(ctx, key, val); err != nil {
		return err
	}

	// Schedule delayed cache deletion to handle race conditions
	return cache.Options.Gopher(func() {
		// Wait for the specified delay duration
		time.Sleep(cache.Options.DelayDuration)

		// Create a new context without the original cancellation
		ctx := context.WithoutCancel(ctx)

		// Add timeout to the context
		ctx, cancel := context.WithTimeout(ctx, cache.Options.DeleteTimeout)
		defer cancel()

		// Perform the second cache deletion
		if err := cache.Cache.Delete(ctx, key); err != nil {
			cache.Options.ErrorHandler(err)
		}
	})
}

// Delete removes a value from both the cache and database. It first deletes
// the value from the cache, then deletes it from the database, and finally
// schedules a delayed deletion of the cache entry to handle race conditions.
//
// Parameters:
//   - ctx: Context for the operation
//   - key: The key of the value to delete
//
// Returns:
//   - An error if the operation fails
func (cache *cache) Delete(ctx context.Context, key string) error {
	// Delete from cache
	if err := cache.Cache.Delete(ctx, key); err != nil {
		return err
	}

	// Delete from database
	if err := cache.Database.Delete(ctx, key); err != nil {
		return err
	}

	// Schedule delayed cache deletion to handle race conditions
	return cache.Options.Gopher(func() {
		// Wait for the specified delay duration
		time.Sleep(cache.Options.DelayDuration)

		// Create a new context without the original cancellation
		ctx := context.WithoutCancel(ctx)

		// Add timeout to the context
		ctx, cancel := context.WithTimeout(ctx, cache.Options.DeleteTimeout)
		defer cancel()

		// Perform the second cache deletion
		if err := cache.Cache.Delete(ctx, key); err != nil {
			cache.Options.ErrorHandler(err)
		}
	})
}
