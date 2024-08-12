# Zwis
An in-memory cache system in Go that supports expirable cache entries and various cache eviction policies including Least Frequently Used (LFU), Least Recently Used (LRU), and Adaptive Replacement Cache (ARC).

[![zwis library workflow](https://github.com/NonsoAmadi10/zwis/actions/workflows/main.yaml/badge.svg)](https://github.com/NonsoAmadi10/zwis/actions/workflows/main.yaml)


## Installation
```bash
go get github.com/NonsoAmadi10/zwis
```

## Usage

Here's a quick example of how to use the cache:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/NonsoAmadi10/zwis/zwis"
)

func main() {
    ctx := context.Background()

    // Create an LRU cache
    lruCache, err := zwis.NewCache(cache.LRUCacheType, 100)
    if err != nil {
        panic(err)
    }

    // Set a value
    lruCache.Set(ctx, "key1", "value1", 5*time.Second)

    // Get a value
    if value, ok := lruCache.Get(ctx, "key1"); ok {
        fmt.Printf("key1: %v\n", value)
    }

    // Delete a value
    lruCache.Delete(ctx, "key1")

    // Clear the cache
    lruCache.Clear(ctx)
}
```

## Available Cache Types

* MemoryCache: Simple in-memory cache
* LRUCache: Least Recently Used cache
* LFUCache: Least Frequently Used cache
* ARCCache: Adaptive Replacement Cache
* DiskStore: Implementing a Simple Disk-Backed Cache (coming soon)

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

This project structure provides a solid foundation for a Go cache library. It demonstrates several important Go concepts and best practices:

1. Interface-based design
2. Concurrency handling with mutexes
3. Use of context for cancelation and timeouts
4. Proper error handling
5. Package organization
6. Unit testing
7. Documentation

Areas for potential improvement or expansion:

1. Implement a cache using a diskstore
2. Add benchmarking tests to compare performance of different cache types
3. Implement cache statistics (hit rate, miss rate, etc.)
4. Add support for cache serialization/deserialization for persistence
5. Implement a distributed cache using Redis or similar

Remember, this is a learning exercise, so feel free to experiment with different approaches and optimizations as you build out this library. Good luck with your project!
