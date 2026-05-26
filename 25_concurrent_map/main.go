package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 25: CONCURRENT MAP
// ==================================================
//
// PROBLEM STATEMENT:
// Go's built-in map is NOT safe for concurrent use. If multiple goroutines
// read and write a map simultaneously, the program will panic with
// "concurrent map writes". Two solutions:
//   1. Custom map protected by sync.RWMutex (read-write lock)
//   2. sync.Map (built-in concurrent map, optimized for specific patterns)
//
// sync.RWMutex allows multiple concurrent readers OR one exclusive writer.
// This is optimal for read-heavy workloads.
//

// --- Approach 1: RWMutex-based Concurrent Map ---

// ConcurrentMap is a thread-safe map using RWMutex.
// RWMutex allows concurrent reads but exclusive writes.
type ConcurrentMap[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// NewConcurrentMap creates a new thread-safe map.
func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		items: make(map[K]V),
	}
}

// Set adds or updates a key-value pair (exclusive write lock).
func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	m.mu.Lock() // Write lock — blocks all other readers AND writers
	defer m.mu.Unlock()
	m.items[key] = value
}

// Get retrieves a value by key (shared read lock).
func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock() // Read lock — allows other concurrent readers
	defer m.mu.RUnlock()
	val, ok := m.items[key]
	return val, ok
}

// Delete removes a key (exclusive write lock).
func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

// Len returns the number of items (shared read lock).
func (m *ConcurrentMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// ForEach iterates over all items with a read lock.
func (m *ConcurrentMap[K, V]) ForEach(fn func(K, V)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.items {
		fn(k, v)
	}
}

// GetOrSet returns the existing value for the key, or sets and returns
// the given value if the key doesn't exist. This is atomic.
func (m *ConcurrentMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.items[key]; ok {
		return existing, true // Key already existed
	}
	m.items[key] = value
	return value, false // Key was new
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 25: CONCURRENT MAP                  ")
	fmt.Println("==================================================")

	// --- Demo 1: Custom ConcurrentMap with RWMutex ---
	fmt.Println("\n--- Demo 1: ConcurrentMap[string, int] ---")

	cmap := NewConcurrentMap[string, int]()
	var wg sync.WaitGroup

	// Multiple writers
	fmt.Println("Launching 5 writer goroutines...")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("writer%d-key%d", writerID, j)
				cmap.Set(key, writerID*1000+j)
			}
		}(i)
	}

	// Multiple readers running concurrently with writers
	fmt.Println("Launching 3 reader goroutines...")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			hits := 0
			for j := 0; j < 200; j++ {
				key := fmt.Sprintf("writer%d-key%d", readerID%5, j%100)
				if _, ok := cmap.Get(key); ok {
					hits++
				}
			}
			fmt.Printf("  [Reader %d] Found %d keys.\n", readerID, hits)
		}(i)
	}

	wg.Wait()
	fmt.Printf("  Map size: %d entries\n", cmap.Len())

	// --- Demo 2: sync.Map (standard library) ---
	fmt.Println("\n--- Demo 2: sync.Map ---")
	fmt.Println("sync.Map is optimized for two patterns:")
	fmt.Println("  1. Write-once, read-many (like a cache)")
	fmt.Println("  2. Multiple goroutines read/write disjoint key sets\n")

	var smap sync.Map

	// Concurrent writes to sync.Map
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				smap.Store(fmt.Sprintf("skey-%d-%d", id, j), id*100+j)
			}
		}(i)
	}
	wg.Wait()

	// Count entries using Range
	count := 0
	smap.Range(func(key, value interface{}) bool {
		count++
		return true // Continue iterating
	})
	fmt.Printf("  sync.Map size: %d entries\n", count)

	// LoadOrStore — atomic get-or-set
	actual, loaded := smap.LoadOrStore("unique-key", "first-value")
	fmt.Printf("  LoadOrStore(\"unique-key\"): value=%v, alreadyExisted=%v\n", actual, loaded)

	actual, loaded = smap.LoadOrStore("unique-key", "second-value")
	fmt.Printf("  LoadOrStore(\"unique-key\"): value=%v, alreadyExisted=%v\n", actual, loaded)

	// --- Demo 3: Performance comparison ---
	fmt.Println("\n--- Demo 3: Performance Comparison ---")

	numOps := 100000
	numGoroutines := 10

	// ConcurrentMap benchmark
	cmap2 := NewConcurrentMap[int, int]()
	start := time.Now()
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < numOps/numGoroutines; i++ {
				key := id*numOps + i
				cmap2.Set(key, i)
				cmap2.Get(key)
			}
		}(g)
	}
	wg.Wait()
	cmapDuration := time.Since(start)

	// sync.Map benchmark
	var smap2 sync.Map
	start = time.Now()
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < numOps/numGoroutines; i++ {
				key := id*numOps + i
				smap2.Store(key, i)
				smap2.Load(key)
			}
		}(g)
	}
	wg.Wait()
	smapDuration := time.Since(start)

	fmt.Printf("  ConcurrentMap: %v for %d ops\n", cmapDuration.Round(time.Microsecond), numOps)
	fmt.Printf("  sync.Map:      %v for %d ops\n", smapDuration.Round(time.Microsecond), numOps)

	fmt.Println("\nConcurrent map demo complete!")
	fmt.Println("Tutorial 25 complete!")
}
