package gouache

import (
	"context"
	"hash"
	"hash/fnv"
	"testing"

	"github.com/soyacen/gouache"
)

// mockCache is a simple in-memory cache implementation for testing purposes.
type mockCache struct {
	data map[string]any
}

// newMockCache creates a new mockCache instance.
func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]any),
	}
}

// Get retrieves a value from the cache by its key.
func (m *mockCache) Get(ctx context.Context, key string) (any, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, gouache.ErrCacheMiss
}

// Set stores a value in the cache under the specified key.
func (m *mockCache) Set(ctx context.Context, key string, val any) error {
	m.data[key] = val
	return nil
}

// Delete removes a value from the cache by its key.
func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// TestNew tests the New function for creating a sharded cache.
func TestNew(t *testing.T) {
	// Test successful creation
	buckets := []gouache.Cache{newMockCache(), newMockCache()}
	cache := New(buckets)
	if cache == nil {
		t.Error("Expected cache to be created, but got nil")
	}

	// Test panic when buckets is empty
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when buckets is empty, but did not panic")
		}
	}()
	New([]gouache.Cache{})
}

// TestShardedCache_Get tests the Get method of the sharded cache.
func TestShardedCache_Get(t *testing.T) {
	buckets := []gouache.Cache{newMockCache(), newMockCache()}
	cache := New(buckets)

	// Set up a value in one of the buckets
	key := "test-key"
	value := "test-value"
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Fatalf("Failed to set up test value: %v", err)
	}

	// Test getting an existing value
	result, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Errorf("Unexpected error when getting value: %v", err)
	}
	if result != value {
		t.Errorf("Expected %v, but got %v", value, result)
	}

	// Test getting a non-existing value
	_, err = cache.Get(context.Background(), "non-existent-key")
	if err == nil {
		t.Error("Expected ErrCacheMiss for non-existent key, but got no error")
	} else if err != gouache.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss for non-existent key, but got: %v", err)
	}
}

// TestShardedCache_Set tests the Set method of the sharded cache.
func TestShardedCache_Set(t *testing.T) {
	buckets := []gouache.Cache{newMockCache(), newMockCache()}
	cache := New(buckets)

	// Test setting a value
	key := "test-key"
	value := "test-value"
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Errorf("Unexpected error when setting value: %v", err)
	}

	// Verify the value was set by getting it
	result, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Errorf("Unexpected error when getting value: %v", err)
	}
	if result != value {
		t.Errorf("Expected %v, but got %v", value, result)
	}
}

// TestShardedCache_Delete tests the Delete method of the sharded cache.
func TestShardedCache_Delete(t *testing.T) {
	buckets := []gouache.Cache{newMockCache(), newMockCache()}
	cache := New(buckets)

	// Set up a value
	key := "test-key"
	value := "test-value"
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Fatalf("Failed to set up test value: %v", err)
	}

	// Test deleting the value
	err = cache.Delete(context.Background(), key)
	if err != nil {
		t.Errorf("Unexpected error when deleting value: %v", err)
	}

	// Verify the value was deleted
	_, err = cache.Get(context.Background(), key)
	if err == nil {
		t.Error("Expected ErrCacheMiss after deletion, but got no error")
	} else if err != gouache.ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after deletion, but got: %v", err)
	}

	// Test deleting a non-existent key (should not return an error)
	err = cache.Delete(context.Background(), "non-existent-key")
	if err != nil {
		t.Errorf("Unexpected error when deleting non-existent key: %v", err)
	}
}

// customHashFactory is a custom HashFactory for testing purposes.
func customHashFactory(ctx context.Context, key string) (hash.Hash, error) {
	return fnv.New64a(), nil
}

// TestShardedCache_WithCustomHashFactory tests the sharded cache with a custom hash factory.
func TestShardedCache_WithCustomHashFactory(t *testing.T) {
	buckets := []gouache.Cache{newMockCache(), newMockCache()}
	cache := New(buckets, WithHashFactory(customHashFactory))

	// Test setting and getting a value with custom hash factory
	key := "test-key"
	value := "test-value"
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Errorf("Unexpected error when setting value with custom hash factory: %v", err)
	}

	result, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Errorf("Unexpected error when getting value with custom hash factory: %v", err)
	}
	if result != value {
		t.Errorf("Expected %v, but got %v when using custom hash factory", value, result)
	}
}

// TestShardedCache_BucketDistribution tests that keys are distributed across buckets.
func TestShardedCache_BucketDistribution(t *testing.T) {
	// Create mock caches that can track which keys they store
	bucket1 := newMockCache()
	bucket2 := newMockCache()
	buckets := []gouache.Cache{bucket1, bucket2}
	cache := New(buckets)

	// Set multiple keys
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, key := range keys {
		err := cache.Set(context.Background(), key, "value")
		if err != nil {
			t.Errorf("Unexpected error when setting key %s: %v", key, err)
		}
	}

	// Verify that keys are distributed across buckets (at least some in each)
	bucket1Count := len(bucket1.data)
	bucket2Count := len(bucket2.data)

	if bucket1Count == 0 {
		t.Error("Expected bucket1 to contain some keys, but it was empty")
	}

	if bucket2Count == 0 {
		t.Error("Expected bucket2 to contain some keys, but it was empty")
	}

	if bucket1Count+bucket2Count != len(keys) {
		t.Errorf("Expected total keys to be %d, but got %d", len(keys), bucket1Count+bucket2Count)
	}
}
