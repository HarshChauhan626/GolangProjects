package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 12: CONTEXT CANCELLATION
// ==================================================
//
// PROBLEM STATEMENT:
// When a parent operation is cancelled (e.g., user navigates away, request
// is no longer needed), all child goroutines should stop immediately to
// free resources. Go's context.Context provides a standard way to propagate
// cancellation signals through the call tree.
//
// KEY CONCEPTS:
// - context.WithCancel creates a child context with a cancel function.
// - Calling cancel() broadcasts a signal to ALL goroutines watching ctx.Done().
// - Every long-running goroutine should select on ctx.Done() to respond
//   to cancellation promptly.
//
// ARCHITECTURE:
//
//   Parent Context
//       │
//       ├── cancel() called
//       │
//       ├──▶ Worker 1 (watches ctx.Done()) → stops
//       ├──▶ Worker 2 (watches ctx.Done()) → stops
//       └──▶ Worker 3 (watches ctx.Done()) → stops
//

// longRunningTask simulates a worker that does iterative work.
// It checks for cancellation on each iteration via ctx.Done().
func longRunningTask(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			// ctx.Done() is closed when cancel() is called or the context expires.
			// ctx.Err() tells us WHY: context.Canceled or context.DeadlineExceeded.
			fmt.Printf("  [Worker %d] Cancelled after %d iterations. Reason: %v\n",
				id, i-1, ctx.Err())
			return

		default:
			// Do a unit of work
			time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
			fmt.Printf("  [Worker %d] Completed iteration %d\n", id, i)
		}
	}
}

// searchDatabase simulates a database query that respects cancellation.
func searchDatabase(ctx context.Context, query string) (string, error) {
	// Simulate a slow database query
	select {
	case <-time.After(500 * time.Millisecond):
		return fmt.Sprintf("Results for '%s': 42 rows found", query), nil
	case <-ctx.Done():
		return "", fmt.Errorf("database search cancelled: %w", ctx.Err())
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 12: CONTEXT CANCELLATION              ")
	fmt.Println("==================================================")

	// --- Demo 1: Cancel multiple workers ---
	fmt.Println("\n--- Demo 1: Cancelling Multiple Workers ---")

	// Create a cancellable context.
	// cancel() is the function we call to trigger cancellation.
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	numWorkers := 3

	// Launch workers that will run until cancelled
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go longRunningTask(ctx, i, &wg)
	}

	// Let workers run for a bit, then cancel them all
	time.Sleep(400 * time.Millisecond)
	fmt.Println("\n  >>> Calling cancel() — all workers should stop <<<")
	cancel() // This broadcasts cancellation to all workers

	// Wait for all workers to acknowledge cancellation and exit
	wg.Wait()
	fmt.Println("  All workers stopped gracefully.")

	// --- Demo 2: Cancel a slow operation ---
	fmt.Println("\n--- Demo 2: Cancelling a Slow Database Query ---")

	// Create a new context for this demo
	ctx2, cancel2 := context.WithCancel(context.Background())

	// Start a database search in a goroutine
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := searchDatabase(ctx2, "SELECT * FROM users")
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	// Cancel after 200ms (before the 500ms query completes)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("  >>> Cancelling database query after 200ms <<<")
	cancel2()

	// Check the result
	select {
	case result := <-resultCh:
		fmt.Printf("  Got result: %s\n", result)
	case err := <-errCh:
		fmt.Printf("  Got error (expected): %v\n", err)
	}

	fmt.Println("\nContext cancellation demo complete!")
	fmt.Println("Tutorial 12 complete!")
}
