package main

// ==================================================
// FUNCTIONS & POINTERS (Questions 31 - 35)
// ==================================================

// 31. VariadicSum calculates the sum of zero or more integer arguments.
func VariadicSum(nums ...int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

// 32. Factorial calculates n! using recursion.
// Returns 0 for negative numbers.
func Factorial(n int) int {
	if n < 0 {
		return 0
	}
	if n == 0 || n == 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// 33. Fibonacci calculates the n-th Fibonacci number using recursion.
// Returns 0 for negative index inputs.
func Fibonacci(n int) int {
	if n < 0 {
		return 0
	}
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

// 34. Swap swaps the values of two integers using their pointer addresses.
func Swap(a, b *int) {
	*a, *b = *b, *a
}

// 35. NewCounter returns a closure function that maintains and increments an internal counter.
func NewCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}
