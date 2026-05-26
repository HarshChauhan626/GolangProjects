package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("       TUTORIAL 05: SELECT MULTIPLEXING           ")
	fmt.Println("==================================================")

	ch1 := make(chan string)
	ch2 := make(chan string)

	// 1. Multiplexing multiple channel reads
	fmt.Println("\n--- 1. Multiplexing Multiple Channels ---")
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "Message from Channel 1"
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		ch2 <- "Message from Channel 2"
	}()

	// We have 2 messages to receive, so we loop twice
	for i := 0; i < 2; i++ {
		// select blocks until one of its cases is ready to execute
		select {
		case msg1 := <-ch1:
			fmt.Printf("Received: %s\n", msg1)
		case msg2 := <-ch2:
			fmt.Printf("Received: %s\n", msg2)
		}
	}

	// 2. Implementing a Timeout using select
	fmt.Println("\n--- 2. Implementing a Timeout ---")
	slowChan := make(chan string)

	go func() {
		// This worker takes 2 seconds to send
		time.Sleep(2 * time.Second)
		slowChan <- "Slow Response"
	}()

	select {
	case res := <-slowChan:
		fmt.Printf("Received response: %s\n", res)
	case <-time.After(500 * time.Millisecond):
		// time.After returns a channel that sends the current time after the duration
		fmt.Println("Timeout! The worker was too slow (exceeded 500ms limit).")
	}

	// 3. Non-blocking Channel Operations using 'default'
	fmt.Println("\n--- 3. Non-Blocking Channel Operations ---")
	messageChan := make(chan string) // Unbuffered, no sender ready

	// A normal receive <-messageChan would block.
	// But select with a default case makes it non-blocking.
	select {
	case msg := <-messageChan:
		fmt.Printf("Received message: %s\n", msg)
	default:
		fmt.Println("No message received (non-blocking).")
	}

	// Similarly, a non-blocking send:
	select {
	case messageChan <- "hello":
		fmt.Println("Message sent!")
	default:
		fmt.Println("No receiver ready, skipped sending (non-blocking).")
	}

	fmt.Println("\nTutorial 05 complete!")
}
