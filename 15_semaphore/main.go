package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 15: SEMAPHORE USING CHANNELS
// ==================================================
//
// PROBLEM STATEMENT:
// A semaphore limits the number of goroutines that can access a resource
// or perform an action concurrently. Unlike a mutex (which allows exactly
// 1), a counting semaphore allows up to N concurrent accesses.
//
// Go doesn't have a built-in semaphore type, but a buffered channel of
// capacity N provides exactly this behavior:
//   - Sending to the channel = acquiring the semaphore (blocks if full)
//   - Receiving from the channel = releasing the semaphore
//
// USE CASES:
// - Limiting concurrent database connections
// - Throttling outbound HTTP requests
// - Controlling file descriptor usage
//
// ARCHITECTURE:
//
//   sem := make(chan struct{}, N)  // N = max concurrent goroutines
//
//   sem <- struct{}{}   // Acquire (blocks if N goroutines already running)
//   // ... do work ...
//   <-sem               // Release (allows another goroutine to proceed)
//

// Semaphore is a channel-based counting semaphore.
type Semaphore chan struct{}

// NewSemaphore creates a semaphore with the given maximum concurrency.
func NewSemaphore(maxConcurrent int) Semaphore {
	return make(Semaphore, maxConcurrent)
}

// Acquire blocks until a slot is available.
func (s Semaphore) Acquire() {
	s <- struct{}{} // Send an empty struct to occupy a slot
}

// Release frees a slot for another goroutine.
func (s Semaphore) Release() {
	<-s // Receive to free the slot
}

// simulateAPICall represents an expensive operation we want to throttle.
func simulateAPICall(id int, sem Semaphore, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[Task %2d] Waiting to acquire semaphore...\n", id)

	sem.Acquire()
	startTime := time.Now()
	fmt.Printf("[Task %2d] ✅ Acquired! Starting work.\n", id)

	// Simulate work
	duration := time.Duration(100+rand.Intn(300)) * time.Millisecond
	time.Sleep(duration)

	sem.Release()
	elapsed := time.Since(startTime)
	fmt.Printf("[Task %2d] Released after %v.\n", id, elapsed)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 15: SEMAPHORE USING CHANNELS          ")
	fmt.Println("==================================================")

	maxConcurrent := 3
	totalTasks := 10

	fmt.Printf("Semaphore capacity: %d (max %d goroutines run simultaneously)\n",
		maxConcurrent, maxConcurrent)
	fmt.Printf("Total tasks: %d\n\n", totalTasks)

	sem := NewSemaphore(maxConcurrent)
	var wg sync.WaitGroup

	start := time.Now()

	// Launch all tasks — the semaphore ensures only N run at a time.
	for i := 1; i <= totalTasks; i++ {
		wg.Add(1)
		go simulateAPICall(i, sem, &wg)
	}

	// Wait for all tasks to finish
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nAll %d tasks completed in %v.\n", totalTasks, elapsed)
	fmt.Printf("Without semaphore they'd all run at once; with semaphore,\n")
	fmt.Printf("only %d ran concurrently at any given time.\n", maxConcurrent)
	fmt.Println("\nTutorial 15 complete!")
}
