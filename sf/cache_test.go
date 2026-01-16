package sf

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/soyacen/gouache"
)

// mockCache is a simple in-memory cache implementation for testing purposes.
type mockCache struct {
	data  map[string]any
	mu    sync.RWMutex
	delay time.Duration // Simulate slow operations
}

// newMockCache creates a new mockCache instance.
func newMockCache(delay time.Duration) *mockCache {
	return &mockCache{
		data:  make(map[string]any),
		delay: delay,
	}
}

// Get retrieves a value from the cache by its key.
func (m *mockCache) Get(ctx context.Context, key string) (any, error) {
	// Simulate delay
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, gouache.ErrCacheMiss
}

// Set stores a value in the cache under the specified key.
func (m *mockCache) Set(ctx context.Context, key string, val any) error {
	// Simulate delay
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = val
	return nil
}

// Delete removes a value from the cache by its key.
func (m *mockCache) Delete(ctx context.Context, key string) error {
	// Simulate delay
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

// errorCache is a cache implementation that always returns an error for testing purposes.
type errorCache struct{}

func (e *errorCache) Get(ctx context.Context, key string) (any, error) {
	return nil, errors.New("intentional error")
}

func (e *errorCache) Set(ctx context.Context, key string, val any) error {
	return errors.New("intentional error")
}

func (e *errorCache) Delete(ctx context.Context, key string) error {
	return errors.New("intentional error")
}

// TestSF_Cache_Get tests the Get method of the singleflight cache.
func TestSF_Cache_Get(t *testing.T) {
	// Test successful Get operation
	t.Run("Successful Get", func(t *testing.T) {
		underlying := newMockCache(0)
		sfCache := &Cache{Cache: underlying}

		// Set up test data
		key := "test-key"
		value := "test-value"
		err := underlying.Set(context.Background(), key, value)
		if err != nil {
			t.Fatalf("Failed to set up test data: %v", err)
		}

		// Test Get
		result, err := sfCache.Get(context.Background(), key)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != value {
			t.Errorf("Expected %v, but got %v", value, result)
		}
	})

	// Test cache miss
	t.Run("Cache Miss", func(t *testing.T) {
		underlying := newMockCache(0)
		sfCache := &Cache{Cache: underlying}

		// Test Get with non-existent key
		_, err := sfCache.Get(context.Background(), "non-existent-key")
		if err == nil {
			t.Error("Expected error for cache miss, but got nil")
		} else if err != gouache.ErrCacheMiss {
			t.Errorf("Expected ErrCacheMiss, but got: %v", err)
		}
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		sfCache := &Cache{Cache: &errorCache{}}

		// Test Get with error cache
		_, err := sfCache.Get(context.Background(), "any-key")
		if err == nil {
			t.Error("Expected error, but got nil")
		}
	})
}

// TestSF_Cache_Set tests the Set method of the singleflight cache.
func TestSF_Cache_Set(t *testing.T) {
	// Test successful Set operation
	t.Run("Successful Set", func(t *testing.T) {
		underlying := newMockCache(0)
		sfCache := &Cache{Cache: underlying}

		// Test Set
		key := "test-key"
		value := "test-value"
		err := sfCache.Set(context.Background(), key, value)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify the value was set
		result, err := underlying.Get(context.Background(), key)
		if err != nil {
			t.Errorf("Unexpected error when verifying set value: %v", err)
		}
		if result != value {
			t.Errorf("Expected %v, but got %v", value, result)
		}
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		sfCache := &Cache{Cache: &errorCache{}}

		// Test Set with error cache
		err := sfCache.Set(context.Background(), "any-key", "any-value")
		if err == nil {
			t.Error("Expected error, but got nil")
		}
	})
}

// TestSF_Cache_Delete tests the Delete method of the singleflight cache.
func TestSF_Cache_Delete(t *testing.T) {
	// Test successful Delete operation
	t.Run("Successful Delete", func(t *testing.T) {
		underlying := newMockCache(0)
		sfCache := &Cache{Cache: underlying}

		// Set up test data
		key := "test-key"
		value := "test-value"
		err := underlying.Set(context.Background(), key, value)
		if err != nil {
			t.Fatalf("Failed to set up test data: %v", err)
		}

		// Test Delete
		err = sfCache.Delete(context.Background(), key)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify the value was deleted
		_, err = underlying.Get(context.Background(), key)
		if err == nil {
			t.Error("Expected ErrCacheMiss after deletion, but got no error")
		} else if err != gouache.ErrCacheMiss {
			t.Errorf("Expected ErrCacheMiss after deletion, but got: %v", err)
		}
	})

	// Test deleting non-existent key
	t.Run("Delete Non-existent Key", func(t *testing.T) {
		underlying := newMockCache(0)
		sfCache := &Cache{Cache: underlying}

		// Test Delete with non-existent key
		err := sfCache.Delete(context.Background(), "non-existent-key")
		// Depending on the underlying cache implementation, this might or might not return an error
		// We're primarily checking that the method doesn't panic
		if err != nil {
			// If there's an error, it should be handled gracefully
			t.Logf("Delete returned error (this may be expected): %v", err)
		}
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		sfCache := &Cache{Cache: &errorCache{}}

		// Test Delete with error cache
		err := sfCache.Delete(context.Background(), "any-key")
		if err == nil {
			t.Error("Expected error, but got nil")
		}
	})
}

// TestSF_Cache_Get_Singleflight tests that concurrent Get operations are properly deduplicated.
func TestSF_Cache_Get_Singleflight(t *testing.T) {
	// Create a mock cache with a delay to simulate slow operations
	underlying := newMockCache(100 * time.Millisecond)
	sfCache := &Cache{Cache: underlying}

	// Set up test data
	key := "test-key"
	value := "test-value"
	err := underlying.Set(context.Background(), key, value)
	if err != nil {
		t.Fatalf("Failed to set up test data: %v", err)
	}

	// Number of concurrent goroutines
	goroutines := 10
	results := make(chan any, goroutines)
	errors := make(chan error, goroutines)

	// Start time
	start := time.Now()

	// Launch multiple goroutines that all try to get the same key
	for i := 0; i < goroutines; i++ {
		go func() {
			result, err := sfCache.Get(context.Background(), key)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	// Collect results
	resultCount := 0
	errorCount := 0
	var firstResult any

	for i := 0; i < goroutines; i++ {
		select {
		case result := <-results:
			resultCount++
			if firstResult == nil {
				firstResult = result
			} else if result != firstResult {
				t.Errorf("Inconsistent results: expected %v, but got %v", firstResult, result)
			}
		case err := <-errors:
			errorCount++
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Check that all goroutines completed successfully
	if resultCount != goroutines {
		t.Errorf("Expected %d successful results, but got %d", goroutines, resultCount)
	}
	if errorCount != 0 {
		t.Errorf("Expected 0 errors, but got %d", errorCount)
	}

	// Check that the operation was faster than if each had been executed separately
	// (This is a basic check - in reality, the underlying cache delay should only
	// be experienced once due to singleflight)
	expectedMaxDuration := 200 * time.Millisecond // Allow some overhead
	actualDuration := time.Since(start)
	if actualDuration > expectedMaxDuration {
		t.Errorf("Expected duration to be less than %v, but was %v", expectedMaxDuration, actualDuration)
	}

	// Verify the result is correct
	if firstResult != value {
		t.Errorf("Expected %v, but got %v", value, firstResult)
	}
}
