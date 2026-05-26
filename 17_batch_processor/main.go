package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 17: BATCH PROCESSOR
// ==================================================
//
// PROBLEM STATEMENT:
// Instead of processing events one at a time (which creates overhead per
// event), a Batch Processor collects events and processes them in groups.
// A batch is flushed when either:
//   - The batch reaches a maximum size (e.g., 10 items)
//   - A timeout expires (e.g., every 2 seconds)
//
// This pattern is common in:
//   - Log aggregation (send logs in batches)
//   - Database bulk inserts
//   - Metrics collection and reporting
//
// ARCHITECTURE:
//
//   Events ──▶ [Collector] ──batch──▶ [Processor]
//                   │
//              (flush when full OR timeout)
//

// Event represents an incoming data point to be batched.
type Event struct {
	ID        int
	Payload   string
	Timestamp time.Time
}

// BatchProcessor collects events and flushes them in batches.
type BatchProcessor struct {
	input     chan Event
	batchSize int
	timeout   time.Duration
	processor func([]Event) // Function called to process each batch
	done      chan struct{}
}

// NewBatchProcessor creates a batch processor.
//   - batchSize: max number of events per batch before forced flush.
//   - timeout: max time to wait before flushing a partial batch.
//   - processor: the function that handles each batch of events.
func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]Event)) *BatchProcessor {
	bp := &BatchProcessor{
		input:     make(chan Event, batchSize*2), // Buffer to prevent producer blocking
		batchSize: batchSize,
		timeout:   timeout,
		processor: processor,
		done:      make(chan struct{}),
	}

	go bp.run()
	return bp
}

// run is the main loop that collects events and flushes batches.
func (bp *BatchProcessor) run() {
	defer close(bp.done)

	batch := make([]Event, 0, bp.batchSize)
	timer := time.NewTimer(bp.timeout)
	defer timer.Stop()

	for {
		select {
		case event, ok := <-bp.input:
			if !ok {
				// Channel closed — flush remaining events and exit
				if len(batch) > 0 {
					fmt.Printf("  [BatchProcessor] Channel closed. Flushing final batch (%d events).\n", len(batch))
					bp.processor(batch)
				}
				return
			}

			batch = append(batch, event)

			// Flush if batch is full
			if len(batch) >= bp.batchSize {
				fmt.Printf("  [BatchProcessor] Batch full (%d events). Flushing.\n", len(batch))
				bp.processor(batch)
				batch = make([]Event, 0, bp.batchSize) // Reset batch
				timer.Reset(bp.timeout)                 // Reset timeout
			}

		case <-timer.C:
			// Timeout expired — flush whatever we have
			if len(batch) > 0 {
				fmt.Printf("  [BatchProcessor] Timeout expired. Flushing partial batch (%d events).\n", len(batch))
				bp.processor(batch)
				batch = make([]Event, 0, bp.batchSize)
			}
			timer.Reset(bp.timeout)
		}
	}
}

// Submit adds an event to the batch processor.
func (bp *BatchProcessor) Submit(event Event) {
	bp.input <- event
}

// Close stops the batch processor and waits for it to flush.
func (bp *BatchProcessor) Close() {
	close(bp.input)
	<-bp.done // Wait for the run loop to finish
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("       TUTORIAL 17: BATCH PROCESSOR                ")
	fmt.Println("==================================================")

	batchNumber := 0
	var mu sync.Mutex

	// Create a batch processor with:
	//   - Max batch size: 5 events
	//   - Flush timeout: 1 second
	processor := NewBatchProcessor(5, 1*time.Second, func(events []Event) {
		mu.Lock()
		batchNumber++
		num := batchNumber
		mu.Unlock()

		fmt.Printf("\n  ✅ Processing Batch #%d (%d events):\n", num, len(events))
		for _, e := range events {
			fmt.Printf("    - Event %d: %s (at %s)\n",
				e.ID, e.Payload, e.Timestamp.Format("15:04:05.000"))
		}
		// Simulate batch processing time
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("  Batch #%d processed.\n\n", num)
	})

	// Simulate events arriving at variable rates
	fmt.Println("\n--- Phase 1: Rapid burst (should trigger size-based flush) ---")
	for i := 1; i <= 7; i++ {
		processor.Submit(Event{
			ID:        i,
			Payload:   fmt.Sprintf("rapid-event-%d", i),
			Timestamp: time.Now(),
		})
		time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)
	}

	fmt.Println("\n--- Phase 2: Slow trickle (should trigger timeout-based flush) ---")
	for i := 8; i <= 10; i++ {
		processor.Submit(Event{
			ID:        i,
			Payload:   fmt.Sprintf("slow-event-%d", i),
			Timestamp: time.Now(),
		})
		time.Sleep(400 * time.Millisecond) // Slow — won't fill batch before timeout
	}

	// Wait for the timeout flush
	time.Sleep(1200 * time.Millisecond)

	fmt.Println("--- Phase 3: Shutting down (should flush remaining) ---")
	for i := 11; i <= 13; i++ {
		processor.Submit(Event{
			ID:        i,
			Payload:   fmt.Sprintf("final-event-%d", i),
			Timestamp: time.Now(),
		})
	}

	processor.Close()

	fmt.Println("Batch processor shut down. All events processed!")
	fmt.Println("Tutorial 17 complete!")
}
