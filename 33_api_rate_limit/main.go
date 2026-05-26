package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 33: API RATE LIMITING MIDDLEWARE
// ==================================================
//
// PROBLEM STATEMENT:
// Protect your API endpoints from abuse by limiting the number of requests
// a client can make within a time window. This middleware tracks requests
// per client IP using a sliding window approach.
//
// FEATURES:
// - Per-client rate limiting (by IP address)
// - Configurable rate (requests per window)
// - Proper HTTP headers (X-RateLimit-Limit, X-RateLimit-Remaining, etc.)
// - 429 Too Many Requests response when limit exceeded
// - Automatic cleanup of expired client records
//
// ARCHITECTURE:
//
//   Request → Extract IP → Check rate → Allow (200) or Reject (429)
//                              │
//                    [Per-IP sliding window counter]
//

// ClientRecord tracks request timestamps for a single client.
type ClientRecord struct {
	Timestamps []time.Time
	mu         sync.Mutex
}

// RateLimiter manages rate limiting across all clients.
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*ClientRecord
	limit    int           // Max requests per window
	window   time.Duration // Time window
}

// NewRateLimiter creates a rate limiter.
// For example: NewRateLimiter(10, time.Minute) allows 10 requests per minute.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientRecord),
		limit:   limit,
		window:  window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a client is within their rate limit.
// Returns (allowed, remaining requests, reset time).
func (rl *RateLimiter) Allow(clientIP string) (bool, int, time.Time) {
	rl.mu.Lock()
	record, exists := rl.clients[clientIP]
	if !exists {
		record = &ClientRecord{}
		rl.clients[clientIP] = record
	}
	rl.mu.Unlock()

	record.mu.Lock()
	defer record.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove timestamps outside the current window (sliding window)
	validTimestamps := make([]time.Time, 0)
	for _, ts := range record.Timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	record.Timestamps = validTimestamps

	// Calculate remaining requests
	remaining := rl.limit - len(record.Timestamps)

	// Calculate reset time (when the oldest timestamp in the window expires)
	var resetTime time.Time
	if len(record.Timestamps) > 0 {
		resetTime = record.Timestamps[0].Add(rl.window)
	} else {
		resetTime = now.Add(rl.window)
	}

	// Check if under the limit
	if len(record.Timestamps) >= rl.limit {
		return false, 0, resetTime
	}

	// Allow — record this request
	record.Timestamps = append(record.Timestamps, now)
	remaining = rl.limit - len(record.Timestamps)

	return true, remaining, resetTime
}

// cleanup periodically removes stale client records.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for ip, record := range rl.clients {
			record.mu.Lock()
			// Remove clients with no recent activity
			hasRecent := false
			for _, ts := range record.Timestamps {
				if ts.After(windowStart) {
					hasRecent = true
					break
				}
			}
			if !hasRecent {
				delete(rl.clients, ip)
			}
			record.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates an HTTP middleware that enforces rate limits.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP
			clientIP := getClientIP(r)

			// Check rate limit
			allowed, remaining, resetTime := limiter.Allow(clientIP)

			// Always set rate limit headers (RFC 6585 / draft-ietf-httpapi-ratelimit-headers)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			if !allowed {
				retryAfter := time.Until(resetTime).Seconds()
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter))

				fmt.Printf("[RateLimit] ❌ %s exceeded limit (%d/%d). Retry after %.0fs\n",
					clientIP, limiter.limit, limiter.limit, retryAfter)

				http.Error(w,
					fmt.Sprintf(`{"error": "rate limit exceeded", "retry_after_seconds": %.0f}`, retryAfter),
					http.StatusTooManyRequests) // 429
				return
			}

			fmt.Printf("[RateLimit] ✅ %s — %d/%d remaining\n",
				clientIP, remaining, limiter.limit)

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP, handling proxies.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (for clients behind proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := splitAndTrim(xff, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	for {
		i := indexOf(s, sep)
		if i < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:i])
		s = s[i+len(sep):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trim(s string) string {
	start := 0
	for start < len(s) && s[start] == ' ' {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("  TUTORIAL 33: API RATE LIMITING MIDDLEWARE         ")
	fmt.Println("==================================================")

	// Create a rate limiter: 5 requests per 10 seconds
	limiter := NewRateLimiter(5, 10*time.Second)
	rateLimitMW := RateLimitMiddleware(limiter)

	mux := http.NewServeMux()

	// Protected endpoint
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "API response", "status": "success"}`)
	})

	// Apply rate limiting middleware
	mux.Handle("/api/data", rateLimitMW(apiHandler))

	// Health check (not rate limited)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status": "healthy"}`)
	})

	fmt.Println("\nRate Limit: 5 requests per 10 seconds (per IP)")
	fmt.Println("\nServer starting on :8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET /api/data  — Rate limited endpoint")
	fmt.Println("  GET /health    — Not rate limited")
	fmt.Println("\nTest: Run this 6+ times quickly:")
	fmt.Println("  curl -v http://localhost:8080/api/data")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
