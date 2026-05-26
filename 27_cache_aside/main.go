package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 27: CACHE-ASIDE PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// The Cache-Aside (or Lazy Loading) pattern is a caching strategy where:
//   1. The application first checks the cache for data.
//   2. On a cache hit, return the cached data.
//   3. On a cache miss, load from the database, store in cache, return.
//   4. On writes, update the database first, then invalidate the cache.
//
// This separates caching concerns from the database, giving the app
// full control over what gets cached and when.
//
// FLOW:
//
//   Read:  App → Cache? → Hit → return data
//                       → Miss → DB → store in cache → return data
//
//   Write: App → Update DB → Invalidate cache entry
//

// User represents a domain model.
type User struct {
	ID        int
	Name      string
	Email     string
	UpdatedAt time.Time
}

// --- Simulated Database ---

// Database simulates a slow persistent storage layer.
type Database struct {
	mu      sync.Mutex
	data    map[int]User
	queries int // Track number of DB queries for demonstration
}

// NewDatabase creates a simulated database with seed data.
func NewDatabase() *Database {
	db := &Database{
		data: make(map[int]User),
	}
	// Seed data
	db.data[1] = User{1, "Alice", "alice@example.com", time.Now()}
	db.data[2] = User{2, "Bob", "bob@example.com", time.Now()}
	db.data[3] = User{3, "Charlie", "charlie@example.com", time.Now()}
	return db
}

// GetUser simulates a slow database read.
func (db *Database) GetUser(id int) (User, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.queries++

	// Simulate database latency
	time.Sleep(50 * time.Millisecond)

	user, ok := db.data[id]
	return user, ok
}

// UpdateUser simulates a database write.
func (db *Database) UpdateUser(user User) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.queries++

	time.Sleep(30 * time.Millisecond) // Simulate write latency
	user.UpdatedAt = time.Now()
	db.data[user.ID] = user
}

// QueryCount returns the total number of DB queries made.
func (db *Database) QueryCount() int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.queries
}

// --- Cache Layer ---

// Cache is a simple in-memory cache with TTL support.
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
	hits  int
	misses int
}

// cacheEntry wraps a cached value with its expiration time.
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewCache creates a cache with the given TTL for entries.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// Get retrieves a value from the cache. Returns false if not found or expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	c.mu.RUnlock()
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	c.mu.RLock()
	return entry.value, true
}

// Set stores a value in the cache with the configured TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Delete removes a value from the cache (invalidation).
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Stats returns cache hit/miss stats.
func (c *Cache) Stats() (hits, misses int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// --- Cache-Aside Service ---

// UserService implements the cache-aside pattern.
type UserService struct {
	db    *Database
	cache *Cache
}

// NewUserService creates a service with cache-aside behavior.
func NewUserService(db *Database, cache *Cache) *UserService {
	return &UserService{db: db, cache: cache}
}

// GetUser implements cache-aside read:
//  1. Check cache
//  2. If miss, read from DB
//  3. Store in cache
//  4. Return
func (s *UserService) GetUser(id int) (User, bool) {
	cacheKey := fmt.Sprintf("user:%d", id)

	// Step 1: Check cache
	if cached, ok := s.cache.Get(cacheKey); ok {
		fmt.Printf("  [Cache HIT] User %d found in cache.\n", id)
		return cached.(User), true
	}

	// Step 2: Cache miss — load from database
	fmt.Printf("  [Cache MISS] User %d not in cache. Querying database...\n", id)
	user, ok := s.db.GetUser(id)
	if !ok {
		return User{}, false
	}

	// Step 3: Store in cache for future reads
	s.cache.Set(cacheKey, user)
	fmt.Printf("  [Cache SET] User %d stored in cache.\n", id)

	return user, true
}

// UpdateUser implements cache-aside write:
//  1. Update database
//  2. Invalidate cache entry
func (s *UserService) UpdateUser(user User) {
	cacheKey := fmt.Sprintf("user:%d", user.ID)

	// Step 1: Update database first
	s.db.UpdateUser(user)
	fmt.Printf("  [DB] User %d updated in database.\n", user.ID)

	// Step 2: Invalidate cache (not update — avoids stale data race)
	s.cache.Delete(cacheKey)
	fmt.Printf("  [Cache INVALIDATE] User %d removed from cache.\n", user.ID)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("     TUTORIAL 27: CACHE-ASIDE PATTERN              ")
	fmt.Println("==================================================")

	db := NewDatabase()
	cache := NewCache(5 * time.Second)
	service := NewUserService(db, cache)

	// --- Demo 1: First read = cache miss, second read = cache hit ---
	fmt.Println("\n--- Demo 1: Read Pattern ---")

	start := time.Now()
	user, _ := service.GetUser(1) // Cache miss → DB query
	fmt.Printf("  Got: %+v (took %v)\n", user, time.Since(start).Round(time.Millisecond))

	start = time.Now()
	user, _ = service.GetUser(1) // Cache hit → no DB query
	fmt.Printf("  Got: %+v (took %v)\n", user, time.Since(start).Round(time.Millisecond))

	// --- Demo 2: Write invalidation ---
	fmt.Println("\n--- Demo 2: Write Invalidation ---")

	fmt.Println("  Updating user 1's email...")
	service.UpdateUser(User{ID: 1, Name: "Alice", Email: "alice.new@example.com"})

	fmt.Println("  Reading user 1 again (should be cache miss):")
	start = time.Now()
	user, _ = service.GetUser(1) // Cache miss (was invalidated)
	fmt.Printf("  Got: %+v (took %v)\n", user, time.Since(start).Round(time.Millisecond))

	// --- Demo 3: Multiple users ---
	fmt.Println("\n--- Demo 3: Multiple Users ---")
	for round := 1; round <= 2; round++ {
		fmt.Printf("  Round %d:\n", round)
		for id := 1; id <= 3; id++ {
			user, found := service.GetUser(id)
			if found {
				fmt.Printf("    User %d: %s (%s)\n", id, user.Name, user.Email)
			}
		}
	}

	// --- Stats ---
	fmt.Println("\n--- Stats ---")
	hits, misses := cache.Stats()
	fmt.Printf("  Cache hits:   %d\n", hits)
	fmt.Printf("  Cache misses: %d\n", misses)
	fmt.Printf("  DB queries:   %d\n", db.QueryCount())
	if hits+misses > 0 {
		fmt.Printf("  Hit rate:     %.1f%%\n", float64(hits)/float64(hits+misses)*100)
	}

	fmt.Println("\nCache-Aside pattern demo complete!")
	fmt.Println("Tutorial 27 complete!")
}
