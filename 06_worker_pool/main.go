package main

import (
	"fmt"
	"sync"
	"time"
)

// A job structure representing the unit of work
type Job struct {
	ID        int
	Payload   int
	CreatedAt time.Time
}

// A result structure representing the output of a job
type Result struct {
	JobID      int
	Output     int
	WorkerID   int
	Duration   time.Duration
}

// Worker function: Processes jobs from the jobs channel and writes results to the results channel.
// The jobs channel is receive-only (<-chan), and results is send-only (chan<-).
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done() // Signal to the WaitGroup when this worker exits

	fmt.Printf("[Worker %d] Started and waiting for jobs...\n", id)

	// Range loop terminates when the 'jobs' channel is closed AND empty
	for job := range jobs {
		startTime := time.Now()
		fmt.Printf("[Worker %d] Processing Job %d (payload: %d)...\n", id, job.ID, job.Payload)

		// Simulate processing delay (e.g., squaring the payload)
		time.Sleep(100 * time.Millisecond)
		output := job.Payload * job.Payload

		duration := time.Since(startTime)
		results <- Result{
			JobID:    job.ID,
			Output:   output,
			WorkerID: id,
			Duration: duration,
		}
		fmt.Printf("[Worker %d] Finished Job %d.\n", id, job.ID)
	}

	fmt.Printf("[Worker %d] No more jobs. Exiting.\n", id)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 06: WORKER POOLS                ")
	fmt.Println("==================================================")

	numJobs := 9
	numWorkers := 3

	// Channels for communication
	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// WaitGroup to track workers
	var wg sync.WaitGroup

	// 1. Spawning the worker pool
	fmt.Printf("Spawning %d worker goroutines...\n", numWorkers)
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// 2. Feeding the jobs channel
	fmt.Printf("\nAdding %d jobs to the queue...\n", numJobs)
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{
			ID:        j,
			Payload:   j * 10,
			CreatedAt: time.Now(),
		}
	}

	// Close the jobs channel to tell workers there is no more work.
	// Workers will finish processing any currently queued jobs, then exit their 'range jobs' loop.
	close(jobs)
	fmt.Println("Closed the 'jobs' channel. Senders are done.")

	// 3. Monitor workers in a separate goroutine
	// Once all workers exit, we close the results channel so the main range loop can finish.
	go func() {
		wg.Wait()
		fmt.Println("\nAll workers have completed. Closing results channel...")
		close(results)
	}()

	// 4. Collect results from the results channel
	// This range loop runs until the results channel is closed by the monitor goroutine.
	fmt.Println("Collecting results...")
	for res := range results {
		fmt.Printf("--> [Result] Job %d processed by Worker %d | Output: %d | Time Taken: %v\n",
			res.JobID, res.WorkerID, res.Output, res.Duration)
	}

	fmt.Println("\nAll results collected! Worker pool execution complete.")
	fmt.Println("Tutorial 06 complete!")
}
