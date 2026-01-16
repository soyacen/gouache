package lru

import (
	"context"
	"testing"

	"github.com/soyacen/gouache"
	lru "github.com/hashicorp/golang-lru"
)

// TestNewCache tests the creation of a new Cache instance
func TestNewCache(t *testing.T) {
	lruCache, err := lru.New(100)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
	}

	if cache.Cache == nil {
		t.Error("LRUCache should not be nil")
	}
}

// TestCache_GetSet tests basic Get and Set operations
func TestCache_GetSet(t *testing.T) {
	lruCache, err := lru.New(100)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
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
	lruCache, err := lru.New(100)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
	}

	ctx := context.Background()
	key := "non-existent-key"

	// Test Get with non-existent key
	_, err = cache.Get(ctx, key)
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss, got %v", err)
	}
}

// TestCache_Delete tests deleting a key
func TestCache_Delete(t *testing.T) {
	lruCache, err := lru.New(100)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
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
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss after deletion, got %v", err)
	}
}

// TestCache_LRU_Eviction tests LRU eviction behavior
func TestCache_LRU_Eviction(t *testing.T) {
	// Create a small LRU cache with capacity of 2
	lruCache, err := lru.New(2)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
	}

	ctx := context.Background()

	// Add 3 items to trigger eviction (capacity is 2)
	err = cache.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Failed to set value1: %v", err)
	}

	err = cache.Set(ctx, "key2", "value2")
	if err != nil {
		t.Errorf("Failed to set value2: %v", err)
	}

	err = cache.Set(ctx, "key3", "value3")
	if err != nil {
		t.Errorf("Failed to set value3: %v", err)
	}

	// key1 should have been evicted (least recently used)
	_, err = cache.Get(ctx, "key1")
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss for evicted key, got %v", err)
	}

	// key2 and key3 should still be present
	_, err = cache.Get(ctx, "key2")
	if err != nil {
		t.Errorf("Failed to get value2: %v", err)
	}

	_, err = cache.Get(ctx, "key3")
	if err != nil {
		t.Errorf("Failed to get value3: %v", err)
	}
}

// TestCache_LRU_AccessOrder tests LRU access order behavior
func TestCache_LRU_AccessOrder(t *testing.T) {
	// Create a small LRU cache with capacity of 2
	lruCache, err := lru.New(2)
	if err != nil {
		t.Fatalf("Failed to create LRU cache: %v", err)
	}

	cache := &Cache{
		Cache: lruCache,
	}

	ctx := context.Background()

	// Add 2 items
	err = cache.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Failed to set value1: %v", err)
	}

	err = cache.Set(ctx, "key2", "value2")
	if err != nil {
		t.Errorf("Failed to set value2: %v", err)
	}

	// Access key1 to make it recently used
	_, err = cache.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Failed to get value1: %v", err)
	}

	// Add a third item - key2 should be evicted now (least recently used)
	err = cache.Set(ctx, "key3", "value3")
	if err != nil {
		t.Errorf("Failed to set value3: %v", err)
	}

	// key1 should still be present (recently accessed)
	_, err = cache.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Failed to get value1: %v", err)
	}

	// key2 should have been evicted
	_, err = cache.Get(ctx, "key2")
	if err != gouache.ErrCacheMiss {
		t.Errorf("Expected gouache.ErrCacheMiss for evicted key2, got %v", err)
	}

	// key3 should still be present
	_, err = cache.Get(ctx, "key3")
	if err != nil {
		t.Errorf("Failed to get value3: %v", err)
	}
}
