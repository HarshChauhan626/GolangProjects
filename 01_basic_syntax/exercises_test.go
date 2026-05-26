package main

import (
	"math"
	"reflect"
	"sort"
	"testing"
)

// ==================================================
// SLICE TESTS (1 - 10)
// ==================================================

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		input []int
		want  []int
	}{
		{[]int{}, []int{}},
		{[]int{1}, []int{1}},
		{[]int{1, 2}, []int{2, 1}},
		{[]int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
	}

	for _, tt := range tests {
		s := make([]int, len(tt.input))
		copy(s, tt.input)
		ReverseSlice(s)
		if !reflect.DeepEqual(s, tt.want) {
			t.Errorf("ReverseSlice(%v) = %v; want %v", tt.input, s, tt.want)
		}
	}
}

func TestFindMinMax(t *testing.T) {
	// Success cases
	tests := []struct {
		input   []int
		wantMin int
		wantMax int
	}{
		{[]int{1}, 1, 1},
		{[]int{5, 2, 9, 1, 7}, 1, 9},
		{[]int{-5, -10, 0, 10}, -10, 10},
	}
	for _, tt := range tests {
		min, max, err := FindMinMax(tt.input)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.input, err)
		}
		if min != tt.wantMin || max != tt.wantMax {
			t.Errorf("FindMinMax(%v) = (%d, %d); want (%d, %d)", tt.input, min, max, tt.wantMin, tt.wantMax)
		}
	}

	// Error case
	_, _, err := FindMinMax([]int{})
	if err == nil {
		t.Error("expected error for empty slice, got nil")
	}
}

func TestFindSecondLargest(t *testing.T) {
	tests := []struct {
		input   []int
		want    int
		wantErr bool
	}{
		{[]int{1, 2}, 1, false},
		{[]int{2, 1}, 1, false},
		{[]int{5, 2, 9, 1, 9, 7}, 7, false}, // distinct check
		{[]int{9, 9, 9}, 0, true},           // no second distinct
		{[]int{1}, 0, true},                 // too short
		{[]int{}, 0, true},                  // empty
	}

	for _, tt := range tests {
		got, err := FindSecondLargest(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("FindSecondLargest(%v) error status = %v; wantErr = %v", tt.input, err != nil, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("FindSecondLargest(%v) = %d; want %d", tt.input, got, tt.want)
		}
	}
}

func TestRotateSlice(t *testing.T) {
	tests := []struct {
		input []int
		k     int
		want  []int
	}{
		{[]int{1, 2, 3, 4, 5}, 2, []int{4, 5, 1, 2, 3}},
		{[]int{1, 2, 3, 4, 5}, 0, []int{1, 2, 3, 4, 5}},
		{[]int{1, 2, 3, 4, 5}, 5, []int{1, 2, 3, 4, 5}},
		{[]int{1, 2, 3, 4, 5}, 7, []int{4, 5, 1, 2, 3}},
		{[]int{1, 2, 3, 4, 5}, -1, []int{2, 3, 4, 5, 1}}, // Left rotation via negative k
		{[]int{}, 3, []int{}},
		{[]int{1}, 3, []int{1}},
	}

	for _, tt := range tests {
		s := make([]int, len(tt.input))
		copy(s, tt.input)
		RotateSlice(s, tt.k)
		if !reflect.DeepEqual(s, tt.want) {
			t.Errorf("RotateSlice(%v, %d) = %v; want %v", tt.input, tt.k, s, tt.want)
		}
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		input []int
		want  []int
	}{
		{[]int{1, 2, 2, 3, 1, 4}, []int{1, 2, 3, 4}},
		{[]int{}, []int{}},
		{[]int{1, 1, 1}, []int{1}},
	}

	for _, tt := range tests {
		got := RemoveDuplicates(tt.input)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("RemoveDuplicates(%v) = %v; want %v", tt.input, got, tt.want)
		}
	}
}

