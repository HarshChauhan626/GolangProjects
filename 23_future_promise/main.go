package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 23: FUTURE / PROMISE PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// A Future (or Promise) represents a value that will be available at some
// point in the future. You start an asynchronous computation and receive
// a handle (the Future) immediately. Later, you can retrieve the result
// by calling Get(), which blocks until the value is ready.
//
// This pattern cleanly separates the initiation of work from the
// consumption of its result.
//
// KEY CONCEPTS:
// - Async() starts work and returns a Future immediately
// - Future.Get() blocks until the result is ready, then returns it
// - The result is computed exactly once (memoized)
// - Thread-safe: multiple goroutines can call Get() concurrently
//
// ARCHITECTURE:
//
//   caller → Async(fn) → returns Future immediately
//                              │
//                         [goroutine runs fn()]
//                              │
//   caller → future.Get() → blocks → returns (value, error)
//

// Future represents the result of an asynchronous computation.
// It is generic over the result type T.
type Future[T any] struct {
	result T
	err    error
	done   chan struct{} // Closed when the result is ready
	once   sync.Once    // Ensures the result is set exactly once
}

// Async starts an asynchronous computation and returns a Future.
// The provided function runs in a new goroutine.
func Async[T any](fn func() (T, error)) *Future[T] {
	f := &Future[T]{
		done: make(chan struct{}),
	}

	go func() {
		result, err := fn()
		f.once.Do(func() {
			f.result = result
			f.err = err
			close(f.done) // Signal that the result is ready
		})
	}()

	return f
}

// Get blocks until the result is ready, then returns it.
// Multiple goroutines can safely call Get() concurrently.
func (f *Future[T]) Get() (T, error) {
	<-f.done // Block until the channel is closed
	return f.result, f.err
}

// GetWithTimeout blocks until the result is ready or the timeout expires.
func (f *Future[T]) GetWithTimeout(timeout time.Duration) (T, error) {
	select {
	case <-f.done:
		return f.result, f.err
	case <-time.After(timeout):
		var zero T
		return zero, errors.New("future: timeout waiting for result")
	}
}

// IsReady returns true if the result is available (non-blocking).
func (f *Future[T]) IsReady() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Then chains a transformation on the future's result, returning a new Future.
func Then[T any, U any](f *Future[T], transform func(T) (U, error)) *Future[U] {
	return Async(func() (U, error) {
		result, err := f.Get()
		if err != nil {
			var zero U
			return zero, err
		}
		return transform(result)
	})
}

// AwaitAll waits for multiple futures and returns all results.
func AwaitAll[T any](futures ...*Future[T]) ([]T, []error) {
	results := make([]T, len(futures))
	errs := make([]error, len(futures))

	var wg sync.WaitGroup
	for i, f := range futures {
		wg.Add(1)
		go func(idx int, future *Future[T]) {
			defer wg.Done()
			results[idx], errs[idx] = future.Get()
		}(i, f)
	}
	wg.Wait()

	return results, errs
}

// --- Simulated async operations ---

func fetchUserProfile(userID int) *Future[string] {
	return Async(func() (string, error) {
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		return fmt.Sprintf("User{id=%d, name=Alice}", userID), nil
	})
}

func fetchOrderHistory(userID int) *Future[[]string] {
	return Async(func() ([]string, error) {
		time.Sleep(time.Duration(150+rand.Intn(250)) * time.Millisecond)
		return []string{"Order#1001", "Order#1002", "Order#1003"}, nil
	})
}

func fetchRecommendations(userID int) *Future[[]string] {
	return Async(func() ([]string, error) {
		time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)
		if rand.Float64() < 0.3 {
			return nil, errors.New("recommendation service unavailable")
		}
		return []string{"Product-A", "Product-B"}, nil
	})
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 23: FUTURE / PROMISE PATTERN          ")
	fmt.Println("==================================================")

	// --- Demo 1: Basic Future ---
	fmt.Println("\n--- Demo 1: Basic Future ---")
	future := Async(func() (int, error) {
		time.Sleep(200 * time.Millisecond)
		return 42, nil
	})

	fmt.Printf("  Future created. Is ready? %v\n", future.IsReady())
	result, err := future.Get()
	fmt.Printf("  Result: %d, Error: %v, Is ready now? %v\n", result, err, future.IsReady())

	// --- Demo 2: Parallel API calls with Futures ---
	fmt.Println("\n--- Demo 2: Parallel API Calls ---")
	start := time.Now()

	// Fire all requests concurrently — each returns a Future immediately
	profileFuture := fetchUserProfile(123)
	ordersFuture := fetchOrderHistory(123)
	recosFuture := fetchRecommendations(123)

	fmt.Println("  All futures created (non-blocking). Now awaiting results...")

	// Block on each result as needed
	profile, err := profileFuture.Get()
	fmt.Printf("  Profile: %s (err: %v)\n", profile, err)

	orders, err := ordersFuture.Get()
	fmt.Printf("  Orders: %v (err: %v)\n", orders, err)

	recos, err := recosFuture.Get()
	fmt.Printf("  Recommendations: %v (err: %v)\n", recos, err)

	fmt.Printf("  Total time: %v (all ran in parallel)\n", time.Since(start).Round(time.Millisecond))

	// --- Demo 3: Future chaining with Then ---
	fmt.Println("\n--- Demo 3: Future Chaining (Then) ---")
	lengthFuture := Then(fetchUserProfile(456), func(profile string) (int, error) {
		return len(profile), nil
	})
	length, err := lengthFuture.Get()
	fmt.Printf("  Profile length: %d (err: %v)\n", length, err)

	// --- Demo 4: AwaitAll ---
	fmt.Println("\n--- Demo 4: AwaitAll ---")
	futures := make([]*Future[string], 5)
	for i := range futures {
		idx := i
		futures[i] = Async(func() (string, error) {
			time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)
			return fmt.Sprintf("result-%d", idx), nil
		})
	}

	results, errs := AwaitAll(futures...)
	for i, r := range results {
		fmt.Printf("  Future %d: %s (err: %v)\n", i, r, errs[i])
	}

	// --- Demo 5: Timeout ---
	fmt.Println("\n--- Demo 5: GetWithTimeout ---")
	slowFuture := Async(func() (string, error) {
		time.Sleep(2 * time.Second)
		return "slow result", nil
	})

	val, err := slowFuture.GetWithTimeout(100 * time.Millisecond)
	fmt.Printf("  Result with 100ms timeout: '%s' (err: %v)\n", val, err)

	fmt.Println("\nFuture/Promise pattern demo complete!")
	fmt.Println("Tutorial 23 complete!")
}
