package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ==================================================
// TUTORIAL 18: RETRY MECHANISM WITH EXPONENTIAL BACKOFF
// ==================================================
//
// PROBLEM STATEMENT:
// Transient failures (network timeouts, temporary server errors) are common
// in distributed systems. Instead of failing immediately, a retry mechanism
// re-attempts the operation with increasing delays between retries.
//
// Exponential Backoff: delay = baseDelay * 2^attempt
// With Jitter: adds randomness to prevent "thundering herd" problems where
// many clients retry at exactly the same moment.
//
// ARCHITECTURE:
//
//   Attempt 1 → fail → wait 100ms
//   Attempt 2 → fail → wait 200ms
//   Attempt 3 → fail → wait 400ms
//   Attempt 4 → success ✅
//

// RetryConfig holds the configuration for the retry mechanism.
type RetryConfig struct {
	MaxRetries  int           // Maximum number of retry attempts
	BaseDelay   time.Duration // Initial delay between retries
	MaxDelay    time.Duration // Maximum delay cap (prevents excessively long waits)
	Multiplier  float64       // Multiplier applied to delay after each retry (typically 2.0)
	JitterRatio float64       // Random jitter as a fraction of delay (0.0 to 1.0)
}

// DefaultRetryConfig returns sensible defaults for most use cases.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		JitterRatio: 0.3, // ±30% jitter
	}
}

// calculateDelay computes the delay for a given attempt using exponential backoff + jitter.
func (rc RetryConfig) calculateDelay(attempt int) time.Duration {
	// Exponential: baseDelay * multiplier^attempt
	delay := float64(rc.BaseDelay) * math.Pow(rc.Multiplier, float64(attempt))

	// Cap the delay at MaxDelay
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}

	// Add jitter: delay ± (jitterRatio * delay)
	if rc.JitterRatio > 0 {
		jitter := delay * rc.JitterRatio
		delay = delay - jitter + (rand.Float64() * 2 * jitter)
	}

	return time.Duration(delay)
}

// RetryableFunc is a function that may fail and should be retried.
// It returns an error to indicate failure.
type RetryableFunc func(attempt int) error

// Retry executes the given function with retry logic.
// It returns nil if the function succeeds, or the last error if all retries are exhausted.
func Retry(config RetryConfig, operation RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := config.calculateDelay(attempt - 1)
			fmt.Printf("  ⏳ Waiting %v before retry #%d...\n", delay.Round(time.Millisecond), attempt)
			time.Sleep(delay)
		}

		err := operation(attempt)
		if err == nil {
			if attempt > 0 {
				fmt.Printf("  ✅ Succeeded on attempt #%d\n", attempt+1)
			}
			return nil // Success!
		}

		lastErr = err
		fmt.Printf("  ❌ Attempt #%d failed: %v\n", attempt+1, err)
	}

	return fmt.Errorf("all %d attempts failed. Last error: %w", config.MaxRetries+1, lastErr)
}

// --- Simulated operations for demonstration ---

// flakyAPICall simulates an API that fails some percentage of the time.
func flakyAPICall(failRate float64) RetryableFunc {
	return func(attempt int) error {
		if rand.Float64() < failRate {
			return errors.New("connection refused")
		}
		return nil
	}
}

// failNTimes returns an operation that fails exactly N times, then succeeds.
func failNTimes(n int) RetryableFunc {
	count := 0
	return func(attempt int) error {
		count++
		if count <= n {
			return fmt.Errorf("transient error (failure %d/%d)", count, n)
		}
		return nil
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("  TUTORIAL 18: RETRY WITH EXPONENTIAL BACKOFF      ")
	fmt.Println("==================================================")

	config := DefaultRetryConfig()
	fmt.Printf("Config: MaxRetries=%d, BaseDelay=%v, Multiplier=%.1f, MaxDelay=%v\n\n",
		config.MaxRetries, config.BaseDelay, config.Multiplier, config.MaxDelay)

	// --- Demo 1: Operation that fails 3 times then succeeds ---
	fmt.Println("--- Demo 1: Fail 3 times, then succeed ---")
	err := Retry(config, failNTimes(3))
	if err != nil {
		fmt.Printf("  Final result: FAILED — %v\n", err)
	} else {
		fmt.Println("  Final result: SUCCESS")
	}

	// --- Demo 2: Operation that always fails (exhausts retries) ---
	fmt.Println("\n--- Demo 2: Always fails (exhausts all retries) ---")
	err = Retry(config, func(attempt int) error {
		return errors.New("server is down")
	})
	if err != nil {
		fmt.Printf("  Final result: FAILED — %v\n", err)
	}

	// --- Demo 3: Flaky API with ~50% failure rate ---
	fmt.Println("\n--- Demo 3: Flaky API (~50% failure rate) ---")
	err = Retry(config, flakyAPICall(0.5))
	if err != nil {
		fmt.Printf("  Final result: FAILED — %v\n", err)
	} else {
		fmt.Println("  Final result: SUCCESS")
	}

	// --- Demo 4: Show backoff delay progression ---
	fmt.Println("\n--- Demo 4: Delay Progression ---")
	for i := 0; i < 8; i++ {
		delay := config.calculateDelay(i)
		fmt.Printf("  Attempt %d → delay: %v\n", i+1, delay.Round(time.Millisecond))
	}

	fmt.Println("\nRetry mechanism demo complete!")
	fmt.Println("Tutorial 18 complete!")
}
