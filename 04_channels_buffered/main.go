package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("        TUTORIAL 04: BUFFERED CHANNELS            ")
	fmt.Println("==================================================")

	// 1. Creating a buffered channel
	// ch := make(chan T, capacity)
	// This channel has room for up to 3 strings before it blocks the sender.
	ch := make(chan string, 3)

	fmt.Printf("Initial - Length (items in buffer): %d, Capacity: %d\n", len(ch), cap(ch))

	// 2. Non-blocking writes
	fmt.Println("\n--- 1. Sending values within capacity (Non-blocking) ---")
	ch <- "message 1"
	fmt.Printf("Sent 'message 1' -> Length: %d, Capacity: %d\n", len(ch), cap(ch))

	ch <- "message 2"
	fmt.Printf("Sent 'message 2' -> Length: %d, Capacity: %d\n", len(ch), cap(ch))

	ch <- "message 3"
	fmt.Printf("Sent 'message 3' -> Length: %d, Capacity: %d\n", len(ch), cap(ch))

	// Attempting another write now would block because the buffer is full (length == capacity)
	fmt.Println("Channel is now FULL. Sending a 4th item would block the current goroutine.")

	// 3. Reading from the channel
	fmt.Println("\n--- 2. Receiving values ---")
	fmt.Printf("Received: %s\n", <-ch)
	fmt.Printf("After 1 receive -> Length: %d, Capacity: %d\n", len(ch), cap(ch))

	fmt.Printf("Received: %s\n", <-ch)
	fmt.Printf("Received: %s\n", <-ch)
	fmt.Printf("After 3 receives -> Length: %d, Capacity: %d\n", len(ch), cap(ch))

	// 4. Using Buffered Channels as a Queue (with Range Loop)
	fmt.Println("\n--- 3. Using range to consume closed buffered channels ---")
	jobs := make(chan int, 5)

	// Send 5 jobs
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	fmt.Println("Sent 5 jobs to the channel.")

	// We MUST close the channel so that the range loop knows when to stop.
	// A closed channel can still be read from until it is empty.
	close(jobs)
	fmt.Println("Closed the 'jobs' channel. Ready to consume via range loop.")

	// Range loop automatically stops when the channel is closed AND empty.
	for job := range jobs {
		fmt.Printf("Consumed job: %d (remaining in channel: %d)\n", job, len(jobs))
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("\nTutorial 04 complete!")
}
