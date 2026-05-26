package main

import (
	"unicode"
)

// ==================================================
// MAPS (Questions 21 - 25)
// ==================================================

// 21. BuildWordFrequencyCounter builds a word frequency counter using a map.
func BuildWordFrequencyCounter(text string) map[string]int {
	freq := make(map[string]int)
	var currentWord []rune

	addWord := func() {
		if len(currentWord) > 0 {
			freq[string(currentWord)]++
			currentWord = nil
		}
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord = append(currentWord, unicode.ToLower(r))
		} else {
			addWord()
		}
	}
	addWord()

	return freq
}

// 22. FindDuplicates finds all duplicate elements in a slice using a map.
// It returns the unique elements that appeared more than once.
func FindDuplicates(s []int) []int {
	counts := make(map[int]int)
	for _, v := range s {
		counts[v]++
	}

	duplicates := make([]int, 0)
	// Maintain insertion order based on the original slice to make it deterministic
	seenInDuplicates := make(map[int]bool)
	for _, v := range s {
		if counts[v] > 1 && !seenInDuplicates[v] {
			seenInDuplicates[v] = true
			duplicates = append(duplicates, v)
		}
	}
	return duplicates
}

// 23. Set is a generic Set implementation using map[T]struct{}.
type Set[T comparable] struct {
	data map[T]struct{}
}

// NewSet creates a new Set instance.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}),
	}
}

// Add adds an element to the set.
func (s *Set[T]) Add(val T) {
	s.data[val] = struct{}{}
}

// Remove removes an element from the set.
func (s *Set[T]) Remove(val T) {
	delete(s.data, val)
}

// Contains checks if the set contains the element.
func (s *Set[T]) Contains(val T) bool {
	_, exists := s.data[val]
	return exists
}

// Size returns the number of elements in the set.
func (s *Set[T]) Size() int {
	return len(s.data)
}

// Elements returns a slice containing all elements in the set.
func (s *Set[T]) Elements() []T {
	elements := make([]T, 0, len(s.data))
	for k := range s.data {
		elements = append(elements, k)
	}
	return elements
}

// 24. MergeMaps merges two maps. If key collisions occur, the value from the second map (m2) overwrites m1.
func MergeMaps[K comparable, V any](m1, m2 map[K]V) map[K]V {
	merged := make(map[K]V)
	for k, v := range m1 {
		merged[k] = v
	}
	for k, v := range m2 {
		merged[k] = v
	}
	return merged
}

// 25. InvertMap inverts a map where the original values become the keys.
// Since multiple keys can map to the same value, it maps each value to a slice of keys to prevent data loss.
func InvertMap[K comparable, V comparable](m map[K]V) map[V][]K {
	inverted := make(map[V][]K)
	for k, v := range m {
		inverted[v] = append(inverted[v], k)
	}
	return inverted
}
