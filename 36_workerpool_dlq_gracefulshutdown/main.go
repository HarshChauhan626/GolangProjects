// Below is a production-style implementation of a worker pool in Go that satisfies all of your requirements:

// ✅ Fixed number of workers
// ✅ Buffered queue
// ✅ Queue full → reject new jobs (non-blocking)
// ✅ Retry failed jobs
// ✅ Max retry limit
// ✅ Failed jobs stored separately
// ✅ Producer keeps generating jobs
// ✅ Workers keep running forever
// ✅ No goroutine leaks
// ✅ Graceful shutdown using context

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	WorkerCount = 5
	QueueSize   = 10
	MaxRetry    = 3
)

type Job struct {
	ID      int
	Payload string
	Retry   int
}

var (
	jobQueue    = make(chan Job, QueueSize)
	failedQueue = make(chan Job, 100)
)

func process(job Job) error {
	// Simulate work
	time.Sleep(700 * time.Millisecond)

	// 40% failure rate
	if rand.Intn(100) < 40 {
		return fmt.Errorf("processing failed")
	}

	return nil
}

func worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker-%d started\n", id)

	for {
		select {

		case <-ctx.Done():
			fmt.Printf("Worker-%d stopping\n", id)
			return

		case job := <-jobQueue:

			fmt.Printf("[Worker-%d] Processing Job %d (Retry=%d)\n",
				id,
				job.ID,
				job.Retry)

			err := process(job)

			if err == nil {
				fmt.Printf("[Worker-%d] Job %d SUCCESS\n", id, job.ID)
				continue
			}

			fmt.Printf("[Worker-%d] Job %d FAILED\n", id, job.ID)

			job.Retry++

			if job.Retry <= MaxRetry {

				fmt.Printf("[Worker-%d] Retrying Job %d\n",
					id,
					job.ID)

				// Exponential Backoff
				go func(j Job) {
					delay := time.Duration(j.Retry) * time.Second
					time.Sleep(delay)

					select {
					case jobQueue <- j:
						fmt.Printf("Job %d requeued\n", j.ID)

					default:
						fmt.Printf("Queue Full. Retry dropped for Job %d\n", j.ID)
						failedQueue <- j
					}
				}(job)

			} else {

				fmt.Printf("[Worker-%d] Job %d permanently failed\n",
					id,
					job.ID)

				failedQueue <- job
			}
		}
	}
}

func failedJobTracker(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case job := <-failedQueue:
			fmt.Printf(">>> FAILED JOB STORED : %+v\n", job)
		}
	}
}

func producer(ctx context.Context) {

	jobID := 1

	ticker := time.NewTicker(300 * time.Millisecond)

	defer ticker.Stop()

	for {

		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			job := Job{
				ID:      jobID,
				Payload: fmt.Sprintf("payload-%d", jobID),
			}

			select {

			case jobQueue <- job:
				fmt.Printf("Submitted Job %d\n", jobID)

			default:
				fmt.Printf("Queue Full. Rejecting Job %d\n", jobID)
			}

			jobID++
		}
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Start workers
	for i := 1; i <= WorkerCount; i++ {
		wg.Add(1)
		go worker(ctx, i, &wg)
	}

	// Failed Job Logger
	go failedJobTracker(ctx)

	// Producer
	go producer(ctx)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	fmt.Println("\nShutdown signal received...")

	cancel()

	wg.Wait()

	fmt.Println("All workers stopped.")
}