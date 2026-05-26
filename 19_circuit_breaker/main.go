package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 19: CIRCUIT BREAKER
// ==================================================
//
// PROBLEM STATEMENT:
// When a downstream service is failing, continuing to send requests wastes
// resources and can cause cascading failures. A Circuit Breaker detects
// failures and "trips open" to prevent further calls, giving the service
// time to recover.
//
// STATE MACHINE:
//
//   ┌──────────┐   failures >= threshold   ┌──────────┐
//   │  CLOSED  │ ─────────────────────────▶ │   OPEN   │
//   │(normal)  │                            │(blocking)│
//   └──────────┘                            └────┬─────┘
//        ▲                                       │
//        │ success                          timeout elapsed
//        │                                       │
//   ┌────┴──────┐                           ┌────▼─────┐
//   │  CLOSED   │ ◀───── success ────────── │HALF-OPEN │
//   └───────────┘                           │(testing) │
//                  ─────── failure ────────▶ └──────────┘
//                           (back to OPEN)
//

// State represents the circuit breaker's current state.
type State int

const (
	StateClosed   State = iota // Normal operation — requests pass through
	StateOpen                  // Tripped — requests are blocked immediately
	StateHalfOpen              // Testing — one request allowed to test recovery
)

// String returns a human-readable name for the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker tracks failure rates and controls access to a service.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failureCount     int
	successCount     int
	failureThreshold int           // Number of consecutive failures before opening
	successThreshold int           // Number of successes in half-open before closing
	timeout          time.Duration // How long to stay open before transitioning to half-open
	lastFailureTime  time.Time
}

// NewCircuitBreaker creates a circuit breaker with the given configuration.
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Execute runs the given function through the circuit breaker.
// Returns an error if the circuit is open or the function fails.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from OPEN to HALF-OPEN
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			fmt.Println("  [CB] Timeout elapsed. Transitioning OPEN → HALF-OPEN")
			cb.state = StateHalfOpen
			cb.successCount = 0
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is OPEN — request blocked")
		}
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Failure
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.state == StateHalfOpen {
			// Any failure in half-open sends us back to open
			fmt.Println("  [CB] Failure in HALF-OPEN. Transitioning back to OPEN.")
			cb.state = StateOpen
		} else if cb.failureCount >= cb.failureThreshold {
			fmt.Printf("  [CB] Failure threshold reached (%d). Transitioning CLOSED → OPEN.\n",
				cb.failureThreshold)
			cb.state = StateOpen
		}
		return err
	}

	// Success
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			fmt.Printf("  [CB] Success threshold reached (%d). Transitioning HALF-OPEN → CLOSED.\n",
				cb.successThreshold)
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	} else {
		// Reset failure count on success in closed state
		cb.failureCount = 0
	}

	return nil
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// flakyService simulates an external service that can fail.
func flakyService(failProbability float64) func() error {
	return func() error {
		if rand.Float64() < failProbability {
			return errors.New("service unavailable (503)")
		}
		return nil
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 19: CIRCUIT BREAKER                 ")
	fmt.Println("==================================================")

	// Create a circuit breaker:
	//   - Opens after 3 consecutive failures
	//   - Moves to half-open after 1 second
	//   - Closes after 2 consecutive successes in half-open
	cb := NewCircuitBreaker(3, 2, 1*time.Second)

	// --- Phase 1: Service is failing (100% failure rate) ---
	fmt.Println("\n--- Phase 1: Service Failing (100% failure) ---")
	for i := 1; i <= 6; i++ {
		err := cb.Execute(func() error {
			return errors.New("connection refused")
		})
		state := cb.GetState()
		if err != nil {
			fmt.Printf("  Request %d: ❌ %v [State: %s]\n", i, err, state)
		} else {
			fmt.Printf("  Request %d: ✅ Success [State: %s]\n", i, state)
		}
	}

	// --- Phase 2: Wait for timeout to allow half-open ---
	fmt.Println("\n--- Phase 2: Waiting for circuit breaker timeout (1s) ---")
	time.Sleep(1100 * time.Millisecond)

	// --- Phase 3: Service has recovered ---
	fmt.Println("\n--- Phase 3: Service Recovered (0% failure) ---")
	for i := 7; i <= 12; i++ {
		err := cb.Execute(func() error {
			return nil // Service is healthy now
		})
		state := cb.GetState()
		if err != nil {
			fmt.Printf("  Request %d: ❌ %v [State: %s]\n", i, err, state)
		} else {
			fmt.Printf("  Request %d: ✅ Success [State: %s]\n", i, state)
		}
	}

	// --- Phase 4: Flaky service (50% failure) ---
	fmt.Println("\n--- Phase 4: Flaky Service (50% failure) ---")
	cb2 := NewCircuitBreaker(3, 2, 500*time.Millisecond)
	service := flakyService(0.5)

	for i := 1; i <= 15; i++ {
		err := cb2.Execute(service)
		state := cb2.GetState()
		if err != nil {
			fmt.Printf("  Request %2d: ❌ %v [State: %s]\n", i, err, state)
		} else {
			fmt.Printf("  Request %2d: ✅ Success [State: %s]\n", i, state)
		}

		// If the circuit is open, wait for it to transition
		if state == StateOpen {
			fmt.Println("  ⏳ Circuit is OPEN. Waiting 600ms for half-open...")
			time.Sleep(600 * time.Millisecond)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}

	fmt.Println("\nCircuit breaker demo complete!")
	fmt.Println("Tutorial 19 complete!")
}
