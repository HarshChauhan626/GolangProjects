package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 20: CONCURRENT TASK QUEUE
// ==================================================
//
// PROBLEM STATEMENT:
// A task queue accepts tasks from multiple producers, stores them in a
// thread-safe queue, and processes them with a configurable pool of workers.
// Unlike a simple channel, this implementation provides:
//   - Enqueue/Dequeue API
//   - Task status tracking
//   - Graceful shutdown
//   - Configurable concurrency
//
// ARCHITECTURE:
//
//   Producers ──Enqueue──▶ [Thread-safe Queue] ──▶ Worker Pool ──▶ Results
//

// TaskStatus tracks the lifecycle of a task.
type TaskStatus int

const (
	TaskPending    TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
)

func (s TaskStatus) String() string {
	switch s {
	case TaskPending:
		return "PENDING"
	case TaskRunning:
		return "RUNNING"
	case TaskCompleted:
		return "COMPLETED"
	case TaskFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// Task represents a unit of work in the queue.
type Task struct {
	ID       int
	Name     string
	Status   TaskStatus
	Execute  func() error // The actual work to perform
	Result   string
	Error    error
}

// TaskQueue is a concurrent, thread-safe task processing queue.
type TaskQueue struct {
	mu         sync.Mutex
	tasks      []*Task
	taskChan   chan *Task
	numWorkers int
	wg         sync.WaitGroup
	done       chan struct{}
	results    chan *Task
}

// NewTaskQueue creates a task queue with the specified number of workers.
func NewTaskQueue(numWorkers int) *TaskQueue {
	tq := &TaskQueue{
		tasks:      make([]*Task, 0),
		taskChan:   make(chan *Task, 100),
		numWorkers: numWorkers,
		done:       make(chan struct{}),
		results:    make(chan *Task, 100),
	}

	// Start worker pool
	for i := 1; i <= numWorkers; i++ {
		tq.wg.Add(1)
		go tq.worker(i)
	}

	return tq
}

// worker processes tasks from the internal channel.
func (tq *TaskQueue) worker(id int) {
	defer tq.wg.Done()

	for task := range tq.taskChan {
		fmt.Printf("  [Worker %d] Executing task %d: %s\n", id, task.ID, task.Name)

		// Update status to running
		tq.mu.Lock()
		task.Status = TaskRunning
		tq.mu.Unlock()

		// Execute the task
		err := task.Execute()

		// Update status based on result
		tq.mu.Lock()
		if err != nil {
			task.Status = TaskFailed
			task.Error = err
		} else {
			task.Status = TaskCompleted
			task.Result = fmt.Sprintf("Task %d completed by worker %d", task.ID, id)
		}
		tq.mu.Unlock()

		tq.results <- task
	}
}

// Enqueue adds a task to the queue.
func (tq *TaskQueue) Enqueue(task *Task) {
	tq.mu.Lock()
	task.Status = TaskPending
	tq.tasks = append(tq.tasks, task)
	tq.mu.Unlock()

	tq.taskChan <- task
	fmt.Printf("  [Queue] Enqueued task %d: %s\n", task.ID, task.Name)
}

// Results returns the results channel for reading completed tasks.
func (tq *TaskQueue) Results() <-chan *Task {
	return tq.results
}

// Shutdown gracefully shuts down the task queue.
func (tq *TaskQueue) Shutdown() {
	close(tq.taskChan) // Stop accepting new tasks
	tq.wg.Wait()       // Wait for all workers to finish
	close(tq.results)   // Close results channel
}

// GetStats returns counts of tasks by status.
func (tq *TaskQueue) GetStats() map[TaskStatus]int {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	stats := make(map[TaskStatus]int)
	for _, t := range tq.tasks {
		stats[t.Status]++
	}
	return stats
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("     TUTORIAL 20: CONCURRENT TASK QUEUE            ")
	fmt.Println("==================================================")

	// Create a task queue with 3 workers
	queue := NewTaskQueue(3)
	fmt.Println("Task queue created with 3 workers.\n")

	// Enqueue tasks from multiple "producers" concurrently
	var producerWg sync.WaitGroup

	// Producer 1: Sends computation tasks
	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		for i := 1; i <= 4; i++ {
			taskID := i
			queue.Enqueue(&Task{
				ID:   taskID,
				Name: fmt.Sprintf("Compute-%d", taskID),
				Execute: func() error {
					time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
					return nil
				},
			})
		}
	}()

	// Producer 2: Sends I/O tasks (some may fail)
	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		for i := 5; i <= 8; i++ {
			taskID := i
			queue.Enqueue(&Task{
				ID:   taskID,
				Name: fmt.Sprintf("IO-%d", taskID),
				Execute: func() error {
					time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
					if rand.Float64() < 0.3 {
						return fmt.Errorf("I/O error on task %d", taskID)
					}
					return nil
				},
			})
		}
	}()

	// Collect results in another goroutine
	var collectWg sync.WaitGroup
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for task := range queue.Results() {
			if task.Error != nil {
				fmt.Printf("  ❌ Task %d (%s): FAILED — %v\n",
					task.ID, task.Name, task.Error)
			} else {
				fmt.Printf("  ✅ Task %d (%s): %s\n",
					task.ID, task.Name, task.Result)
			}
		}
	}()

	// Wait for producers to finish enqueueing, then shut down
	producerWg.Wait()
	fmt.Println("\nAll tasks enqueued. Shutting down queue...")
	queue.Shutdown()
	collectWg.Wait()

	// Print final stats
	stats := queue.GetStats()
	fmt.Printf("\n--- Final Stats ---\n")
	for status, count := range stats {
		fmt.Printf("  %s: %d\n", status, count)
	}

	fmt.Println("\nConcurrent task queue demo complete!")
	fmt.Println("Tutorial 20 complete!")
}
