package main

import (
	"fmt"
	"sync"
	"time"
)

// A simple worker function that takes a name and a delay
func worker(name string, duration time.Duration, wg *sync.WaitGroup) {
	// Schedule the execution of wg.Done() when the function returns
	defer wg.Done()

	fmt.Printf("[Worker %s] Starting...\n", name)

	// Simulate work by sleeping
	time.Sleep(duration)

	fmt.Printf("[Worker %s] Done!\n", name)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("       TUTORIAL 02: GOROUTINES & WAITGROUPS       ")
	fmt.Println("==================================================")

	// 1. Spawning a goroutine without waiting
	fmt.Println("\n--- 1. Spawning a Goroutine without synchronization ---")
	go func() {
		fmt.Println("--> [Fire & Forget] This might NOT print because the program exits early!")
	}()
	// Sleep for a fraction of a millisecond just to show it *might* or *might not* print
	time.Sleep(10 * time.Microsecond)
	fmt.Println("Main thread continues without waiting.")

	// 2. Using sync.WaitGroup for proper synchronization
	fmt.Println("\n--- 2. Synchronized Execution with sync.WaitGroup ---")
	var wg sync.WaitGroup

	// We have 3 workers to run concurrently
	// Add 3 to the WaitGroup counter
	wg.Add(3)

	fmt.Println("Launching 3 concurrent workers...")

	// Launch worker A, B, and C with different durations
	// We MUST pass WaitGroup by reference (&wg), never by value, because sync.WaitGroup must not be copied!
	go worker("A", 150*time.Millisecond, &wg)
	go worker("B", 50*time.Millisecond, &wg)
	go worker("C", 100*time.Millisecond, &wg)

	fmt.Println("Waiting for all workers to complete...")
	// Wait blocks until the WaitGroup counter goes back to 0
	wg.Wait()
	fmt.Println("All workers finished! Main thread resuming.")

	// 3. Spawning Goroutines in a loop (Best practice check)
	fmt.Println("\n--- 3. Spawning Goroutines in a loop ---")
	var wgLoop sync.WaitGroup

	numTasks := 5
	wgLoop.Add(numTasks)

	fmt.Println("Launching loop workers...")
	for i := 1; i <= numTasks; i++ {
		// Best practice: Pass loop variables as arguments to the goroutine
		// This avoids scoping issues and ensures the goroutine uses the correct value of 'i' at spawn time.
		go func(id int) {
			defer wgLoop.Done()
			fmt.Printf("[Loop Worker %d] Working...\n", id)
			time.Sleep(30 * time.Millisecond)
			fmt.Printf("[Loop Worker %d] Completed.\n", id)
		}(i)
	}

	wgLoop.Wait()
	fmt.Println("\nAll loop tasks finished. Tutorial 02 complete!")
}
