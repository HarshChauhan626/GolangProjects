package main

import (
	"errors"
)

// ==================================================
// SLICES (Questions 1 - 10)
// ==================================================

// 1. ReverseSlice reverses a slice in-place.
func ReverseSlice(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// 2. FindMinMax finds the largest and smallest element in a slice.
func FindMinMax(s []int) (min int, max int, err error) {
	if len(s) == 0 {
		return 0, 0, errors.New("cannot find min/max of an empty slice")
	}
	min, max = s[0], s[0]
	for _, v := range s[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max, nil
}

// 3. FindSecondLargest finds the second largest distinct element in a slice.
func FindSecondLargest(s []int) (int, error) {
	if len(s) < 2 {
		return 0, errors.New("slice must have at least 2 elements")
	}

	var first, second int
	hasFirst, hasSecond := false, false

	for _, v := range s {
		if !hasFirst {
			first = v
			hasFirst = true
		} else if v > first {
			second = first
			hasSecond = true
			first = v
		} else if v < first {
			if !hasSecond || v > second {
				second = v
				hasSecond = true
			}
		}
	}

	if !hasSecond {
		return 0, errors.New("no second largest distinct element exists in the slice")
	}

	return second, nil
}

// 4. RotateSlice rotates a slice by K positions to the right in-place.
func RotateSlice(s []int, k int) {
	n := len(s)
	if n <= 1 {
		return
	}
	k = k % n
	if k < 0 {
		k = k + n
	}
	if k == 0 {
		return
	}

	// Helper to reverse sub-slices in-place
	reverse := func(arr []int) {
		for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	// Right rotation: reverse whole, then first k, then rest
	reverse(s)
	reverse(s[:k])
	reverse(s[k:])
}

// 5. RemoveDuplicates removes duplicate elements from a slice while preserving order.
func RemoveDuplicates(s []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0)
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// 6. MergeSortedSlices merges two pre-sorted slices into a single sorted slice.
func MergeSortedSlices(a, b []int) []int {
	result := make([]int, 0, len(a)+len(b))
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}

	if i < len(a) {
		result = append(result, a[i:]...)
	}
	if j < len(b) {
		result = append(result, b[j:]...)
	}

	return result
}

// 7. FindIntersection finds the common elements between two slices, returning distinct values.
func FindIntersection(a, b []int) []int {
	seen := make(map[int]bool)
	for _, v := range a {
		seen[v] = true
	}

	result := make([]int, 0)
	added := make(map[int]bool)
	for _, v := range b {
		if seen[v] && !added[v] {
			added[v] = true
			result = append(result, v)
		}
	}
	return result
}

// 8. MoveZeroesToEnd moves all zeroes to the end of a slice in-place while keeping relative order.
func MoveZeroesToEnd(s []int) {
	writeIndex := 0
	for _, v := range s {
		if v != 0 {
			s[writeIndex] = v
			writeIndex++
		}
	}
	for i := writeIndex; i < len(s); i++ {
		s[i] = 0
	}
}

// 9. PartitionEvenOdd partitions a slice into separate even and odd slices.
func PartitionEvenOdd(s []int) (evens []int, odds []int) {
	evens = make([]int, 0)
	odds = make([]int, 0)
	for _, v := range s {
		if v%2 == 0 {
			evens = append(evens, v)
		} else {
			odds = append(odds, v)
		}
	}
	return evens, odds
}

// 10. Contains returns true if the generic slice contains the target element.
func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}
