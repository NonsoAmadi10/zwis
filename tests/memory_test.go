package zwis_test

import (
	"context"
	"testing"
	"time"

	"github.com/NonsoAmadi10/zwis/zwis"
)

func TestMemoryCache(t *testing.T) {
	cache := zwis.NewMemoryCache()
	ctx := context.Background()

	// Test Set and Get with TTL
	cache.Set(ctx, "key1", 12, 100*time.Millisecond)
	cache.Set(ctx, "key2", 10, 50*time.Millisecond)

	// Test immediate retrieval
	if v, ok := cache.Get(ctx, "key1"); !ok || v != 12 {
		t.Errorf("Expected 12, got %v", v)
	}

	// Test expiration after waiting
	time.Sleep(60 * time.Millisecond)
	if _, ok := cache.Get(ctx, "key2"); ok {
		t.Error("Expected key2 to be expired")
	}

	// Test Delete
	cache.Set(ctx, "key5", "value5", 0)
	cache.Delete(ctx, "key5")
	if _, ok := cache.Get(ctx, "key5"); ok {
		t.Error("key5 should have been deleted")
	}
}
