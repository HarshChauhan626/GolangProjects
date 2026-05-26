package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 09: FAN-IN PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// Fan-In is the inverse of Fan-Out. Multiple goroutines each produce data
// on their own channels, and a single "fan-in" function merges all those
// channels into one unified output channel. The consumer reads from this
// single merged channel.
//
// USE CASES:
// - Aggregating results from parallel API calls
// - Merging log streams from multiple services
// - Combining search results from different backends
//
// ARCHITECTURE:
//
//   Source 1 ──▶ ch1 ──┐
//                      │
//   Source 2 ──▶ ch2 ──┼──▶ mergedCh ──▶ Consumer
//                      │
//   Source 3 ──▶ ch3 ──┘
//

// dataSource simulates a goroutine that produces data on its own channel.
// Each source produces values at its own pace, then closes the channel.
func dataSource(name string, count int) <-chan string {
	// Each source owns its own channel — this is a key Fan-In characteristic.
	out := make(chan string)

	go func() {
		defer close(out)
		for i := 1; i <= count; i++ {
			// Simulate variable latency
			time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)
			msg := fmt.Sprintf("[%s] message %d", name, i)
			out <- msg
		}
	}()

	return out
}

// fanIn merges multiple input channels into a single output channel.
// It spawns one goroutine per input channel to forward messages.
// The output channel is closed only after ALL inputs are exhausted.
func fanIn(channels ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	// For each input channel, start a goroutine that copies values to 'merged'.
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan string) {
			defer wg.Done()
			for msg := range c {
				merged <- msg
			}
		}(ch)
	}

	// Once all input channels are drained, close the merged channel.
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 09: FAN-IN PATTERN               ")
	fmt.Println("==================================================")

	// 1. Create multiple independent data sources.
	// Each returns its own channel.
	fmt.Println("Starting 3 independent data sources...\n")
	source1 := dataSource("Database", 3)
	source2 := dataSource("API", 4)
	source3 := dataSource("Cache", 2)

	// 2. Fan-In: merge all source channels into one.
	merged := fanIn(source1, source2, source3)

	// 3. Consume from the single merged channel.
	// Messages arrive in whatever order they are produced (non-deterministic).
	fmt.Println("Reading from merged channel:")
	count := 0
	for msg := range merged {
		count++
		fmt.Printf("  #%d received: %s\n", count, msg)
	}

	fmt.Printf("\nTotal messages received: %d\n", count)
	fmt.Println("All sources exhausted. Fan-In demo complete!")
	fmt.Println("Tutorial 09 complete!")
}
