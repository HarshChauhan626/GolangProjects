package main

import (
	"fmt"
	"math"
	"time"
)

// ==================================================
// TUTORIAL 10: PIPELINE PROCESSING
// ==================================================
//
// PROBLEM STATEMENT:
// A pipeline is a series of stages connected by channels, where each stage
// is a group of goroutines running the same function. Each stage:
//   1. Receives values from an upstream channel
//   2. Performs some processing on those values
//   3. Sends the results to a downstream channel
//
// The pipeline allows stages to run concurrently — while stage 2 processes
// item N, stage 1 can already be processing item N+1.
//
// ARCHITECTURE:
//
//   generate → [ch1] → square → [ch2] → filterEven → [ch3] → consumer
//

// generate produces integers from 1..max and sends them downstream.
// This is the first stage of the pipeline (source).
func generate(max int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= max; i++ {
			out <- i
		}
	}()
	return out
}

// square reads integers from the input channel, squares them,
// and sends the results downstream. This is a transformation stage.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// filterEven passes through only even numbers.
// This is a filtering stage that reduces the data flowing downstream.
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

// addLabel converts integers to labeled strings.
// Demonstrates that pipeline stages can change the data type.
func addLabel(in <-chan int) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for n := range in {
			sqrt := math.Sqrt(float64(n))
			out <- fmt.Sprintf("√%d = %.2f", n, sqrt)
		}
	}()
	return out
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      TUTORIAL 10: PIPELINE PROCESSING             ")
	fmt.Println("==================================================")

	start := time.Now()

	// Build the pipeline by chaining stages.
	// Data flows:  generate → square → filterEven → addLabel → consumer
	//
	// Each function returns a channel that the next stage reads from.
	// The pipeline is fully concurrent — all stages run simultaneously.
	fmt.Println("Pipeline: generate(1..20) → square → filterEven → addLabel\n")

	stage1 := generate(20)          // Produces: 1, 2, 3, ..., 20
	stage2 := square(stage1)        // Produces: 1, 4, 9, ..., 400
	stage3 := filterEven(stage2)    // Keeps:    4, 16, 36, 64, 100, ...
	stage4 := addLabel(stage3)      // Formats:  "√4 = 2.00", "√16 = 4.00", ...

	// Consume the final output
	count := 0
	for result := range stage4 {
		count++
		fmt.Printf("  Stage 4 output #%d: %s\n", count, result)
	}

	elapsed := time.Since(start)
	fmt.Printf("\nPipeline processed %d final results in %v\n", count, elapsed)
	fmt.Println("Tutorial 10 complete!")
}
