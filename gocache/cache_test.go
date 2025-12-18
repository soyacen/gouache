package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-leo/gouache"
	"github.com/patrickmn/go-cache"
)

// TestNewCache tests the creation of a new Cache instance
func TestNewCache(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	cacheImpl := &Cache{
		Cache: goCache,
	}

	if cacheImpl.Cache == nil {
		t.Error("Cache should not be nil")
	}
}

// TestCache_GetSet tests basic Get and Set operations
func TestCache_GetSet(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	cacheImpl := &Cache{
		Cache: goCache,
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test Set
	err := cacheImpl.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Test Get
	result, err := cacheImpl.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
}

// TestCache_GetNonExistentKey tests getting a non-existent key
func TestCache_GetNonExistentKey(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	cacheImpl := &Cache{
		Cache: goCache,
	}

	ctx := context.Background()
	key := "non-existent-key"

	// Test Get with non-existent key
	_, err := cacheImpl.Get(ctx, key)
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss, got %v", err)
	}
}

// TestCache_Delete tests deleting a key
func TestCache_Delete(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	cacheImpl := &Cache{
		Cache: goCache,
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value
	err := cacheImpl.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Delete the value
	err = cacheImpl.Delete(ctx, key)
	if err != nil {
		t.Errorf("Failed to delete value: %v", err)
	}

	// Try to get the deleted value
	_, err = cacheImpl.Get(ctx, key)
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss after deletion, got %v", err)
	}
}

// TestCache_WithTTL tests Set operation with custom TTL function
func TestCache_WithTTL(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	cacheImpl := &Cache{
		Cache: goCache,
		TTL: func(ctx context.Context, key string, val any) (time.Duration, error) {
			return 1 * time.Second, nil
		},
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test Set with TTL
	err := cacheImpl.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value with TTL: %v", err)
	}

	// Test Get
	result, err := cacheImpl.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
}

// TestCache_ExpiredKey tests behavior with expired keys
func TestCache_ExpiredKey(t *testing.T) {
	goCache := cache.New(1*time.Millisecond, 1*time.Millisecond)

	cacheImpl := &Cache{
		Cache: goCache,
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value
	err := cacheImpl.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to get the expired value
	_, err = cacheImpl.Get(ctx, key)
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss for expired key, got %v", err)
	}
}

// TestCache_TTLWithError tests TTL function that returns an error
func TestCache_TTLWithError(t *testing.T) {
	goCache := cache.New(5*time.Minute, 10*time.Minute)

	expectedErr := errors.New("TTL error")
	cacheImpl := &Cache{
		Cache: goCache,
		TTL: func(ctx context.Context, key string, val any) (time.Duration, error) {
			return 0, expectedErr
		},
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test Set with TTL that returns error
	err := cacheImpl.Set(ctx, key, value)
	if err != expectedErr {
		t.Errorf("Expected TTL error, got %v", err)
	}
}