func TestMergeSortedSlices(t *testing.T) {
	tests := []struct {
		a    []int
		b    []int
		want []int
	}{
		{[]int{1, 3, 5}, []int{2, 4, 6}, []int{1, 2, 3, 4, 5, 6}},
		{[]int{}, []int{1, 2}, []int{1, 2}},
		{[]int{1, 2}, []int{}, []int{1, 2}},
		{[]int{1, 2, 2}, []int{2, 3}, []int{1, 2, 2, 2, 3}},
	}

	for _, tt := range tests {
		got := MergeSortedSlices(tt.a, tt.b)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("MergeSortedSlices(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFindIntersection(t *testing.T) {
	tests := []struct {
		a    []int
		b    []int
		want []int
	}{
		{[]int{1, 2, 2, 3}, []int{2, 2, 4}, []int{2}},
		{[]int{1, 2, 3}, []int{4, 5, 6}, []int{}},
		{[]int{}, []int{1}, []int{}},
	}

	for _, tt := range tests {
		got := FindIntersection(tt.a, tt.b)
		// Compare elements ignoring order since we want intersection set
		sort.Ints(got)
		sort.Ints(tt.want)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FindIntersection(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMoveZeroesToEnd(t *testing.T) {
	tests := []struct {
		input []int
		want  []int
	}{
		{[]int{0, 1, 0, 3, 12}, []int{1, 3, 12, 0, 0}},
		{[]int{0, 0}, []int{0, 0}},
		{[]int{1, 2, 3}, []int{1, 2, 3}},
		{[]int{}, []int{}},
	}

	for _, tt := range tests {
		s := make([]int, len(tt.input))
		copy(s, tt.input)
		MoveZeroesToEnd(s)
		if !reflect.DeepEqual(s, tt.want) {
			t.Errorf("MoveZeroesToEnd(%v) = %v; want %v", tt.input, s, tt.want)
		}
	}
}

func TestPartitionEvenOdd(t *testing.T) {
	tests := []struct {
		input     []int
		wantEvens []int
		wantOdds  []int
	}{
		{[]int{1, 2, 3, 4, 5}, []int{2, 4}, []int{1, 3, 5}},
		{[]int{2, 4}, []int{2, 4}, []int{}},
		{[]int{1, 3}, []int{}, []int{1, 3}},
		{[]int{}, []int{}, []int{}},
	}

	for _, tt := range tests {
		evens, odds := PartitionEvenOdd(tt.input)
		if !reflect.DeepEqual(evens, tt.wantEvens) || !reflect.DeepEqual(odds, tt.wantOdds) {
			t.Errorf("PartitionEvenOdd(%v) = (%v, %v); want (%v, %v)", tt.input, evens, odds, tt.wantEvens, tt.wantOdds)
		}
	}
}

func TestContains(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		if !Contains(s, 3) {
			t.Error("expected Contains(s, 3) to be true")
		}
		if Contains(s, 5) {
			t.Error("expected Contains(s, 5) to be false")
		}
	})

	t.Run("strings", func(t *testing.T) {
		s := []string{"apple", "banana", "cherry"}
		if !Contains(s, "banana") {
			t.Error("expected Contains(s, 'banana') to be true")
		}
		if Contains(s, "grape") {
			t.Error("expected Contains(s, 'grape') to be false")
		}
	})
}

// ==================================================
// STRING TESTS (11 - 20)
// ==================================================

func TestReverseString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"Go ♥ UTF-8", "8-FTU ♥ oG"}, // check multi-byte rune compatibility
	}

	for _, tt := range tests {
		got := ReverseString(tt.input)
		if got != tt.want {
			t.Errorf("ReverseString(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"racecar", true},
		{"A man, a plan, a canal: Panama", true},
		{"hello", false},
		{"", true},
	}

	for _, tt := range tests {
		got := IsPalindrome(tt.input)
		if got != tt.want {
			t.Errorf("IsPalindrome(%q) = %t; want %t", tt.input, got, tt.want)
		}
	}
}

func TestCharFrequency(t *testing.T) {
	got := CharFrequency("hello")
	want := map[rune]int{'h': 1, 'e': 1, 'l': 2, 'o': 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CharFrequency(\"hello\") = %v; want %v", got, want)
	}
}

func TestCountWordFrequency(t *testing.T) {
	got := CountWordFrequency("Go is great! Go is fast.")
	want := map[string]int{"go": 2, "is": 2, "great": 1, "fast": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CountWordFrequency(...) = %v; want %v", got, want)
	}
}

func TestFindFirstNonRepeatingChar(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := FindFirstNonRepeatingChar("swiss")
		if err != nil || got != 'w' {
			t.Errorf("FindFirstNonRepeatingChar(\"swiss\") = (%q, %v); want ('w', nil)", got, err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		_, err := FindFirstNonRepeatingChar("abcabc")
		if err == nil {
			t.Error("expected error for 'abcabc'")
		}
	})
}

func TestAreAnagrams(t *testing.T) {
	tests := []struct {
		s1   string
		s2   string
		want bool
	}{
		{"listen", "silent", true},
		{"Astronomer", "Moon starer", true}, // spaces/case ignored
		{"hello", "world", false},
	}

	for _, tt := range tests {
		got := AreAnagrams(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("AreAnagrams(%q, %q) = %t; want %t", tt.s1, tt.s2, got, tt.want)
		}
	}
}

func TestRemoveDuplicateChars(t *testing.T) {
	got := RemoveDuplicateChars("google")
	want := "gole"
	if got != want {
		t.Errorf("RemoveDuplicateChars(\"google\") = %q; want %q", got, want)
	}
}

func TestCompressString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"aaabbc", "a3b2c1"},
		{"abcd", "a1b1c1d1"},
		{"", ""},
	}

	for _, tt := range tests {
		got := CompressString(tt.input)
		if got != tt.want {
			t.Errorf("CompressString(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"flower", "flow", "flight"}, "fl"},
		{[]string{"dog", "racecar", "car"}, ""},
		{[]string{"interview", "interval", "intermediate"}, "inter"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		got := LongestCommonPrefix(tt.input)
		if got != tt.want {
			t.Errorf("LongestCommonPrefix(%v) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestToTitleCase(t *testing.T) {
	got := ToTitleCase("hello WORLD, this IS go!")
	want := "Hello World, This Is Go!"
	if got != want {
		t.Errorf("ToTitleCase(...) = %q; want %q", got, want)
	}
}

// ==================================================
// MAP TESTS (21 - 25)
// ==================================================

func TestBuildWordFrequencyCounter(t *testing.T) {
	got := BuildWordFrequencyCounter("apple banana apple")
	want := map[string]int{"apple": 2, "banana": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildWordFrequencyCounter(...) = %v; want %v", got, want)
	}
}

func TestFindDuplicates(t *testing.T) {
	got := FindDuplicates([]int{1, 2, 3, 2, 4, 5, 1, 1})
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindDuplicates(...) = %v; want %v", got, want)
	}
}

func TestSet(t *testing.T) {
	set := NewSet[string]()
	set.Add("a")
	set.Add("b")
	set.Add("a") // duplicate

	if set.Size() != 2 {
		t.Errorf("Size() = %d; want 2", set.Size())
	}
	if !set.Contains("a") {
		t.Error("expected set to contain 'a'")
	}
	if set.Contains("c") {
		t.Error("did not expect set to contain 'c'")
	}

	set.Remove("a")
	if set.Contains("a") {
		t.Error("expected set to not contain 'a' after removal")
	}
}

func TestMergeMaps(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}
	got := MergeMaps(m1, m2)
	want := map[string]int{"a": 1, "b": 3, "c": 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MergeMaps = %v; want %v", got, want)
	}
}

func TestInvertMap(t *testing.T) {
	m := map[string]int{"apple": 1, "banana": 2, "cherry": 1}
	got := InvertMap(m)

	// Since slice ordering is non-deterministic, we sort them before asserting
	for k := range got {
		sort.Strings(got[k])
	}
	want := map[int][]string{
		1: {"apple", "cherry"},
		2: {"banana"},
	}
	for k := range want {
		sort.Strings(want[k])
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("InvertMap = %v; want %v", got, want)
	}
}

// ==================================================
// STRUCTS & METHODS TESTS (26 - 30)
// ==================================================

func TestUserMethods(t *testing.T) {
	u := User{ID: 1, Username: "bob", Email: "bob@example.com", IsActive: false}

	u.Activate()
	if !u.IsActive {
		t.Error("expected user to be active")
	}

	u.Deactivate()
	if u.IsActive {
		t.Error("expected user to be inactive")
	}

	u.UpdateEmail("bob_new@example.com")
	if u.Email != "bob_new@example.com" {
		t.Errorf("UpdateEmail failed; got %s", u.Email)
	}

	expectedStr := "User #1: bob <bob_new@example.com> [Inactive]"
	if u.String() != expectedStr {
		t.Errorf("String() = %q; want %q", u.String(), expectedStr)
	}
}

func TestRectangle(t *testing.T) {
	r := Rectangle{Width: 4, Height: 5}
	if r.Area() != 20 {
		t.Errorf("Area() = %f; want 20", r.Area())
	}
	if r.Perimeter() != 18 {
		t.Errorf("Perimeter() = %f; want 18", r.Perimeter())
	}
}

func TestPointerVsValueReceiver(t *testing.T) {
	r := Rectangle{Width: 2, Height: 3}

	// Value receiver scale
	scaledVal := r.ScaleVal(2.0)
	if r.Width != 2 || r.Height != 3 {
		t.Errorf("value receiver modified original: %v", r)
	}
	if scaledVal.Width != 4 || scaledVal.Height != 6 {
		t.Errorf("value receiver did not return scaled copy: %v", scaledVal)
	}

	// Pointer receiver scale
	r.ScalePtr(2.0)
	if r.Width != 4 || r.Height != 6 {
		t.Errorf("pointer receiver failed to modify original in-place: %v", r)
	}
}

func TestAddressBook(t *testing.T) {
	ab := NewAddressBook()
	c1 := Contact{
		Name:  "Alice Smith",
		Phone: "123-456",
		Address: Address{
			Street:  "123 Main St",
			City:    "Springfield",
			State:   "IL",
			ZipCode: "62701",
		},
	}
	c2 := Contact{
		Name:  "Bob Jones",
		Phone: "789-012",
		Address: Address{
			Street:  "456 Oak Rd",
			City:    "Chicago",
			State:   "IL",
			ZipCode: "60601",
		},
	}

	ab.AddContact(c1)
	ab.AddContact(c2)

	if len(ab.ListContacts()) != 2 {
		t.Errorf("expected 2 contacts, got %d", len(ab.ListContacts()))
	}

	results, err := ab.SearchByName("alice")
	if err != nil || len(results) != 1 || results[0].Name != "Alice Smith" {
		t.Errorf("SearchByName('alice') failed: got %v, err: %v", results, err)
	}

	_, err = ab.SearchByName("charlie")
	if err == nil {
		t.Error("expected error for non-existent contact search")
	}
}

// ==================================================
// FUNCTIONS & POINTERS TESTS (31 - 35)
// ==================================================

func TestVariadicSum(t *testing.T) {
	if got := VariadicSum(); got != 0 {
		t.Errorf("VariadicSum() = %d; want 0", got)
	}
	if got := VariadicSum(1, 2, 3); got != 6 {
		t.Errorf("VariadicSum(1,2,3) = %d; want 6", got)
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{5, 120},
		{-3, 0},
	}
	for _, tt := range tests {
		if got := Factorial(tt.n); got != tt.want {
			t.Errorf("Factorial(%d) = %d; want %d", tt.n, got, tt.want)
		}
	}
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{6, 8},
		{-1, 0},
	}
	for _, tt := range tests {
		if got := Fibonacci(tt.n); got != tt.want {
			t.Errorf("Fibonacci(%d) = %d; want %d", tt.n, got, tt.want)
		}
	}
}

func TestSwap(t *testing.T) {
	x, y := 10, 20
	Swap(&x, &y)
	if x != 20 || y != 10 {
		t.Errorf("Swap failed; got x=%d, y=%d", x, y)
	}
}

func TestNewCounter(t *testing.T) {
	counter := NewCounter()
	if c := counter(); c != 1 {
		t.Errorf("counter() first call = %d; want 1", c)
	}
	if c := counter(); c != 2 {
		t.Errorf("counter() second call = %d; want 2", c)
	}
}

// ==================================================
// INTERFACES TESTS (36 - 40)
// ==================================================

func TestShapeInterface(t *testing.T) {
	var s Shape

	s = Circle{Radius: 2.0}
	if math.Abs(s.Area()-math.Pi*4.0) > 1e-9 {
		t.Errorf("Circle Area = %f; want %f", s.Area(), math.Pi*4.0)
	}

	s = Rectangle{Width: 3.0, Height: 4.0}
	if s.Area() != 12.0 {
		t.Errorf("Rectangle Area = %f; want 12.0", s.Area())
	}
}

func TestPaymentProcessor(t *testing.T) {
	var processor PaymentProcessor

	processor = CreditCardProcessor{CardNumber: "1234567890123456", CardHolder: "John Doe"}
	msg, err := processor.ProcessPayment(100.50)
	if err != nil || msg == "" {
		t.Errorf("CreditCardProcessor failed: msg=%q, err=%v", msg, err)
	}

	processor = PayPalProcessor{Email: "buyer@example.com"}
	msg, err = processor.ProcessPayment(50.25)
	if err != nil || msg == "" {
		t.Errorf("PayPalProcessor failed: msg=%q, err=%v", msg, err)
	}
}

func TestLogger(t *testing.T) {
	bl := &BufferLogger{}
	var logger Logger = bl

	logger.Log("Hello Logger 1")
	logger.Log("Hello Logger 2")

	if len(bl.Logs) != 2 || bl.Logs[0] != "Hello Logger 1" {
		t.Errorf("Logger failed; got logs: %v", bl.Logs)
	}
}

func TestIdentifyShape(t *testing.T) {
	c := Circle{Radius: 5.5}
	r := Rectangle{Width: 10, Height: 20}

	if got := IdentifyShape(c); !reflect.DeepEqual(got, "Concrete Type: Circle (Radius: 5.50)") {
		t.Errorf("IdentifyShape(Circle) = %q", got)
	}
	if got := IdentifyShape(r); !reflect.DeepEqual(got, "Concrete Type: Rectangle (Width: 10.00, Height: 20.00)") {
		t.Errorf("IdentifyShape(Rectangle) = %q", got)
	}
}

func TestHandleType(t *testing.T) {
	if got := HandleType(42); got != "Type: int, Value: 42" {
		t.Errorf("HandleType(42) = %q", got)
	}
	if got := HandleType("hello"); got != "Type: string, Value: \"hello\"" {
		t.Errorf("HandleType('hello') = %q", got)
	}
	if got := HandleType(true); got != "Type: bool, Value: true" {
		t.Errorf("HandleType(true) = %q", got)
	}
	c := Circle{Radius: 2.0}
	if got := HandleType(c); got != "Type: Circle, Value: Circle{Radius: 2.00}" {
		t.Errorf("HandleType(Circle) = %q", got)
	}
}
