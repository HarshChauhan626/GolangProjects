package main

import (
	"fmt"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 16: RATE LIMITER (TOKEN BUCKET)
// ==================================================
//
// PROBLEM STATEMENT:
// A rate limiter controls the rate at which operations are performed.
// The Token Bucket algorithm works as follows:
//   - A bucket holds up to N tokens (the burst capacity).
//   - Tokens are added at a fixed rate (e.g., 5 per second).
//   - Each request consumes one token.
//   - If no tokens are available, the request must wait or be rejected.
//
// This is widely used in API gateways, network traffic shaping, and
// preventing abuse of services.
//
// ARCHITECTURE:
//
//   [Refiller goroutine] ──adds tokens──▶ [Bucket (cap=N)]
//
//   Request arrives → try to take token from bucket
//     - Token available → ✅ proceed
//     - No token → ⏳ wait (blocking) or ❌ reject (non-blocking)
//

// TokenBucket implements the token bucket rate limiting algorithm.
type TokenBucket struct {
	tokens     chan struct{} // Channel acts as the bucket; its capacity = burst size
	rate       time.Duration // How often a new token is added
	stopRefill chan struct{} // Signal to stop the refiller goroutine
}

// NewTokenBucket creates a rate limiter with the given burst capacity and
// refill rate. For example, NewTokenBucket(5, 200ms) allows 5 burst
// requests, then limits to 1 request every 200ms.
func NewTokenBucket(burstCapacity int, refillInterval time.Duration) *TokenBucket {
	tb := &TokenBucket{
		tokens:     make(chan struct{}, burstCapacity),
		rate:       refillInterval,
		stopRefill: make(chan struct{}),
	}

	// Fill the bucket to start
	for i := 0; i < burstCapacity; i++ {
		tb.tokens <- struct{}{}
	}

	// Start the refiller goroutine that continuously adds tokens
	go tb.refill()

	return tb
}

// refill continuously adds tokens to the bucket at the configured rate.
// It runs in its own goroutine until Stop() is called.
func (tb *TokenBucket) refill() {
	ticker := time.NewTicker(tb.rate)
	defer ticker.Stop()

	for {
		select {
		case <-tb.stopRefill:
			return

		case <-ticker.C:
			// Try to add a token. If the bucket is full, skip (don't block).
			select {
			case tb.tokens <- struct{}{}:
				// Token added successfully
			default:
				// Bucket is full; discard the token
			}
		}
	}
}

// Allow blocks until a token is available (blocking rate limiter).
// Use this when you want to throttle throughput rather than reject requests.
func (tb *TokenBucket) Allow() {
	<-tb.tokens // Block until a token is available
}

// TryAllow returns true if a token is available (non-blocking).
// Use this when you want to reject excess requests immediately.
func (tb *TokenBucket) TryAllow() bool {
	select {
	case <-tb.tokens:
		return true
	default:
		return false
	}
}

// Stop shuts down the refiller goroutine.
func (tb *TokenBucket) Stop() {
	close(tb.stopRefill)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 16: RATE LIMITER (TOKEN BUCKET)       ")
	fmt.Println("==================================================")

	// --- Demo 1: Blocking Rate Limiter ---
	fmt.Println("\n--- Demo 1: Blocking Rate Limiter ---")
	fmt.Println("Bucket: capacity=3, refill=every 300ms")

	limiter := NewTokenBucket(3, 300*time.Millisecond)

	// First 3 requests should proceed instantly (burst capacity)
	// Subsequent requests should be throttled to ~1 per 300ms
	for i := 1; i <= 8; i++ {
		start := time.Now()
		limiter.Allow() // Blocks until a token is available
		waited := time.Since(start)
		fmt.Printf("  Request %d: allowed (waited %v)\n", i, waited.Round(time.Millisecond))
	}
	limiter.Stop()

	// --- Demo 2: Non-blocking Rate Limiter ---
	fmt.Println("\n--- Demo 2: Non-blocking Rate Limiter ---")
	fmt.Println("Bucket: capacity=2, refill=every 500ms")

	limiter2 := NewTokenBucket(2, 500*time.Millisecond)

	// Rapid-fire 10 requests; only the first 2 should succeed
	fmt.Println("Sending 10 rapid requests:")
	accepted := 0
	rejected := 0
	for i := 1; i <= 10; i++ {
		if limiter2.TryAllow() {
			accepted++
			fmt.Printf("  Request %d: ✅ Allowed\n", i)
		} else {
			rejected++
			fmt.Printf("  Request %d: ❌ Rejected (rate limited)\n", i)
		}
	}
	fmt.Printf("  Total: %d accepted, %d rejected\n", accepted, rejected)

	// Wait for some tokens to refill, then try again
	fmt.Println("\nWaiting 1 second for tokens to refill...")
	time.Sleep(1 * time.Second)

	fmt.Println("Sending 3 more requests after refill:")
	for i := 11; i <= 13; i++ {
		if limiter2.TryAllow() {
			fmt.Printf("  Request %d: ✅ Allowed\n", i)
		} else {
			fmt.Printf("  Request %d: ❌ Rejected\n", i)
		}
	}
	limiter2.Stop()

	// --- Demo 3: Concurrent Access ---
	fmt.Println("\n--- Demo 3: Concurrent Requests ---")
	fmt.Println("Bucket: capacity=5, refill=every 200ms")

	limiter3 := NewTokenBucket(5, 200*time.Millisecond)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 1; i <= 12; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()
			limiter3.Allow() // Each goroutine waits for its token
			elapsed := time.Since(start)
			fmt.Printf("  [Goroutine %2d] Proceeded at %v\n", reqID, elapsed.Round(time.Millisecond))
		}(i)
	}
	wg.Wait()
	limiter3.Stop()

	fmt.Println("\nRate limiter demo complete!")
	fmt.Println("Tutorial 16 complete!")
}
