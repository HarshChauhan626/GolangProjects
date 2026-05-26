package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ==================================================
// TUTORIAL 14: GRACEFUL SHUTDOWN
// ==================================================
//
// PROBLEM STATEMENT:
// When a server or application receives a termination signal (SIGINT,
// SIGTERM), it should not abruptly kill in-flight requests or background
// workers. Instead, it should:
//   1. Stop accepting new work
//   2. Wait for existing work to complete (with a timeout)
//   3. Clean up resources (close DB connections, flush buffers)
//   4. Exit cleanly
//
// KEY CONCEPTS:
// - os/signal.Notify to intercept OS signals
// - http.Server.Shutdown(ctx) for graceful HTTP server shutdown
// - context.WithTimeout to bound the shutdown grace period
// - sync.WaitGroup to wait for background workers
//
// ARCHITECTURE:
//
//   SIGINT/SIGTERM received
//        │
//        ├──▶ Stop accepting new HTTP requests
//        ├──▶ Signal background workers to stop
//        ├──▶ Wait for in-flight requests to complete
//        ├──▶ Wait for workers to finish (with timeout)
//        └──▶ Clean exit
//

// backgroundWorker simulates a long-running background task (e.g., a job
// processor, metrics collector) that respects cancellation.
func backgroundWorker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[Worker %d] Started\n", id)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	iteration := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[Worker %d] Received shutdown signal. Cleaning up...\n", id)
			// Simulate cleanup (flush buffers, close connections, etc.)
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("[Worker %d] Cleanup complete. Exiting.\n", id)
			return

		case <-ticker.C:
			iteration++
			fmt.Printf("[Worker %d] Processing batch #%d\n", id, iteration)
		}
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 14: GRACEFUL SHUTDOWN               ")
	fmt.Println("==================================================")
	fmt.Println("Press Ctrl+C to initiate graceful shutdown.\n")

	// Master context — cancelling this signals ALL components to shut down.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var workerWg sync.WaitGroup

	// 1. Start background workers
	numWorkers := 2
	for i := 1; i <= numWorkers; i++ {
		workerWg.Add(1)
		go backgroundWorker(ctx, i, &workerWg)
	}

	// 2. Set up the HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a request that takes some time
		fmt.Println("[HTTP] Handling request...")
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintln(w, "Hello! The server is running.")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "healthy"}`)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 3. Start the HTTP server in a goroutine
	go func() {
		fmt.Println("[HTTP] Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HTTP] Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// 4. Listen for OS signals (Ctrl+C = SIGINT, `kill` = SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	sig := <-sigChan
	fmt.Printf("\n>>> Received signal: %v. Starting graceful shutdown...\n", sig)

	// 5. Create a deadline for the entire shutdown process
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 6. Shut down the HTTP server gracefully
	// This stops accepting new connections and waits for in-flight requests.
	fmt.Println("[HTTP] Shutting down server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("[HTTP] Server shutdown error: %v\n", err)
	} else {
		fmt.Println("[HTTP] Server stopped accepting new connections.")
	}

	// 7. Signal background workers to stop
	fmt.Println("[Workers] Signaling workers to stop...")
	cancel() // This cancels the master context, triggering all workers

	// 8. Wait for workers to finish (with timeout)
	doneCh := make(chan struct{})
	go func() {
		workerWg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		fmt.Println("[Workers] All workers shut down cleanly.")
	case <-shutdownCtx.Done():
		fmt.Println("[Workers] Shutdown timeout exceeded! Forcing exit.")
	}

	fmt.Println("\nGraceful shutdown complete. Goodbye!")
	fmt.Println("Tutorial 14 complete!")
}
