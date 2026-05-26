package main

import (
	"context"
	"fmt"
	"time"
)

// ==================================================
// TUTORIAL 13: CONTEXT TIMEOUT
// ==================================================
//
// PROBLEM STATEMENT:
// Long-running operations (API calls, database queries, file I/O) should
// not run indefinitely. context.WithTimeout creates a context that
// automatically cancels after a specified duration. Any goroutine watching
// ctx.Done() will be notified when the deadline expires.
//
// KEY CONCEPTS:
// - context.WithTimeout(parent, duration) — auto-cancels after duration
// - context.WithDeadline(parent, time) — auto-cancels at a specific time
// - Always call the returned cancel function (defer cancel()) to release
//   resources even if the operation completes before the timeout.
//
// ARCHITECTURE:
//
//   Context (timeout=300ms)
//       │
//       ├── Operation A (takes 100ms) → ✅ completes before timeout
//       ├── Operation B (takes 500ms) → ❌ cancelled by timeout
//       └── ctx.Done() fires at 300ms
//

// simulateWork represents a task that takes a variable amount of time.
// It respects the context timeout and returns early if cancelled.
func simulateWork(ctx context.Context, name string, duration time.Duration) (string, error) {
	fmt.Printf("  [%s] Starting (needs %v)...\n", name, duration)

	select {
	case <-time.After(duration):
		// Work completed before the context deadline
		result := fmt.Sprintf("%s completed successfully in %v", name, duration)
		return result, nil

	case <-ctx.Done():
		// Context timed out or was cancelled before work finished
		return "", fmt.Errorf("%s aborted: %w", name, ctx.Err())
	}
}

// fetchWithTimeout demonstrates a real-world pattern: wrapping a slow
// operation with a timeout to prevent it from blocking forever.
func fetchWithTimeout(timeout time.Duration) {
	// Create a context that will automatically cancel after 'timeout'.
	// IMPORTANT: Always defer cancel() to free resources.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("\n  Timeout set to: %v\n", timeout)
	fmt.Printf("  Deadline: %v\n", func() string {
		if deadline, ok := ctx.Deadline(); ok {
			return deadline.Format("15:04:05.000")
		}
		return "none"
	}())

	result, err := simulateWork(ctx, "DatabaseQuery", 200*time.Millisecond)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}
	fmt.Printf("  ✅ Result: %s\n", result)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 13: CONTEXT TIMEOUT                 ")
	fmt.Println("==================================================")

	// --- Demo 1: Operation completes within timeout ---
	fmt.Println("\n--- Demo 1: Generous Timeout (500ms for 200ms work) ---")
	fetchWithTimeout(500 * time.Millisecond)

	// --- Demo 2: Operation exceeds timeout ---
	fmt.Println("\n--- Demo 2: Tight Timeout (100ms for 200ms work) ---")
	fetchWithTimeout(100 * time.Millisecond)

	// --- Demo 3: Multiple operations with shared timeout ---
	fmt.Println("\n--- Demo 3: Multiple Operations, Shared Timeout ---")

	// All operations share a single 350ms timeout.
	// The total time for ALL operations must fit within 350ms.
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	operations := []struct {
		Name     string
		Duration time.Duration
	}{
		{"FastAPI", 50 * time.Millisecond},
		{"MediumAPI", 100 * time.Millisecond},
		{"SlowAPI", 300 * time.Millisecond}, // This will exceed remaining time
	}

	// Run operations sequentially — they share the same deadline.
	// Earlier operations consume part of the timeout budget.
	for _, op := range operations {
		result, err := simulateWork(ctx, op.Name, op.Duration)
		if err != nil {
			fmt.Printf("  ❌ %v\n", err)
			break // No point continuing if we've timed out
		}
		fmt.Printf("  ✅ %s\n", result)
	}

	// --- Demo 4: Using context.WithDeadline ---
	fmt.Println("\n--- Demo 4: Absolute Deadline ---")
	deadline := time.Now().Add(150 * time.Millisecond)
	ctx2, cancel2 := context.WithDeadline(context.Background(), deadline)
	defer cancel2()

	fmt.Printf("  Deadline set to: %s\n", deadline.Format("15:04:05.000"))
	result, err := simulateWork(ctx2, "ReportGenerator", 300*time.Millisecond)
	if err != nil {
		fmt.Printf("  ❌ %v\n", err)
	} else {
		fmt.Printf("  ✅ %s\n", result)
	}

	fmt.Println("\nContext timeout demo complete!")
	fmt.Println("Tutorial 13 complete!")
}
