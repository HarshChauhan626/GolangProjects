package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 26: LRU CACHE
// ==================================================
//
// PROBLEM STATEMENT:
// An LRU (Least Recently Used) Cache has a fixed capacity. When the cache
// is full and a new item needs to be added, the least recently accessed
// item is evicted. This is widely used for:
//   - Database query caching
//   - Web page caching
//   - DNS resolution caching
//
// IMPLEMENTATION:
// - HashMap (map) for O(1) lookups by key
// - Doubly-Linked List for O(1) insertion/removal and maintaining access order
// - sync.RWMutex for thread safety
//
// DATA STRUCTURE:
//
//   Most Recent ← [A] ↔ [B] ↔ [C] ↔ [D] → Least Recent (evicted first)
//                   ↑
//   map[key] ──────┘   (direct pointer for O(1) access)
//

// entry is an item stored in the cache.
type entry struct {
	key   string
	value interface{}
}

// LRUCache is a thread-safe Least Recently Used cache.
type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element // key → linked list node
	order    *list.List               // front = most recent, back = least recent
	hits     int64
	misses   int64
}

// NewLRUCache creates an LRU cache with the given capacity.
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a value from the cache.
// If found, the item is moved to the front (most recently used).
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// Cache hit — move to front (most recently used)
		c.order.MoveToFront(elem)
		c.hits++
		return elem.Value.(*entry).value, true
	}

	// Cache miss
	c.misses++
	return nil, false
}

// Put adds or updates a key-value pair in the cache.
// If the cache is full, the least recently used item is evicted.
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update value and move to front
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	// If at capacity, evict the least recently used item (back of list)
	if c.order.Len() >= c.capacity {
		c.evict()
	}

	// Add new item at the front
	e := &entry{key: key, value: value}
	elem := c.order.PushFront(e)
	c.items[key] = elem
}

// evict removes the least recently used item (must be called with lock held).
func (c *LRUCache) evict() {
	tail := c.order.Back()
	if tail == nil {
		return
	}

	evicted := tail.Value.(*entry)
	delete(c.items, evicted.key)
	c.order.Remove(tail)
	fmt.Printf("  [Cache] Evicted key '%s'\n", evicted.key)
}

// Delete removes a specific key from the cache.
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		delete(c.items, key)
		c.order.Remove(elem)
		return true
	}
	return false
}

// Len returns the current number of items in the cache.
func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.Len()
}

// Stats returns cache hit/miss statistics.
func (c *LRUCache) Stats() (hits, misses int64, hitRate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := c.hits + c.misses
	if total > 0 {
		hitRate = float64(c.hits) / float64(total) * 100
	}
	return c.hits, c.misses, hitRate
}

// Keys returns all keys in order from most to least recently used.
func (c *LRUCache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	keys := make([]string, 0, c.order.Len())
	for e := c.order.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("           TUTORIAL 26: LRU CACHE                  ")
	fmt.Println("==================================================")

	// --- Demo 1: Basic LRU behavior ---
	fmt.Println("\n--- Demo 1: Basic Operations (capacity=3) ---")
	cache := NewLRUCache(3)

	cache.Put("a", "alpha")
	cache.Put("b", "beta")
	cache.Put("c", "gamma")
	fmt.Printf("  After adding a, b, c: keys=%v\n", cache.Keys())

	// Access 'a' — moves it to front
	val, _ := cache.Get("a")
	fmt.Printf("  Get('a') = %v → keys=%v (a moved to front)\n", val, cache.Keys())

	// Add 'd' — should evict 'b' (least recently used)
	fmt.Print("  Adding 'd': ")
	cache.Put("d", "delta")
	fmt.Printf("keys=%v\n", cache.Keys())

	// Try to get evicted key
	_, found := cache.Get("b")
	fmt.Printf("  Get('b') found=%v (was evicted)\n", found)

	// --- Demo 2: Cache Statistics ---
	fmt.Println("\n--- Demo 2: Cache Statistics ---")
	cache2 := NewLRUCache(5)

	// Populate
	for i := 0; i < 10; i++ {
		cache2.Put(fmt.Sprintf("key%d", i), i*10)
	}

	// Access pattern
	for i := 5; i < 15; i++ {
		cache2.Get(fmt.Sprintf("key%d", i))
	}

	hits, misses, rate := cache2.Stats()
	fmt.Printf("  Hits: %d, Misses: %d, Hit Rate: %.1f%%\n", hits, misses, rate)

	// --- Demo 3: Concurrent Access ---
	fmt.Println("\n--- Demo 3: Concurrent Access ---")
	cache3 := NewLRUCache(100)
	var wg sync.WaitGroup

	start := time.Now()

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cache3.Put(fmt.Sprintf("g%d-k%d", id, j), j)
			}
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				cache3.Get(fmt.Sprintf("g%d-k%d", id%10, j%100))
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	hits, misses, rate = cache3.Stats()
	fmt.Printf("  %d items cached in %v\n", cache3.Len(), elapsed)
	fmt.Printf("  Hits: %d, Misses: %d, Hit Rate: %.1f%%\n", hits, misses, rate)

	fmt.Println("\nLRU Cache demo complete!")
	fmt.Println("Tutorial 26 complete!")
}
