package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// ==================================================
// TUTORIAL 34: PANIC RECOVERY MIDDLEWARE
// ==================================================
//
// PROBLEM STATEMENT:
// If an HTTP handler panics (e.g., nil pointer dereference, index out of
// range), the default behavior crashes the entire server. A Recovery
// Middleware catches these panics and:
//   - Returns a proper 500 Internal Server Error to the client
//   - Logs the panic with a stack trace for debugging
//   - Keeps the server running to handle other requests
//
// Without this, a single bad request could take down your entire service.
//
// ARCHITECTURE:
//
//   Request → Recovery Middleware → Handler
//                   │
//             defer recover()
//                   │
//             panic caught → log + return 500 → server keeps running
//

// PanicRecoveryMiddleware catches panics and returns a 500 error.
// It uses defer/recover to intercept panics from downstream handlers.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get the stack trace
				stack := debug.Stack()

				// Log the panic with full details
				fmt.Println("\n" + "=" + repeatChar('=', 60))
				fmt.Printf("🚨 PANIC RECOVERED\n")
				fmt.Printf("   Path:     %s %s\n", r.Method, r.URL.Path)
				fmt.Printf("   Error:    %v\n", err)
				fmt.Printf("   Time:     %s\n", time.Now().Format(time.RFC3339))
				fmt.Printf("   Client:   %s\n", r.RemoteAddr)
				fmt.Println("   Stack Trace:")
				fmt.Println(string(stack))
				fmt.Println(repeatChar('=', 61))

				// Return a 500 Internal Server Error to the client
				// Important: check if headers have already been sent
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error": "internal server error", "message": "an unexpected error occurred"}`)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// repeatChar repeats a character n times.
func repeatChar(c byte, n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}

// --- Custom Recovery with Options ---

// RecoveryConfig configures the recovery middleware behavior.
type RecoveryConfig struct {
	EnableStackTrace bool                          // Include stack trace in logs
	LogFunc          func(err interface{}, stack []byte) // Custom log function
	ResponseFunc     func(w http.ResponseWriter, err interface{}) // Custom response
}

// DefaultRecoveryConfig returns sensible defaults.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		EnableStackTrace: true,
		LogFunc: func(err interface{}, stack []byte) {
			fmt.Printf("[Recovery] Panic: %v\n", err)
			if len(stack) > 0 {
				fmt.Printf("[Recovery] Stack:\n%s\n", stack)
			}
		},
		ResponseFunc: func(w http.ResponseWriter, err interface{}) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "internal server error"}`)
		},
	}
}

// ConfigurableRecoveryMiddleware is a customizable version of the recovery middleware.
func ConfigurableRecoveryMiddleware(config RecoveryConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					var stack []byte
					if config.EnableStackTrace {
						stack = debug.Stack()
					}

					if config.LogFunc != nil {
						config.LogFunc(err, stack)
					}

					if config.ResponseFunc != nil {
						config.ResponseFunc(w, err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// --- Demo Handlers ---

// normalHandler processes a request successfully.
func normalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"message": "This endpoint works fine!"}`)
}

// nilPanicHandler simulates a nil pointer dereference.
func nilPanicHandler(w http.ResponseWriter, r *http.Request) {
	var s *string
	fmt.Fprintln(w, *s) // PANIC: nil pointer dereference
}

// indexPanicHandler simulates an index out of range.
func indexPanicHandler(w http.ResponseWriter, r *http.Request) {
	numbers := []int{1, 2, 3}
	fmt.Fprintln(w, numbers[10]) // PANIC: index out of range
}

// customPanicHandler triggers a custom panic message.
func customPanicHandler(w http.ResponseWriter, r *http.Request) {
	panic("something went terribly wrong in business logic!")
}

// divisionPanicHandler triggers a division by zero (integer).
func divisionPanicHandler(w http.ResponseWriter, r *http.Request) {
	a := 42
	b := 0
	fmt.Fprintln(w, a/b) // PANIC: integer divide by zero
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("   TUTORIAL 34: PANIC RECOVERY MIDDLEWARE          ")
	fmt.Println("==================================================")

	mux := http.NewServeMux()

	// Apply recovery middleware to all handlers
	mux.Handle("/ok", PanicRecoveryMiddleware(http.HandlerFunc(normalHandler)))
	mux.Handle("/panic/nil", PanicRecoveryMiddleware(http.HandlerFunc(nilPanicHandler)))
	mux.Handle("/panic/index", PanicRecoveryMiddleware(http.HandlerFunc(indexPanicHandler)))
	mux.Handle("/panic/custom", PanicRecoveryMiddleware(http.HandlerFunc(customPanicHandler)))
	mux.Handle("/panic/division", PanicRecoveryMiddleware(http.HandlerFunc(divisionPanicHandler)))

	// Custom recovery config: don't show stack trace to client
	customConfig := RecoveryConfig{
		EnableStackTrace: true,
		LogFunc: func(err interface{}, stack []byte) {
			fmt.Printf("[CustomRecovery] Error: %v\n", err)
		},
		ResponseFunc: func(w http.ResponseWriter, err interface{}) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			// Don't expose internal error details to the client
			fmt.Fprintln(w, `{"error": "oops! something went wrong", "support": "contact admin@example.com"}`)
		},
	}
	mux.Handle("/panic/custom-recovery",
		ConfigurableRecoveryMiddleware(customConfig)(http.HandlerFunc(customPanicHandler)))

	fmt.Println("\nServer starting on :8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET /ok                    — Normal response")
	fmt.Println("  GET /panic/nil             — Nil pointer panic")
	fmt.Println("  GET /panic/index           — Index out of range panic")
	fmt.Println("  GET /panic/custom          — Custom panic message")
	fmt.Println("  GET /panic/division        — Division by zero panic")
	fmt.Println("  GET /panic/custom-recovery — Custom recovery handler")
	fmt.Println("\nTest: curl http://localhost:8080/panic/nil")
	fmt.Println("The server will recover and keep running!")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
