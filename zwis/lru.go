package zwis

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type LRUCache struct {
	capacity int
	cache    map[interface{}]*list.Element
	list     *list.List
	mutex    sync.RWMutex
}

type entry struct {
	key        interface{}
	value      interface{}
	expiration time.Time
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element),
		list:     list.New(),
	}
}

func (lru *LRUCache) Get(ctx context.Context, key interface{}) (interface{}, bool) {
	lru.mutex.RLock()
	elem, ok := lru.cache[key]
	lru.mutex.RUnlock()

	if !ok {
		return nil, false
	}

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	entry := elem.Value.(*entry)
	if !entry.expiration.IsZero() && entry.expiration.Before(time.Now()) {
		lru.removeElement(elem)
		return nil, false
	}

	lru.list.MoveToFront(elem)
	return entry.value, true
}

func (lru *LRUCache) Set(ctx context.Context, key, value interface{}, ttl time.Duration) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	if elem, ok := lru.cache[key]; ok {
		lru.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
		elem.Value.(*entry).expiration = expiration
	} else {
		if lru.list.Len() >= lru.capacity {
			lru.removeOldest()
		}
		elem := lru.list.PushFront(&entry{key, value, expiration})
		lru.cache[key] = elem
	}
}

func (lru *LRUCache) Delete(ctx context.Context, key interface{}) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if elem, ok := lru.cache[key]; ok {
		lru.removeElement(elem)
	}
}

func (lru *LRUCache) Clear(ctx context.Context) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.list.Init()
	lru.cache = make(map[interface{}]*list.Element)
}

func (lru *LRUCache) removeOldest() {
	oldest := lru.list.Back()
	if oldest != nil {
		lru.removeElement(oldest)
	}
}

func (lru *LRUCache) removeElement(elem *list.Element) {
	lru.list.Remove(elem)
	delete(lru.cache, elem.Value.(*entry).key)
}
