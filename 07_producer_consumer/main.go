package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 07: PRODUCER-CONSUMER PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// The Producer-Consumer pattern decouples data production from data
// consumption using a shared buffer (channel). Multiple producers generate
// items independently, while multiple consumers process them concurrently.
// The channel acts as a thread-safe queue between the two groups.
//
// KEY CONCEPTS:
// - Multiple goroutines producing data into a shared channel
// - Multiple goroutines consuming data from the same channel
// - sync.WaitGroup to coordinate shutdown of producers and consumers
// - Closing the channel signals consumers that no more data will arrive
//
// ARCHITECTURE:
//
//   Producer 1 ──┐                  ┌── Consumer 1
//   Producer 2 ──┼──▶ [Channel] ──▶─┼── Consumer 2
//   Producer 3 ──┘                  └── Consumer 3
//

// Item represents a unit of data flowing through the system.
type Item struct {
	ID         int
	Value      string
	ProducerID int
}

// producer generates items and sends them into the shared channel.
// It produces 'count' items, then returns. The WaitGroup tracks completion.
func producer(id int, count int, out chan<- Item, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < count; i++ {
		item := Item{
			ID:         id*100 + i, // Unique ID per producer
			Value:      fmt.Sprintf("data-%d-%d", id, i),
			ProducerID: id,
		}

		// Simulate variable production time
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		out <- item
		fmt.Printf("[Producer %d] Produced item %d: %s\n", id, item.ID, item.Value)
	}

	fmt.Printf("[Producer %d] Finished producing.\n", id)
}

// consumer reads items from the shared channel and processes them.
// The for-range loop exits automatically when the channel is closed.
func consumer(id int, in <-chan Item, wg *sync.WaitGroup) {
	defer wg.Done()

	for item := range in {
		// Simulate variable processing time
		time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)
		fmt.Printf("  [Consumer %d] Consumed item %d (from Producer %d): %s\n",
			id, item.ID, item.ProducerID, item.Value)
	}

	fmt.Printf("  [Consumer %d] Channel closed. Exiting.\n", id)
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("    TUTORIAL 07: PRODUCER-CONSUMER PATTERN         ")
	fmt.Println("==================================================")

	numProducers := 3
	numConsumers := 2
	itemsPerProducer := 4

	// Buffered channel acts as the shared queue between producers and consumers.
	// Buffer size allows producers to continue without blocking immediately.
	queue := make(chan Item, 10)

	// Separate WaitGroups for producers and consumers,
	// because we need to close the channel AFTER all producers are done,
	// but BEFORE waiting for consumers to finish.
	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup

	// 1. Start consumers first so they're ready to receive
	fmt.Printf("Starting %d consumers...\n", numConsumers)
	for c := 1; c <= numConsumers; c++ {
		consumerWg.Add(1)
		go consumer(c, queue, &consumerWg)
	}

	// 2. Start producers
	fmt.Printf("Starting %d producers (each producing %d items)...\n\n", numProducers, itemsPerProducer)
	for p := 1; p <= numProducers; p++ {
		producerWg.Add(1)
		go producer(p, itemsPerProducer, queue, &producerWg)
	}

	// 3. Wait for all producers to finish, then close the channel.
	// Closing the channel signals to consumers that no more items will arrive.
	producerWg.Wait()
	fmt.Println("\nAll producers finished. Closing queue channel...")
	close(queue)

	// 4. Wait for all consumers to drain remaining items and exit.
	consumerWg.Wait()

	fmt.Println("\nAll consumers finished. Producer-Consumer demo complete!")
	fmt.Println("Tutorial 07 complete!")
}
