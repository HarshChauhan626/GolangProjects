package main

import (
	"fmt"
	"time"
)

// A function that does some work and signals completion via an unbuffered channel
func process(done chan bool) {
	fmt.Println("[Worker] Starting intensive processing...")
	time.Sleep(100 * time.Millisecond) // Simulate work
	fmt.Println("[Worker] Work finished! Sending signal to 'done' channel...")

	// Sending a value to the channel. This will BLOCK until main receives it.
	done <- true

	fmt.Println("[Worker] Signal sent! Worker exiting.")
}

// A function that sends multiple messages and then closes the channel
func generateData(dataChan chan string) {
	messages := []string{"apple", "banana", "cherry"}

	for _, msg := range messages {
		fmt.Printf("[Generator] Sending fruit: %s (will block until read)\n", msg)
		dataChan <- msg
		// We sleep to demonstrate that the receiver waits for us
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("[Generator] Finished sending all data. Closing channel...")
	// Closing the channel signals to the receiver that no more values will be sent.
	close(dataChan)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("       TUTORIAL 03: UNBUFFERED CHANNELS           ")
	fmt.Println("==================================================")

	// 1. Basic Synchronization with Unbuffered Channels
	fmt.Println("\n--- 1. Synchronization via Unbuffered Channel ---")
	// Make an unbuffered boolean channel
	done := make(chan bool)

	// Launch worker
	go process(done)

	fmt.Println("[Main] Waiting for worker signal (blocking)...")
	// Receiving from the channel. This BLOCKS main until the worker sends.
	workerSignal := <-done
	fmt.Printf("[Main] Received signal from worker: %t. Resuming main.\n", workerSignal)

	// 2. Multi-value streaming & closing channels
	fmt.Println("\n--- 2. Streaming and Channel Closure ---")
	dataChan := make(chan string)

	// Launch generator
	go generateData(dataChan)

	fmt.Println("[Main] Reading stream of fruits...")
	for {
		// Reading from channel returning value and status ok
		// ok is true if the value was sent by a send operation.
		// ok is false if the channel is closed and there are no values left.
		val, ok := <-dataChan
		if !ok {
			fmt.Println("[Main] Channel closed! Exiting read loop.")
			break
		}
		fmt.Printf("[Main] Received: %s\n", val)
	}

	// 3. Demonstrating Deadlock (Important Concept)
	fmt.Println("\n--- 3. Understanding Deadlocks ---")
	fmt.Println("If you try to send to or receive from an unbuffered channel in the same")
	fmt.Println("goroutine without a companion goroutine, Go detects it as a deadlock at runtime.")
	fmt.Println("Example:")
	fmt.Println("  ch := make(chan int)")
	fmt.Println("  ch <- 42 // Blocks forever! There is no other goroutine to receive it.")
	fmt.Println("  <-ch     // Blocks forever! There is no other goroutine to send to it.")
	fmt.Println("\nTutorial 03 complete!")
}
