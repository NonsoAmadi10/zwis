package zwis

import (
	"fmt"
)

type CacheType string

const (
	MemoryCacheType CacheType = "memory"
	LRUCacheType    CacheType = "lru"
	LFUCacheType    CacheType = "lfu"
	ARCCacheType    CacheType = "arc"
)

func NewCache(cacheType CacheType, capacity int) (Cache, error) {
	switch cacheType {
	case MemoryCacheType:
		return NewMemoryCache(), nil
	case LRUCacheType:
		return NewLRUCache(capacity), nil
	case LFUCacheType:
		return NewLFUCache(capacity), nil
	case ARCCacheType:
		return NewARCCache(capacity), nil
	default:
		return nil, fmt.Errorf("unknown cache type: %s", cacheType)
	}
}
