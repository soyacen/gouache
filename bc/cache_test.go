package bc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/soyacen/gouache"
)

// TestNewCache tests the creation of a new Cache instance
func TestNewCache(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
	}

	if cache.Cache == nil {
		t.Error("BigCache should not be nil")
	}
}

// TestCache_GetSet tests basic Get and Set operations
func TestCache_GetSet(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
		Marshal: func(key string, obj any) ([]byte, error) {
			return json.Marshal(obj)
		},
		Unmarshal: func(key string, data []byte) (any, error) {
			var obj any
			err := json.Unmarshal(data, &obj)
			return obj, err
		},
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Test Set
	err = cache.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Test Get
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	if result != value {
		t.Errorf("Expected %v, got %v", value, result)
	}
}

// TestCache_GetNonExistentKey tests getting a non-existent key
func TestCache_GetNonExistentKey(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
	}

	ctx := context.Background()
	key := "non-existent-key"

	// Test Get with non-existent key
	_, err = cache.Get(ctx, key)
	if !errors.Is(err, gouache.ErrCacheMiss) {
		t.Errorf("Expected gouache.ErrCacheMiss, got %v", err)
	}
}

// TestCache_Delete tests deleting a key
func TestCache_Delete(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
		Marshal: func(key string, obj any) ([]byte, error) {
			return json.Marshal(obj)
		},
	}

	ctx := context.Background()
	key := "test-key"
	value := "test-value"

	// Set a value
	err = cache.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Delete the value
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Errorf("Failed to delete value: %v", err)
	}

	// Try to get the deleted value
	_, err = cache.Get(ctx, key)
	if !errors.Is(err, gouache.ErrCacheMiss) {
		t.Errorf("Expected gouache.ErrCacheMiss after deletion, got %v", err)
	}
}

// TestCache_SetWithoutMarshal tests Set operation without custom Marshal function
func TestCache_SetWithoutMarshal(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
	}

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	// Test Set with []byte value and no Marshal function
	err = cache.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set []byte value without Marshal function: %v", err)
	}

	// Test Get
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	if string(result.([]byte)) != string(value) {
		t.Errorf("Expected %v, got %v", string(value), string(result.([]byte)))
	}
}

// TestCache_SetUnsupportedTypeWithoutMarshal tests Set with unsupported type when no Marshal function
func TestCache_SetUnsupportedTypeWithoutMarshal(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
	}

	ctx := context.Background()
	key := "test-key"
	value := 123 // int is not supported without Marshal function

	// Test Set with unsupported type and no Marshal function
	err = cache.Set(ctx, key, value)
	if err == nil {
		t.Error("Expected error when setting unsupported type without Marshal function")
	}
}

// TestCache_GetWithoutUnmarshal tests Get operation without custom Unmarshal function
func TestCache_GetWithoutUnmarshal(t *testing.T) {
	config := bigcache.DefaultConfig(5 * time.Minute)
	bigCache, err := bigcache.NewBigCache(config)
	if err != nil {
		t.Fatalf("Failed to create bigcache: %v", err)
	}

	cache := &Cache{
		Cache: bigCache,
		Marshal: func(key string, obj any) ([]byte, error) {
			return obj.([]byte), nil
		},
		// No Unmarshal function
	}

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	// Set a value
	err = cache.Set(ctx, key, value)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	// Get the value (should return raw bytes)
	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	if string(result.([]byte)) != string(value) {
		t.Errorf("Expected %v, got %v", string(value), string(result.([]byte)))
	}
}
