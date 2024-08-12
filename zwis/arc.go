package zwis

/*
Adaptive Replacement Cache (ARC) is a sophisticated caching algorithm that provides a high hit rate and adapts to varying access patterns. ARC dynamically balances between recent and frequently accessed items by maintaining two lists of pages (recently accessed and frequently accessed) and two ghost lists (recently evicted from each of the main lists).
*/

import (
	"container/list"
	"context"
	"sync"
	"time"
)

// ARCCache implements the Adaptive Replacement Cache algorithm.
// It maintains four lists: T1, T2, B1, and B2.
// T1 and T2 contain cached items, while B1 and B2 contain "ghost" entries (only keys).
type ARCCache struct {
	capacity int                      // Maximum number of items in the cache
	p        int                      // Target size for the T1 list
	t1       *list.List               // List for items accessed once recently
	t2       *list.List               // List for items accessed at least twice recently
	b1       *list.List               // Ghost list for items evicted from T1
	b2       *list.List               // Ghost list for items evicted from T2
	cache    map[string]*list.Element // Map for quick lookup of list elements
	mu       sync.Mutex               // Mutex for thread-safety
}

// arcItem represents an item in the cache.
type arcItem struct {
	key        string
	value      interface{}
	expiration int64 // Unix timestamp for item expiration (0 means no expiration)
}

// NewARCCache creates a new ARC cache with the given capacity.
func NewARCCache(capacity int) *ARCCache {
	return &ARCCache{
		capacity: capacity,
		p:        0,
		t1:       list.New(),
		t2:       list.New(),
		b1:       list.New(),
		b2:       list.New(),
		cache:    make(map[string]*list.Element),
	}
}

// Get retrieves an item from the cache.
func (c *ARCCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elt, ok := c.cache[key]; ok {
		item := elt.Value.(*arcItem)

		if item.expiration > 0 && item.expiration < time.Now().UnixNano() {
			c.remove(key)
			return nil, false
		}

		if c.listContains(c.t1, elt) {
			c.t1.Remove(elt)
			c.t2.PushFront(item)
			c.cache[key] = c.t2.Front()
		} else if c.listContains(c.t2, elt) {
			c.t2.MoveToFront(elt)
		}
		return item.value, true
	}

	// Cache miss, but update ghost lists
	c.request(key)
	return nil, false
}

// Set adds or updates an item in the cache.
// Set adds or updates an item in the cache.
func (c *ARCCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	if elt, ok := c.cache[key]; ok {
		item := elt.Value.(*arcItem)
		item.value = value
		item.expiration = expiration
		if c.listContains(c.t1, elt) {
			c.t1.Remove(elt)
			c.t2.PushFront(item)
			c.cache[key] = c.t2.Front()
		} else if c.listContains(c.t2, elt) {
			c.t2.MoveToFront(elt)
		}
		return nil
	}

	// New item
	c.request(key)

	if c.t1.Len()+c.t2.Len() >= c.capacity {
		c.replace(key)
	}

	item := &arcItem{key: key, value: value, expiration: expiration}
	c.t1.PushFront(item)
	c.cache[key] = c.t1.Front()

	return nil
}

// Delete removes an item from the cache.
func (c *ARCCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.remove(key)
	return nil
}

// Clear removes all items from the cache.
func (c *ARCCache) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.t1.Init()
	c.t2.Init()
	c.b1.Init()
	c.b2.Init()
	c.cache = make(map[string]*list.Element)
	c.p = 0
	return nil
}

// remove deletes an item from the cache and moves it to the appropriate ghost list.
func (c *ARCCache) remove(key string) {
	if elt, ok := c.cache[key]; ok {
		if c.listContains(c.t1, elt) {
			c.t1.Remove(elt)
			c.b1.PushFront(key)
			if c.b1.Len() > c.capacity {
				c.b1.Remove(c.b1.Back())
			}
		} else if c.listContains(c.t2, elt) {
			c.t2.Remove(elt)
			c.b2.PushFront(key)
			if c.b2.Len() > c.capacity {
				c.b2.Remove(c.b2.Back())
			}
		}
		delete(c.cache, key)
	}
}

// replace is called when the cache is full and a new item needs to be added.
// It chooses which item to evict based on the ARC algorithm.
func (c *ARCCache) replace(key string) {
	if c.t1.Len() > 0 && (c.t1.Len() > c.p || (c.listContainsKey(c.b2, key) && c.t1.Len() == c.p)) {
		// Evict from T1
		lru := c.t1.Back()
		c.t1.Remove(lru)
		c.b1.PushFront(lru.Value.(*arcItem).key)
		if c.b1.Len() > c.capacity {
			c.b1.Remove(c.b1.Back())
		}
		delete(c.cache, lru.Value.(*arcItem).key)
	} else {
		// Evict from T2
		lru := c.t2.Back()
		c.t2.Remove(lru)
		c.b2.PushFront(lru.Value.(*arcItem).key)
		if c.b2.Len() > c.capacity {
			c.b2.Remove(c.b2.Back())
		}
		delete(c.cache, lru.Value.(*arcItem).key)
	}
}

// request updates the target size p based on which ghost list contains the requested key.
func (c *ARCCache) request(key string) {
	if c.listContainsKey(c.b1, key) {
		c.p = min(c.capacity, c.p+max(c.b2.Len()/c.b1.Len(), 1))
		c.moveToT2(key)
		item := &arcItem{key: key, value: nil}
		c.t2.PushFront(item)
		c.cache[key] = c.t2.Front()
	} else if c.listContainsKey(c.b2, key) {
		c.p = max(0, c.p-max(c.b1.Len()/c.b2.Len(), 1))
		c.moveToT2(key)
		item := &arcItem{key: key, value: nil}
		c.t2.PushFront(item)
		c.cache[key] = c.t2.Front()
	}
}

func (c *ARCCache) moveToT2(key string) {
	if elt := c.removeFromList(c.b1, key); elt != nil {
		c.b1.Remove(elt)
	} else if elt := c.removeFromList(c.b2, key); elt != nil {
		c.b2.Remove(elt)
	}
}

func (c *ARCCache) removeFromList(l *list.List, key string) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		if k, ok := e.Value.(string); ok && k == key {
			return e
		}
	}
	return nil
}

// listContains checks if a list contains a specific element.
func (c *ARCCache) listContains(l *list.List, element *list.Element) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e == element {
			return true
		}
	}
	return false
}

// listContainsKey checks if a list contains an item with a specific key.
func (c *ARCCache) listContainsKey(l *list.List, key string) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if item, ok := e.Value.(*arcItem); ok && item.key == key {
			return true
		}
		if s, ok := e.Value.(string); ok && s == key {
			return true
		}
	}
	return false
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
