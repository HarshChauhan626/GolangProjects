package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 11: CONCURRENT API CALLS
// ==================================================
//
// PROBLEM STATEMENT:
// When your application needs to call multiple independent APIs (or services),
// calling them sequentially wastes time. Instead, fire all requests in
// parallel and aggregate results once all have responded.
//
// This pattern uses sync.WaitGroup (or errgroup) to wait for all goroutines,
// and a thread-safe mechanism to collect results.
//
// ARCHITECTURE:
//
//   main ──┬──▶ goroutine → call API 1 → result ──┐
//          ├──▶ goroutine → call API 2 → result ──┼──▶ Aggregate
//          └──▶ goroutine → call API 3 → result ──┘
//

// APIResponse simulates a response from an external API.
type APIResponse struct {
	Source   string
	Data     string
	Latency  time.Duration
	Error    error
}

// callAPI simulates an HTTP call to an external service.
// In production, you would use net/http here.
func callAPI(name string, latency time.Duration) APIResponse {
	start := time.Now()

	// Simulate network latency
	time.Sleep(latency)

	// Simulate occasional failure (10% chance)
	if rand.Float64() < 0.1 {
		return APIResponse{
			Source:  name,
			Error:   fmt.Errorf("service %s: connection timeout", name),
			Latency: time.Since(start),
		}
	}

	return APIResponse{
		Source:  name,
		Data:    fmt.Sprintf("Response from %s at %v", name, time.Now().Format("15:04:05.000")),
		Latency: time.Since(start),
	}
}

// fetchAll calls multiple APIs concurrently and collects all results.
// It uses a mutex-protected slice to safely aggregate results from goroutines.
func fetchAll(apis map[string]time.Duration) []APIResponse {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []APIResponse
	)

	for name, latency := range apis {
		wg.Add(1)
		// Launch each API call in its own goroutine
		go func(apiName string, apiLatency time.Duration) {
			defer wg.Done()

			fmt.Printf("[Calling] %s...\n", apiName)
			resp := callAPI(apiName, apiLatency)

			// Safely append result using mutex
			mu.Lock()
			results = append(results, resp)
			mu.Unlock()
		}(name, latency)
	}

	// Block until all API calls have returned
	wg.Wait()
	return results
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 11: CONCURRENT API CALLS              ")
	fmt.Println("==================================================")

	// Define the APIs we need to call with their simulated latencies
	apis := map[string]time.Duration{
		"UserService":    200 * time.Millisecond,
		"OrderService":   350 * time.Millisecond,
		"PaymentService": 150 * time.Millisecond,
		"InventoryAPI":   400 * time.Millisecond,
		"NotificationSvc": 100 * time.Millisecond,
	}

	// --- Sequential Approach (for comparison) ---
	fmt.Println("\n--- Sequential Execution ---")
	seqStart := time.Now()
	for name, latency := range apis {
		callAPI(name, latency)
	}
	seqDuration := time.Since(seqStart)
	fmt.Printf("Sequential total time: %v\n", seqDuration)

	// --- Concurrent Approach ---
	fmt.Println("\n--- Concurrent Execution ---")
	concStart := time.Now()
	results := fetchAll(apis)
	concDuration := time.Since(concStart)

	fmt.Println("\n--- Results ---")
	for _, r := range results {
		if r.Error != nil {
			fmt.Printf("  ❌ %s: ERROR — %v (took %v)\n", r.Source, r.Error, r.Latency)
		} else {
			fmt.Printf("  ✅ %s: %s (took %v)\n", r.Source, r.Data, r.Latency)
		}
	}

	fmt.Printf("\nConcurrent total time: %v\n", concDuration)
	fmt.Printf("Speedup: %.1fx faster than sequential\n",
		float64(seqDuration)/float64(concDuration))
	fmt.Println("\nTutorial 11 complete!")
}
