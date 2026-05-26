package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 08: FAN-OUT PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// Fan-Out is a concurrency pattern where a single source of work is
// distributed (fanned-out) across multiple goroutines for parallel
// processing. Each worker independently processes a subset of the tasks.
//
// USE CASES:
// - CPU-bound work that benefits from parallelism (e.g., image processing)
// - I/O-bound tasks that can run concurrently (e.g., calling multiple APIs)
// - Any workload that can be divided into independent units
//
// ARCHITECTURE:
//
//                    ┌──▶ Worker 1 ──▶ Result
//                    │
//   Input Channel ───┼──▶ Worker 2 ──▶ Result
//                    │
//                    └──▶ Worker 3 ──▶ Result
//

// Task represents work to be processed.
type Task struct {
	ID   int
	Data int
}

// Result represents the output of processing a task.
type Result struct {
	TaskID   int
	WorkerID int
	Output   int
}

// fanOutWorker reads tasks from the shared input channel, processes them,
// and sends results to its own (or shared) results channel.
func fanOutWorker(id int, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		fmt.Printf("[Worker %d] Processing task %d (data: %d)\n", id, task.ID, task.Data)

		// Simulate processing — compute the factorial iteratively
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		output := factorial(task.Data)

		results <- Result{
			TaskID:   task.ID,
			WorkerID: id,
			Output:   output,
		}
	}
}

// factorial computes n! iteratively (our simulated "expensive" computation).
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 08: FAN-OUT PATTERN              ")
	fmt.Println("==================================================")

	numWorkers := 4
	numTasks := 12

	// Single input channel — all workers read from this one channel.
	// Go's channel semantics ensure each task is delivered to exactly one worker.
	tasks := make(chan Task, numTasks)
	results := make(chan Result, numTasks)

	var wg sync.WaitGroup

	// 1. Launch multiple workers — this is the "fan-out"
	fmt.Printf("Fanning out work to %d workers...\n\n", numWorkers)
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go fanOutWorker(w, tasks, results, &wg)
	}

	// 2. Send all tasks into the shared channel
	for t := 1; t <= numTasks; t++ {
		tasks <- Task{ID: t, Data: t + 2} // factorial of (t+2)
	}
	close(tasks) // Signal that no more tasks will be sent

	// 3. Close results channel once all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Collect all results
	fmt.Println("\n--- Results ---")
	for res := range results {
		fmt.Printf("Task %2d → Worker %d computed factorial(%d) = %d\n",
			res.TaskID, res.WorkerID, res.TaskID+2, res.Output)
	}

	fmt.Println("\nAll tasks processed via Fan-Out!")
	fmt.Println("Tutorial 08 complete!")
}
