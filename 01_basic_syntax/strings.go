package main

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// ==================================================
// STRINGS (Questions 11 - 20)
// ==================================================

// 11. ReverseString reverses a string (correctly handling multi-byte UTF-8 runes).
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// 12. IsPalindrome checks if a string is a palindrome, ignoring non-alphanumeric characters and case.
func IsPalindrome(s string) bool {
	var cleaned []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cleaned = append(cleaned, unicode.ToLower(r))
		}
	}

	n := len(cleaned)
	for i := 0; i < n/2; i++ {
		if cleaned[i] != cleaned[n-1-i] {
			return false
		}
	}
	return true
}

// 13. CharFrequency counts the frequency of each character (rune) in a string.
func CharFrequency(s string) map[rune]int {
	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}
	return freq
}

// 14. CountWordFrequency counts the frequency of each word in a string, normalizing to lowercase and ignoring punctuation.
func CountWordFrequency(s string) map[string]int {
	freq := make(map[string]int)
	var currentWord []rune

	addWord := func() {
		if len(currentWord) > 0 {
			freq[string(currentWord)]++
			currentWord = nil
		}
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord = append(currentWord, unicode.ToLower(r))
		} else {
			addWord()
		}
	}
	addWord()

	return freq
}

// 15. FindFirstNonRepeatingChar finds the first non-repeating character (rune) in a string.
func FindFirstNonRepeatingChar(s string) (rune, error) {
	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}

	for _, r := range s {
		if freq[r] == 1 {
			return r, nil
		}
	}
	return 0, errors.New("no non-repeating character found")
}

// 16. AreAnagrams checks whether two strings are anagrams of each other (case-insensitive, ignoring non-alphanumeric).
func AreAnagrams(s1, s2 string) bool {
	normalize := func(s string) map[rune]int {
		freq := make(map[rune]int)
		for _, r := range s {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				freq[unicode.ToLower(r)]++
			}
		}
		return freq
	}

	f1, f2 := normalize(s1), normalize(s2)
	if len(f1) != len(f2) {
		return false
	}

	for r, count := range f1 {
		if f2[r] != count {
			return false
		}
	}
	return true
}

// 17. RemoveDuplicateChars removes duplicate characters from a string, preserving original order of first occurrence.
func RemoveDuplicateChars(s string) string {
	seen := make(map[rune]bool)
	var builder strings.Builder
	for _, r := range s {
		if !seen[r] {
			seen[r] = true
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// 18. CompressString compresses a string (e.g., aaabbc → a3b2c1).
func CompressString(s string) string {
	if len(s) == 0 {
		return ""
	}

	runes := []rune(s)
	var builder strings.Builder
	currentChar := runes[0]
	count := 1

	for i := 1; i < len(runes); i++ {
		if runes[i] == currentChar {
			count++
		} else {
			builder.WriteRune(currentChar)
			builder.WriteString(strconv.Itoa(count))
			currentChar = runes[i]
			count = 1
		}
	}
	builder.WriteRune(currentChar)
	builder.WriteString(strconv.Itoa(count))

	return builder.String()
}

// 19. LongestCommonPrefix finds the longest common prefix among an array of strings.
func LongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for _, str := range strs[1:] {
		for !strings.HasPrefix(str, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			runes := []rune(prefix)
			prefix = string(runes[:len(runes)-1])
		}
	}
	return prefix
}

// 20. ToTitleCase converts a sentence to title case.
func ToTitleCase(s string) string {
	var builder strings.Builder
	inWord := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			builder.WriteRune(r)
			inWord = false
		} else {
			if !inWord {
				builder.WriteRune(unicode.ToUpper(r))
				inWord = true
			} else {
				builder.WriteRune(unicode.ToLower(r))
			}
		}
	}
	return builder.String()
}
