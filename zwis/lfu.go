package zwis

/*
Least Frequently Used (LFU) is a caching algorithm in which the least frequently used cache block is removed whenever the cache is overflowed. In LFU we check the old page as well as the frequency of that page and if the frequency of the page is larger than the old page we cannot remove it and if all the old pages are having same frequency then take last i.e FIFO method for that and remove that page.
*/
import (
	"context"
	"sync"
	"time"
)

type LFUCache struct {
	capacity int
	items    map[string]*lfuItem
	freqs    map[int]*freqNode
	minFreq  int
	mu       sync.Mutex
}

type lfuItem struct {
	key        string
	value      interface{}
	frequency  int
	expiration int64
	freqNode   *freqNode
}

type freqNode struct {
	freq  int
	items map[string]*lfuItem
	prev  *freqNode
	next  *freqNode
}

func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		items:    make(map[string]*lfuItem),
		freqs:    make(map[int]*freqNode),
	}
}

func (c *LFUCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		if item.expiration > 0 && item.expiration < time.Now().UnixNano() {
			c.remove(item)
			return nil, false
		}
		c.incrementFreq(item)
		return item.value, true
	}
	return nil, false
}

func (c *LFUCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	if item, ok := c.items[key]; ok {
		item.value = value
		item.expiration = expiration
		c.incrementFreq(item)
	} else {
		if len(c.items) >= c.capacity {
			c.evict()
		}
		item := &lfuItem{key: key, value: value, frequency: 0, expiration: expiration}
		c.items[key] = item
		c.incrementFreq(item)
	}
	return nil
}

func (c *LFUCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		c.remove(item)
	}
	return nil
}

func (c *LFUCache) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*lfuItem)
	c.freqs = make(map[int]*freqNode)
	c.minFreq = 0
	return nil
}

func (c *LFUCache) incrementFreq(item *lfuItem) {
	if item.freqNode != nil {
		delete(item.freqNode.items, item.key)
		if len(item.freqNode.items) == 0 {
			c.removeFreqNode(item.freqNode)
		}
	}

	item.frequency++
	nextFreq := item.frequency

	if node, ok := c.freqs[nextFreq]; ok {
		node.items[item.key] = item
		item.freqNode = node
	} else {
		node := &freqNode{freq: nextFreq, items: make(map[string]*lfuItem)}
		c.freqs[nextFreq] = node
		c.addFreqNode(node)
		node.items[item.key] = item
		item.freqNode = node
	}

	if item.frequency == 1 {
		c.minFreq = 1
	} else if item.frequency-1 == c.minFreq && len(c.freqs[c.minFreq].items) == 0 {
		c.minFreq++
	}
}

func (c *LFUCache) evict() {
	if node, ok := c.freqs[c.minFreq]; ok {
		for _, item := range node.items {
			c.remove(item)
			break
		}
	}
}

func (c *LFUCache) remove(item *lfuItem) {
	delete(c.items, item.key)
	delete(item.freqNode.items, item.key)
	if len(item.freqNode.items) == 0 {
		c.removeFreqNode(item.freqNode)
	}
}

func (c *LFUCache) removeFreqNode(node *freqNode) {
	delete(c.freqs, node.freq)
	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
}

func (c *LFUCache) addFreqNode(node *freqNode) {
	if prevNode, ok := c.freqs[node.freq-1]; ok {
		node.prev = prevNode
		node.next = prevNode.next
		prevNode.next = node
		if node.next != nil {
			node.next.prev = node
		}
	}
}
