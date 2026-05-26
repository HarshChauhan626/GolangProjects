package main

import (
	"errors"
	"fmt"
)

// divide is a helper function for the basic tutorial (Part 5)
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil
}

func main() {
	fmt.Println("==================================================================")
	fmt.Println("            TUTORIAL 01: BASIC GOLANG SYNTAX & EXERCISES          ")
	fmt.Println("==================================================================")

	// ==================================================
	// PART A: ORIGINAL BASIC TUTORIAL
	// ==================================================
	fmt.Println("\n>>> PART A: GOLANG LANGUAGE BASICS <<<")

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
	fmt.Print("Standard loop (0..4): ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	fmt.Print("While-like loop (0..2): ")
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

	badResult, err := divide(10, 0)
	if err != nil {
		fmt.Printf("Expected error occurred: %s (Result: %f)\n", err, badResult)
	}

	// ==================================================
	// PART B: 40 EXERCISE ANSWERS
	// ==================================================
	fmt.Println("\n>>> PART B: 40 SYNTAX EXERCISE ANSWERS <<<")

	// --------------------------------------------------
	// SLICES (1 - 10)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" SLICES (Questions 1 - 10)")
	fmt.Println("==================================================")

	// 1. Reverse a slice in-place
	s1 := []int{1, 2, 3, 4, 5}
	fmt.Printf("[1] Reverse Slice (in-place):\n    Before: %v\n", s1)
	ReverseSlice(s1)
	fmt.Printf("    After:  %v\n", s1)

	// 2. Find the largest and smallest element
	s2 := []int{5, 2, 9, 1, 7}
	minVal, maxVal, _ := FindMinMax(s2)
	fmt.Printf("[2] Min & Max of %v:\n    Min: %d, Max: %d\n", s2, minVal, maxVal)

	// 3. Find the second largest element
	s3 := []int{5, 2, 9, 1, 9, 7}
	secLargest, _ := FindSecondLargest(s3)
	fmt.Printf("[3] Second Largest (distinct) of %v:\n    Result: %d\n", s3, secLargest)

	// 4. Rotate a slice by K positions
	s4 := []int{1, 2, 3, 4, 5}
	fmt.Printf("[4] Rotate %v by 2 positions:\n", s4)
	RotateSlice(s4, 2)
	fmt.Printf("    Result: %v\n", s4)

	// 5. Remove duplicates
	s5 := []int{1, 2, 2, 3, 1, 4}
	fmt.Printf("[5] Remove Duplicates from %v:\n    Result: %v\n", s5, RemoveDuplicates(s5))

	// 6. Merge two sorted slices
	a6, b6 := []int{1, 3, 5}, []int{2, 4, 6}
	fmt.Printf("[6] Merge sorted %v and %v:\n    Result: %v\n", a6, b6, MergeSortedSlices(a6, b6))

	// 7. Find intersection of two slices
	a7, b7 := []int{1, 2, 2, 3}, []int{2, 2, 4}
	fmt.Printf("[7] Intersection of %v and %v:\n    Result: %v\n", a7, b7, FindIntersection(a7, b7))

	// 8. Move all zeroes to the end
	s8 := []int{0, 1, 0, 3, 12}
	fmt.Printf("[8] Move Zeroes to End for %v:\n", s8)
	MoveZeroesToEnd(s8)
	fmt.Printf("    Result: %v\n", s8)

	// 9. Partition even and odd numbers
	s9 := []int{1, 2, 3, 4, 5}
	evens, odds := PartitionEvenOdd(s9)
	fmt.Printf("[9] Partition Even/Odd for %v:\n    Evens: %v, Odds: %v\n", s9, evens, odds)

	// 10. Generic Contains() function
	fruitSlice := []string{"apple", "banana", "cherry"}
	hasBanana := Contains(fruitSlice, "banana")
	hasGrape := Contains(fruitSlice, "grape")
	fmt.Printf("[10] Generic Contains:\n     %v contains 'banana'? %t\n     %v contains 'grape'? %t\n", fruitSlice, hasBanana, fruitSlice, hasGrape)

	// --------------------------------------------------
	// STRINGS (11 - 20)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" STRINGS (Questions 11 - 20)")
	fmt.Println("==================================================")

	// 11. Reverse a string
	str11 := "Go ♥ UTF-8"
	fmt.Printf("[11] Reverse string:\n     %q → %q\n", str11, ReverseString(str11))

	// 12. Check if a string is a palindrome
	str12_1 := "A man, a plan, a canal: Panama"
	str12_2 := "hello"
	fmt.Printf("[12] Palindrome check:\n     %q is palindrome? %t\n     %q is palindrome? %t\n", str12_1, IsPalindrome(str12_1), str12_2, IsPalindrome(str12_2))

	// 13. Count frequency of each character
	str13 := "hello"
	fmt.Printf("[13] Character frequencies in %q:\n", str13)
	for char, count := range CharFrequency(str13) {
		fmt.Printf("     '%c': %d\n", char, count)
	}

	// 14. Count frequency of each word
	str14 := "Go is great! Go is fast."
	fmt.Printf("[14] Word frequencies in %q:\n", str14)
	for word, count := range CountWordFrequency(str14) {
		fmt.Printf("     %q: %d\n", word, count)
	}

	// 15. Find first non-repeating character
	str15 := "swiss"
	r15, _ := FindFirstNonRepeatingChar(str15)
	fmt.Printf("[15] First non-repeating char in %q:\n     Result: '%c'\n", str15, r15)

	// 16. Check anagrams
	str16_1, str16_2 := "listen", "silent"
	fmt.Printf("[16] Are Anagrams (%q, %q)? %t\n", str16_1, str16_2, AreAnagrams(str16_1, str16_2))

	// 17. Remove duplicate characters
	str17 := "google"
	fmt.Printf("[17] Remove duplicate characters from %q:\n     Result: %q\n", str17, RemoveDuplicateChars(str17))

	// 18. Compress string
	str18 := "aaabbc"
	fmt.Printf("[18] Compress %q:\n     Result: %q\n", str18, CompressString(str18))

	// 19. Find longest common prefix
	strs19 := []string{"flower", "flow", "flight"}
	fmt.Printf("[19] Longest common prefix of %v:\n     Result: %q\n", strs19, LongestCommonPrefix(strs19))

	// 20. Convert sentence to title case
	str20 := "hello WORLD, this IS go!"
	fmt.Printf("[20] Sentence to Title Case:\n     %q → %q\n", str20, ToTitleCase(str20))

	// --------------------------------------------------
	// MAPS (21 - 25)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" MAPS (Questions 21 - 25)")
	fmt.Println("==================================================")

	// 21. Word frequency counter using map
	str21 := "apple banana apple"
	fmt.Printf("[21] Build word frequency counter using map:\n     Text: %q\n     Result: %v\n", str21, BuildWordFrequencyCounter(str21))

	// 22. Find duplicates using map
	s22 := []int{1, 2, 3, 2, 4, 5, 1}
	fmt.Printf("[22] Find duplicate elements in %v:\n     Result: %v\n", s22, FindDuplicates(s22))

	// 23. Set using map[T]struct{}
	set := NewSet[string]()
	set.Add("apple")
	set.Add("banana")
	set.Add("apple") // duplicate addition
	fmt.Printf("[23] Set implementation:\n     Added: 'apple', 'banana', 'apple'\n     Set elements: %v (Size: %d)\n", set.Elements(), set.Size())
	fmt.Printf("     Contains 'apple'? %t\n", set.Contains("apple"))
	set.Remove("apple")
	fmt.Printf("     After removing 'apple', Set elements: %v\n", set.Elements())

	// 24. Merge two maps
	m24_1 := map[string]int{"a": 1, "b": 2}
	m24_2 := map[string]int{"b": 3, "c": 4}
	fmt.Printf("[24] Merge maps %v and %v:\n     Result: %v\n", m24_1, m24_2, MergeMaps(m24_1, m24_2))

	// 25. Invert a map
	m25 := map[string]int{"apple": 1, "banana": 2, "cherry": 1}
	fmt.Printf("[25] Invert map %v:\n     Result: %v\n", m25, InvertMap(m25))

	// --------------------------------------------------
	// STRUCTS & METHODS (26 - 30)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" STRUCTS & METHODS (Questions 26 - 30)")
	fmt.Println("==================================================")

	// 26 & 27. User struct, methods, String() method
	user := User{ID: 101, Username: "gopher", Email: "gopher@golang.org", IsActive: false}
	fmt.Printf("[26 & 27] User struct with String() method:\n     Initial: %s\n", user)
	user.Activate()
	fmt.Printf("     After user.Activate():   %s\n", user)
	user.UpdateEmail("new_gopher@golang.org")
	fmt.Printf("     After user.UpdateEmail():%s\n", user)

	// 28. Rectangle Struct area & perimeter
	rect := Rectangle{Width: 10, Height: 5}
	fmt.Printf("[28] Rectangle %v:\n     Area: %.2f, Perimeter: %.2f\n", rect, rect.Area(), rect.Perimeter())

	// 29. Pointer vs Value receiver
	rect2 := Rectangle{Width: 3, Height: 4}
	fmt.Printf("[29] Pointer Receiver vs Value Receiver:\n     Initial rect: %v\n", rect2)
	rect2Val := rect2.ScaleVal(2.0)
	fmt.Printf("     After ScaleVal(2.0) (value receiver) -> rect: %v, returned rect: %v\n", rect2, rect2Val)
	rect2.ScalePtr(2.0)
	fmt.Printf("     After ScalePtr(2.0) (pointer receiver) -> rect: %v\n", rect2)

	// 30. Nested Structs for Address Book
	ab := NewAddressBook()
	c := Contact{
		Name:  "Harsh Chauhan",
		Phone: "+91-99999-99999",
		Address: Address{
			Street:  "123 Tech Lane",
			City:    "New Delhi",
			State:   "Delhi",
			ZipCode: "110001",
		},
	}
	ab.AddContact(c)
	fmt.Printf("[30] Address Book (Nested Structs):\n     Added contact: %s, Phone: %s, City: %s\n", ab.ListContacts()[0].Name, ab.ListContacts()[0].Phone, ab.ListContacts()[0].Address.City)
	searchRes, _ := ab.SearchByName("Harsh")
	fmt.Printf("     Search for 'Harsh' found: %d contact(s)\n", len(searchRes))

	// --------------------------------------------------
	// FUNCTIONS & POINTERS (31 - 35)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" FUNCTIONS & POINTERS (Questions 31 - 35)")
	fmt.Println("==================================================")

	// 31. Variadic Sum
	fmt.Printf("[31] Variadic Sum of (1, 2, 3, 4, 5):\n     Result: %d\n", VariadicSum(1, 2, 3, 4, 5))

	// 32. Factorial recursion
	fmt.Printf("[32] Recursion - Factorial(5):\n     Result: %d\n", Factorial(5))

	// 33. Fibonacci recursion
	fmt.Printf("[33] Recursion - Fibonacci(7):\n     Result: %d\n", Fibonacci(7))

	// 34. Swap variables using pointers
	valA, valB := 100, 200
	fmt.Printf("[34] Swap variables using pointers:\n     Before: A = %d, B = %d\n", valA, valB)
	Swap(&valA, &valB)
	fmt.Printf("     After:  A = %d, B = %d\n", valA, valB)

	// 35. Closure maintaining counter
	counter := NewCounter()
	fmt.Printf("[35] Closure maintaining counter:\n     1st call: %d\n     2nd call: %d\n     3rd call: %d\n", counter(), counter(), counter())

	// --------------------------------------------------
	// INTERFACES (36 - 40)
	// --------------------------------------------------
	fmt.Println("\n==================================================")
	fmt.Println(" INTERFACES (Questions 36 - 40)")
	fmt.Println("==================================================")

	// 36. Shape interface with Circle and Rectangle
	var sh1 Shape = Circle{Radius: 5}
	var sh2 Shape = Rectangle{Width: 4, Height: 5}
	fmt.Printf("[36] Shape Interface polymorph:\n     Circle Area: %.2f, Perimeter: %.2f\n     Rectangle Area: %.2f, Perimeter: %.2f\n", sh1.Area(), sh1.Perimeter(), sh2.Area(), sh2.Perimeter())

	// 37. PaymentProcessor interface
	var payProcessor PaymentProcessor
	payProcessor = CreditCardProcessor{CardNumber: "1111222233334444", CardHolder: "Harsh Chauhan"}
	payMsg, _ := payProcessor.ProcessPayment(250.00)
	fmt.Printf("[37] Payment Processor:\n     %s\n", payMsg)
	payProcessor = PayPalProcessor{Email: "harsh@example.com"}
	payMsg, _ = payProcessor.ProcessPayment(120.50)
	fmt.Printf("     %s\n", payMsg)

	// 38. Logger interface with multiple implementations
	var logger Logger
	logger = ConsoleLogger{Prefix: "APP-DEBUG"}
	fmt.Print("[38] Logger interface console test:\n     ")
	logger.Log("This is a log message printed directly to stdout.")

	bufLogger := &BufferLogger{}
	logger = bufLogger
	logger.Log("Log entry #1 stored in memory.")
	logger.Log("Log entry #2 stored in memory.")
	fmt.Printf("     Buffer Logger stored: %v\n", bufLogger.Logs)

	// 39. Type assertions
	fmt.Printf("[39] Type assertions:\n     %s\n     %s\n", IdentifyShape(sh1), IdentifyShape(sh2))

	// 40. Type switch
	fmt.Printf("[40] Type switch:\n     %s\n     %s\n     %s\n     %s\n", HandleType(42), HandleType("Go is awesome"), HandleType(true), HandleType(sh1))

	fmt.Println("\n==================================================")
	fmt.Println("Tutorial 01 exercises complete!")
	fmt.Println("==================================================")
}
