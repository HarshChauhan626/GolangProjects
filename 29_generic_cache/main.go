package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 29: GENERIC CACHE
// ==================================================
//
// PROBLEM STATEMENT:
// Go 1.18+ introduced generics, allowing us to write type-safe, reusable
// data structures. A Generic Cache stores any value type with:
//   - Type safety at compile time (no interface{} casting)
//   - TTL (Time To Live) per entry
//   - Thread safety
//   - Automatic cleanup of expired entries
//
// This is a reusable building block that works with any type:
//   Cache[string]  — cache string values
//   Cache[User]    — cache User structs
//   Cache[[]byte]  — cache byte slices
//

// CacheEntry holds a cached value with its expiration time.
type CacheEntry[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// IsExpired returns true if the entry has passed its TTL.
func (e CacheEntry[V]) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache is a generic, thread-safe cache with TTL support.
type Cache[K comparable, V any] struct {
	mu         sync.RWMutex
	items      map[K]CacheEntry[V]
	defaultTTL time.Duration
	onEvict    func(K, V) // Optional callback when entries are evicted
}

// NewCache creates a new generic cache with the given default TTL.
func NewCache[K comparable, V any](defaultTTL time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items:      make(map[K]CacheEntry[V]),
		defaultTTL: defaultTTL,
	}

	// Start background cleanup goroutine
	go c.cleanupLoop()

	return c
}

// SetEvictionCallback registers a function called when entries expire.
func (c *Cache[K, V]) SetEvictionCallback(fn func(K, V)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = fn
}

// Set stores a value with the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a custom TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheEntry[V]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired, otherwise zero value and false.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	if !ok || entry.IsExpired() {
		var zero V
		return zero, false
	}

	return entry.Value, true
}

// GetOrSet returns the cached value if it exists, or calls the loader
// function, caches the result, and returns it.
func (c *Cache[K, V]) GetOrSet(key K, loader func() (V, error)) (V, error) {
	// Check cache first
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	// Cache miss — call loader
	val, err := loader()
	if err != nil {
		var zero V
		return zero, err
	}

	c.Set(key, val)
	return val, nil
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.items[key]
	if ok {
		delete(c.items, key)
	}
	return ok
}

// Len returns the number of non-expired items in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, entry := range c.items {
		if !entry.IsExpired() {
			count++
		}
	}
	return count
}

// Keys returns all non-expired keys.
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0)
	for k, entry := range c.items {
		if !entry.IsExpired() {
			keys = append(keys, k)
		}
	}
	return keys
}

// cleanupLoop periodically removes expired entries.
func (c *Cache[K, V]) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.items {
			if entry.IsExpired() {
				if c.onEvict != nil {
					c.onEvict(key, entry.Value)
				}
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// --- Domain types for demonstration ---

type Product struct {
	ID    int
	Name  string
	Price float64
}

type Config struct {
	Key   string
	Value string
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 29: GENERIC CACHE                ")
	fmt.Println("==================================================")

	// --- Demo 1: Cache[string, string] ---
	fmt.Println("\n--- Demo 1: String Cache ---")
	stringCache := NewCache[string, string](5 * time.Second)

	stringCache.Set("greeting", "Hello, World!")
	stringCache.Set("language", "Go")

	if val, ok := stringCache.Get("greeting"); ok {
		fmt.Printf("  greeting = %s\n", val)
	}
	fmt.Printf("  Cache size: %d\n", stringCache.Len())

	// --- Demo 2: Cache[int, Product] (typed structs) ---
	fmt.Println("\n--- Demo 2: Product Cache (typed structs) ---")
	productCache := NewCache[int, Product](3 * time.Second)

	productCache.Set(1, Product{1, "Laptop", 999.99})
	productCache.Set(2, Product{2, "Mouse", 29.99})
	productCache.Set(3, Product{3, "Keyboard", 79.99})

	if product, ok := productCache.Get(1); ok {
		// Type-safe — no casting needed!
		fmt.Printf("  Product: %s ($%.2f)\n", product.Name, product.Price)
	}

	// --- Demo 3: GetOrSet (cache-aside built in) ---
	fmt.Println("\n--- Demo 3: GetOrSet (lazy loading) ---")
	configCache := NewCache[string, Config](10 * time.Second)

	// First call — loader is invoked
	cfg, err := configCache.GetOrSet("db.host", func() (Config, error) {
		fmt.Println("  [Loader] Loading config from database...")
		return Config{Key: "db.host", Value: "localhost:5432"}, nil
	})
	if err == nil {
		fmt.Printf("  Config: %s = %s\n", cfg.Key, cfg.Value)
	}

	// Second call — cached, loader NOT invoked
	cfg, err = configCache.GetOrSet("db.host", func() (Config, error) {
		fmt.Println("  [Loader] This should NOT print (cached)!")
		return Config{}, nil
	})
	if err == nil {
		fmt.Printf("  Config (cached): %s = %s\n", cfg.Key, cfg.Value)
	}

	// --- Demo 4: TTL expiration ---
	fmt.Println("\n--- Demo 4: TTL Expiration ---")
	shortCache := NewCache[string, string](500 * time.Millisecond)
	shortCache.SetEvictionCallback(func(key string, val string) {
		fmt.Printf("  [Evicted] key=%s, value=%s\n", key, val)
	})

	shortCache.Set("temp", "this will expire soon")

	if _, ok := shortCache.Get("temp"); ok {
		fmt.Println("  Before expiry: found")
	}

	fmt.Println("  Waiting 1 second for TTL to expire...")
	time.Sleep(1 * time.Second)

	// Force cleanup cycle
	time.Sleep(100 * time.Millisecond)

	if _, ok := shortCache.Get("temp"); !ok {
		fmt.Println("  After expiry: not found (expired)")
	}

	// --- Demo 5: Custom TTL per entry ---
	fmt.Println("\n--- Demo 5: Custom TTL Per Entry ---")
	mixedCache := NewCache[string, string](5 * time.Second)

	mixedCache.SetWithTTL("short", "expires in 300ms", 300*time.Millisecond)
	mixedCache.SetWithTTL("long", "expires in 10s", 10*time.Second)

	fmt.Printf("  Keys before wait: %v\n", mixedCache.Keys())

	time.Sleep(400 * time.Millisecond)

	_, shortOK := mixedCache.Get("short")
	_, longOK := mixedCache.Get("long")
	fmt.Printf("  After 400ms: short found=%v, long found=%v\n", shortOK, longOK)

	// --- Demo 6: Concurrent access ---
	fmt.Println("\n--- Demo 6: Concurrent Access ---")
	concCache := NewCache[int, int](5 * time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				concCache.Set(id*100+j, j*j)
				concCache.Get(id*100 + j)
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("  Concurrent writes complete. Cache size: %d\n", concCache.Len())

	fmt.Println("\nGeneric cache demo complete!")
	fmt.Println("Tutorial 29 complete!")
}
