package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ==================================================
// TUTORIAL 30: MIDDLEWARE CHAIN
// ==================================================
//
// PROBLEM STATEMENT:
// Middleware is code that runs before or after an HTTP handler. A Middleware
// Chain allows you to compose multiple middleware functions into a pipeline
// that wraps your handler. Each middleware can:
//   - Modify the request (add headers, parse auth)
//   - Short-circuit the chain (reject unauthorized requests)
//   - Modify the response (add CORS headers)
//   - Measure timing, log requests, etc.
//
// ARCHITECTURE:
//
//   Request → Logger → Auth → CORS → RateLimit → Handler → Response
//              ↕        ↕      ↕        ↕
//          (wraps the next handler in the chain)
//

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middlewares into a single middleware.
// Middlewares are applied in the order they are passed:
//   Chain(A, B, C)(handler) → A(B(C(handler)))
// So A runs first (outermost), then B, then C, then the handler.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		// Apply middlewares in reverse order so the first middleware
		// in the slice is the outermost (runs first).
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// --- Middleware Implementations ---

// LoggingMiddleware logs every request with method, path, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("[Logger] → %s %s\n", r.Method, r.URL.Path)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		duration := time.Since(start)
		fmt.Printf("[Logger] ← %s %s (took %v)\n", r.Method, r.URL.Path, duration)
	})
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware checks for a valid Authorization header.
// In production, you would validate JWT tokens here.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Skip auth for certain paths
		if r.URL.Path == "/health" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		if authHeader == "" {
			fmt.Println("[Auth] ❌ No Authorization header")
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return // Short-circuit — don't call next handler
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("[Auth] ❌ Invalid token format")
			http.Error(w, `{"error": "invalid token format"}`, http.StatusUnauthorized)
			return
		}

		fmt.Printf("[Auth] ✅ Valid token: %s...\n", authHeader[:20])
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds a unique request ID to the response headers.
var requestCounter int

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter++
		requestID := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), requestCounter)
		w.Header().Set("X-Request-ID", requestID)
		fmt.Printf("[RequestID] Assigned %s\n", requestID)
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware catches panics and returns a 500 error.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[Recovery] ⚠️  Recovered from panic: %v\n", err)
				http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 30: MIDDLEWARE CHAIN                 ")
	fmt.Println("==================================================")

	// Build the middleware chain.
	// Order: Recovery → Logger → RequestID → CORS → Auth → Handler
	chain := Chain(
		RecoveryMiddleware, // Outermost — catches panics from any layer
		LoggingMiddleware,  // Logs all requests
		RequestIDMiddleware, // Assigns request IDs
		CORSMiddleware,     // Adds CORS headers
		AuthMiddleware,     // Checks authentication
	)

	// --- Define handlers ---

	// Protected endpoint
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "Hello from the API!", "status": "success"}`)
	})

	// Public endpoint (auth skipped)
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "healthy", "uptime": "running"}`)
	})

	// Panicking endpoint (tests recovery middleware)
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went terribly wrong!")
	})

	// Apply the middleware chain to each handler
	mux := http.NewServeMux()
	mux.Handle("/api/data", chain(apiHandler))
	mux.Handle("/health", chain(healthHandler))
	mux.Handle("/panic", chain(panicHandler))

	// Start the server
	fmt.Println("\nMiddleware Chain Order:")
	fmt.Println("  Request → Recovery → Logger → RequestID → CORS → Auth → Handler")
	fmt.Println("\nServer starting on :8080")
	fmt.Println("  GET /health        — public (no auth)")
	fmt.Println("  GET /api/data      — protected (requires Bearer token)")
	fmt.Println("  GET /panic         — triggers panic recovery")
	fmt.Println("\nTest commands:")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println(`  curl -H "Authorization: Bearer my-token-123" http://localhost:8080/api/data`)
	fmt.Println("  curl http://localhost:8080/api/data  (should return 401)")
	fmt.Println("  curl http://localhost:8080/panic     (should return 500)")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
