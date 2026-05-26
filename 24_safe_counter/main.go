package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ==================================================
// TUTORIAL 24: GOROUTINE-SAFE COUNTER
// ==================================================
//
// PROBLEM STATEMENT:
// A shared counter accessed by multiple goroutines is a classic race
// condition. Without synchronization, concurrent increments can be lost.
// Go provides two primary solutions:
//   1. sync.Mutex — mutual exclusion lock
//   2. sync/atomic — lock-free atomic operations
//
// This tutorial implements both approaches and compares them.
//
// THE RACE CONDITION:
//
//   goroutine A: read counter (value=5)
//   goroutine B: read counter (value=5)   ← reads stale value
//   goroutine A: write counter (value=6)
//   goroutine B: write counter (value=6)  ← overwrites A's increment!
//   Expected: 7, Got: 6 → Lost update!
//

// --- Approach 1: Mutex-based Counter ---

// MutexCounter uses sync.Mutex for thread-safe counting.
type MutexCounter struct {
	mu    sync.Mutex
	value int64
}

// Increment safely adds 1 to the counter.
func (c *MutexCounter) Increment() {
	c.mu.Lock()   // Acquire the lock — other goroutines wait here
	c.value++
	c.mu.Unlock() // Release the lock
}

// Decrement safely subtracts 1 from the counter.
func (c *MutexCounter) Decrement() {
	c.mu.Lock()
	c.value--
	c.mu.Unlock()
}

// Get safely reads the current counter value.
func (c *MutexCounter) Get() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// --- Approach 2: Atomic Counter ---

// AtomicCounter uses sync/atomic for lock-free thread-safe counting.
// Atomic operations are generally faster than mutex for simple counters
// because they don't require OS-level locking.
type AtomicCounter struct {
	value atomic.Int64
}

// Increment atomically adds 1 to the counter.
func (c *AtomicCounter) Increment() {
	c.value.Add(1) // Atomic add — no lock needed
}

// Decrement atomically subtracts 1 from the counter.
func (c *AtomicCounter) Decrement() {
	c.value.Add(-1)
}

// Get atomically reads the current counter value.
func (c *AtomicCounter) Get() int64 {
	return c.value.Load() // Atomic load
}

// --- Unsafe Counter (for demonstrating the race condition) ---

// UnsafeCounter has NO synchronization — will produce incorrect results.
type UnsafeCounter struct {
	value int64
}

func (c *UnsafeCounter) Increment() {
	c.value++ // NOT thread-safe!
}

func (c *UnsafeCounter) Get() int64 {
	return c.value
}

// benchmarkCounter runs N goroutines, each incrementing the counter M times.
func benchmarkCounter(name string, numGoroutines, incrementsPerGoroutine int, increment func(), get func() int64) {
	var wg sync.WaitGroup

	start := time.Now()

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < incrementsPerGoroutine; i++ {
				increment()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	expected := int64(numGoroutines * incrementsPerGoroutine)
	actual := get()
	status := "✅ CORRECT"
	if actual != expected {
		status = "❌ RACE CONDITION"
	}

	fmt.Printf("  %-15s → expected: %d, got: %d — %s (took %v)\n",
		name, expected, actual, status, elapsed.Round(time.Microsecond))
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("   TUTORIAL 24: GOROUTINE-SAFE COUNTER             ")
	fmt.Println("==================================================")

	numGoroutines := 100
	incrementsEach := 10000
	expectedTotal := numGoroutines * incrementsEach

	fmt.Printf("\nRunning %d goroutines × %d increments = %d expected total\n\n",
		numGoroutines, incrementsEach, expectedTotal)

	// --- Test 1: Unsafe Counter (will likely have incorrect result) ---
	fmt.Println("--- Comparing Counter Implementations ---")
	unsafeCounter := &UnsafeCounter{}
	benchmarkCounter("UnsafeCounter",
		numGoroutines, incrementsEach,
		unsafeCounter.Increment, unsafeCounter.Get)

	// --- Test 2: Mutex Counter ---
	mutexCounter := &MutexCounter{}
	benchmarkCounter("MutexCounter",
		numGoroutines, incrementsEach,
		mutexCounter.Increment, mutexCounter.Get)

	// --- Test 3: Atomic Counter ---
	atomicCounter := &AtomicCounter{}
	benchmarkCounter("AtomicCounter",
		numGoroutines, incrementsEach,
		atomicCounter.Increment, atomicCounter.Get)

	// --- Demo: Increment and Decrement ---
	fmt.Println("\n--- Demo: Increment & Decrement ---")
	mc := &MutexCounter{}
	ac := &AtomicCounter{}

	var wg sync.WaitGroup

	// 50 goroutines increment, 30 decrement → net +20
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); mc.Increment(); ac.Increment() }()
	}
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); mc.Decrement(); ac.Decrement() }()
	}

	wg.Wait()
	fmt.Printf("  MutexCounter (50 inc, 30 dec): %d (expected 20)\n", mc.Get())
	fmt.Printf("  AtomicCounter (50 inc, 30 dec): %d (expected 20)\n", ac.Get())

	fmt.Println("\nGoroutine-safe counter demo complete!")
	fmt.Println("Tutorial 24 complete!")
}
