package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NonsoAmadi10/zwis"
)

func main() {
	ctx := context.Background()

	arcCache, err := zwis.NewCache(zwis.ARCCacheType, 100)
	if err != nil {
		panic(err)
	}

	// Set some values
	arcCache.Set(ctx, "key1", "value1", 5*time.Second)
	arcCache.Set(ctx, "key2", "value2", 0) // 0 means no expiration

	// Get values
	if value, ok := arcCache.Get(ctx, "key1"); ok {
		fmt.Printf("key1: %v\n", value)
	}

	// Wait for expiration
	time.Sleep(6 * time.Second)

	// Try to get expired value
	if _, ok := arcCache.Get(ctx, "key1"); !ok {
		fmt.Println("key1 has expired")
	}

	// key2 should still be there
	if value, ok := arcCache.Get(ctx, "key2"); ok {
		fmt.Printf("key2: %v\n", value)
	}

	// Demonstrate adaptiveness
	for i := 0; i < 10; i++ {
		arcCache.Set(ctx, fmt.Sprintf("key%d", i), i, 0)
	}

	// Access some keys multiple times
	for i := 0; i < 5; i++ {
		arcCache.Get(ctx, "key0")
		arcCache.Get(ctx, "key1")
	}

	// Add a new key
	arcCache.Set(ctx, "key10", 10, 0)

	// Check which keys are still in the cache
	for i := 0; i < 11; i++ {
		if value, ok := arcCache.Get(ctx, fmt.Sprintf("key%d", i)); ok {
			fmt.Printf("key%d is still in cache with value %v\n", i, value)
		}
	}
}
