package sample

import (
	"context"
	"fmt"
	"testing"

	"github.com/soyacen/gouache"
)

// TestCache_Get tests the Get method of the Cache implementation.
func TestCache_Get(t *testing.T) {
	// Create a new cache instance
	cache := &Cache{}

	// Define test cases
	tests := []struct {
		name        string
		key         string
		value       any
		setup       bool // whether to set up the value before testing
		expectError bool
	}{
		{
			name:        "Existing key",
			key:         "test-key",
			value:       "test-value",
			setup:       true,
			expectError: false,
		},
		{
			name:        "Non-existing key",
			key:         "non-existent-key",
			value:       nil,
			setup:       false,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: store a value if required
			if tt.setup {
				err := cache.Set(context.Background(), tt.key, tt.value)
				if err != nil {
					t.Fatalf("Failed to set up test value: %v", err)
				}
			}

			// Test: try to get the value
			result, err := cache.Get(context.Background(), tt.key)

			// Check error expectations
			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				} else if err != gouache.ErrCacheMiss {
					t.Errorf("Expected ErrCacheMiss, but got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.value {
					t.Errorf("Expected %v, but got %v", tt.value, result)
				}
			}
		})
	}
}

// TestCache_Set tests the Set method of the Cache implementation.
func TestCache_Set(t *testing.T) {
	// Create a new cache instance
	cache := &Cache{}

	// Define test data
	key := "test-key"
	value := "test-value"

	// Test setting a value
	err := cache.Set(context.Background(), key, value)
	if err != nil {
		t.Errorf("Unexpected error when setting value: %v", err)
	}

	// Verify the value was set correctly
	result, err := cache.Get(context.Background(), key)
	if err != nil {
		t.Errorf("Unexpected error when getting value: %v", err)
	}
	if result != value {
		t.Errorf("Expected %v, but got %v", value, result)
	}
}

// TestCache_Delete tests the Delete method of the Cache implementation.
func TestCache_Delete(t *testing.T) {
	// Create a new cache instance
	cache := &Cache{}

	// Define test data
	key := "test-key"
	value := "test-value"

	// Setup: store a value
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

// TestCache_ConcurrentAccess tests concurrent access to the cache.
func TestCache_ConcurrentAccess(t *testing.T) {
	// Create a new cache instance
	cache := &Cache{}

	// Number of goroutines to run
	goroutines := 100

	// Channel to signal completion
	done := make(chan bool, goroutines)

	// Start multiple goroutines that perform cache operations
	for i := 0; i < goroutines; i++ {
		go func(index int) {
			key := fmt.Sprintf("key-%d", index)
			value := fmt.Sprintf("value-%d", index)

			// Set a value
			_ = cache.Set(context.Background(), key, value)

			// Get the value
			result, err := cache.Get(context.Background(), key)
			if err != nil {
				t.Errorf("Goroutine %d: Unexpected error when getting value: %v", index, err)
			}
			if result != value {
				t.Errorf("Goroutine %d: Expected %v, but got %v", index, value, result)
			}

			// Delete the value
			_ = cache.Delete(context.Background(), key)

			// Signal completion
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}
