package zwis_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/NonsoAmadi10/zwis/zwis"
)

func TestLRUCacheConcurrency(t *testing.T) {
	cache := zwis.NewLRUCache(100)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent writes
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			cache.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
		}
	}()

	// Concurrent reads
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			cache.Get(ctx, fmt.Sprintf("key%d", i))
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	// Additional verification
	for i := 900; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value, ok := cache.Get(ctx, key)
		if !ok {
			t.Errorf("Expected key %s to be in cache, but it wasn't", key)
		} else if value != i {
			t.Errorf("Expected value for key %s to be %d, but got %v", key, i, value)
		}
	}
}
