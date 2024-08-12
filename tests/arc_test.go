package zwis_test

import (
	"context"
	"testing"
	"time"

	"github.com/NonsoAmadi10/zwis/zwis"
)

func TestARCCache(t *testing.T) {
	ctx := context.Background()
	cache := zwis.NewARCCache(3)

	// Test Set and Get
	cache.Set(ctx, "key1", "value1", 0)
	cache.Set(ctx, "key2", "value2", 0)
	cache.Set(ctx, "key3", "value3", 0)

	if v, ok := cache.Get(ctx, "key1"); !ok || v != "value1" {
		t.Errorf("Expected value1, got %v", v)
	}

	// Test eviction
	cache.Set(ctx, "key4", "value4", 0)
	if _, ok := cache.Get(ctx, "key2"); ok {
		t.Error("key2 should have been evicted")
	}

	// Test updating existing key
	cache.Set(ctx, "key1", "new_value1", 0)
	if v, ok := cache.Get(ctx, "key1"); !ok || v != "new_value1" {
		t.Errorf("Expected new_value1, got %v", v)
	}

	// Test expiration
	cache.Set(ctx, "key5", "value5", 50*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	if _, ok := cache.Get(ctx, "key5"); ok {
		t.Error("key5 should have expired")
	}

	// Test Delete
	cache.Set(ctx, "key6", "value6", 0)
	cache.Delete(ctx, "key6")
	if _, ok := cache.Get(ctx, "key6"); ok {
		t.Error("key6 should have been deleted")
	}

	// Test Clear
	cache.Set(ctx, "key7", "value7", 0)
	cache.Clear(ctx)
	if _, ok := cache.Get(ctx, "key7"); ok {
		t.Error("Cache should be empty after Clear")
	}
}

func TestARCCacheAdaptiveness(t *testing.T) {
	ctx := context.Background()
	cache := zwis.NewARCCache(5)

	// Fill the cache
	cache.Set(ctx, "A", "A", 0)
	cache.Set(ctx, "B", "B", 0)
	cache.Set(ctx, "C", "C", 0)
	cache.Set(ctx, "D", "D", 0)

	// Access pattern: B, C, D, E
	cache.Get(ctx, "B")
	cache.Get(ctx, "C")
	cache.Get(ctx, "D")
	cache.Set(ctx, "E", "E", 0)

	// // A should be evicted
	// if _, ok := cache.Get(ctx, "A"); ok {
	// 	t.Error("A should have been evicted")
	// }

	// B, C, D, E should still be in the cache
	for _, key := range []string{"B", "C", "D", "E"} {
		if _, ok := cache.Get(ctx, key); !ok {
			t.Errorf("%s should still be in the cache", key)
		}
	}

	// Now, let's access A multiple times
	cache.Set(ctx, "A", "A", 0)
	cache.Get(ctx, "A")
	cache.Get(ctx, "A")

	// Access a new item F
	cache.Set(ctx, "F", "F", 0)

	// B should be evicted now, as it was least recently used among B, C, D, E
	if _, ok := cache.Get(ctx, "B"); ok {
		t.Error("B should have been evicted")
	}

	// A, C, D, E, F should be in the cache
	for _, key := range []string{"A", "C", "D", "E", "F"} {
		if _, ok := cache.Get(ctx, key); !ok {
			t.Errorf("%s should be in the cache", key)
		}
	}
}
