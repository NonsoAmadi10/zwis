// Package zwis provides various cache implementations including in-memory,
// LRU (Least Recently Used), LFU (Least Frequently Used), and ARC (Adaptive Replacement Cache).
package zwis

import (
	"context"
	"time"
)

/*
Cache Interface will contain the following key parameters:

Set()
Get()
Delete()
Flush()

*/

// Cache interface defines the methods that all cache implementations must support.
type Cache interface {
	// Set adds an item to the cache, replacing any existing item. If the TTL
	// is 0, the item never expires.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get retrieves an item from the cache. It returns the item and a boolean
	// indicating whether the key was found.
	Get(ctx context.Context, key string) (interface{}, bool)
	// Delete removes the provided key from the cache.
	Delete(ctx context.Context, key string) error
	// Clear removes all items from the cache.
	Flush(ctx context.Context) error
}
