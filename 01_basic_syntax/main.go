package main

import (
	"errors"
	"fmt"
	"math"
)

// Define a struct representing a 2D Shape
type Circle struct {
	Radius float64
}

// Define a struct representing a Rectangle
type Rectangle struct {
	Width, Height float64
}

// Define an Interface
// In Go, interfaces are implemented implicitly.
// Any type that implements Area() float64 automatically satisfies the Shape interface.
type Shape interface {
	Area() float64
}

// Method with a Value Receiver for Circle
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Method with a Value Receiver for Rectangle
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Method with a Pointer Receiver to modify the Circle's radius
// We use a pointer receiver (*Circle) when we want to modify the struct's internal state.
func (c *Circle) Scale(factor float64) {
	c.Radius = c.Radius * factor
}

// Function showing multiple return values and error handling
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("       TUTORIAL 01: BASIC GOLANG SYNTAX          ")
	fmt.Println("==================================================")

	// 1. Variables and Constants
	fmt.Println("\n--- 1. Variables and Constants ---")
	var explicitInt int = 42
	typeInferred := "Hello Go!" // Short variable declaration (only inside functions)
	const Pi = 3.14159

	fmt.Printf("Explicit Int: %d (type: %T)\n", explicitInt, explicitInt)
	fmt.Printf("Inferred String: %s (type: %T)\n", typeInferred, typeInferred)
	fmt.Printf("Constant Pi: %f\n", Pi)

	// 2. Control Structures: If-Else
	fmt.Println("\n--- 2. Control Structures: If-Else ---")
	num := 10
	if num%2 == 0 {
		fmt.Printf("%d is even\n", num)
	} else {
		fmt.Printf("%d is odd\n", num)
	}

	// Go allows initializing a statement inside 'if' before the conditional check
	if val := 5 * 2; val > 8 {
		fmt.Printf("Val %d is greater than 8 (val is only scoped inside if/else block)\n", val)
	}

	// 3. Control Structures: Switch
	fmt.Println("\n--- 3. Control Structures: Switch ---")
	os := "windows"
	switch os {
	case "darwin":
		fmt.Println("OS X.")
	case "linux":
		fmt.Println("Linux.")
	case "windows":
		fmt.Println("Windows.")
	default:
		fmt.Println("Other OS.")
	}

	// 4. Loops (Go only has 'for')
	fmt.Println("\n--- 4. Loops ---")
	// Standard C-like for loop
	fmt.Print("Standard loop: ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// While-like for loop
	fmt.Print("While-like loop: ")
	count := 0
	for count < 3 {
		fmt.Printf("%d ", count)
		count++
	}
	fmt.Println()

	// 5. Functions, Multiple Returns, and Errors
	fmt.Println("\n--- 5. Functions & Error Handling ---")
	result, err := divide(10, 2)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		fmt.Printf("10 / 2 = %f\n", result)
	}

	// Handling an error case
	badResult, err := divide(10, 0)
	if err != nil {
		fmt.Printf("Expected error occurred: %s (Result: %f)\n", err, badResult)
	}

	// 6. Structs, Methods, and Interfaces
	fmt.Println("\n--- 6. Structs, Methods & Interfaces ---")
	c := Circle{Radius: 5}
	r := Rectangle{Width: 4, Height: 5}

	fmt.Printf("Initial Circle Area: %.2f\n", c.Area())
	fmt.Printf("Rectangle Area: %.2f\n", r.Area())

	// Call pointer receiver method to modify Circle state
	fmt.Println("Scaling Circle by factor of 2...")
	c.Scale(2.0)
	fmt.Printf("New Circle Area (scaled): %.2f (Radius: %.2f)\n", c.Area(), c.Radius)

	// Interfaces: Using a slice of the Shape interface
	shapes := []Shape{c, r}
	fmt.Println("Iterating through shapes polymorphically via interface:")
	for idx, shape := range shapes {
		fmt.Printf("Shape %d Area: %.2f (Concrete Type: %T)\n", idx+1, shape.Area(), shape)
	}
}
