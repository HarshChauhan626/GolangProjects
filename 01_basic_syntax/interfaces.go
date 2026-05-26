package main

import (
	"fmt"
	"math"
)

// ==================================================
// INTERFACES (Questions 36 - 40)
// ==================================================

// 36. Shape interface defines the contract for two-dimensional shapes.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Circle implements Shape.
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// 37. PaymentProcessor interface defines the contract for processing payments.
type PaymentProcessor interface {
	ProcessPayment(amount float64) (string, error)
}

// CreditCardProcessor implements PaymentProcessor.
type CreditCardProcessor struct {
	CardNumber string
	CardHolder string
}

func (cc CreditCardProcessor) ProcessPayment(amount float64) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("invalid credit card payment amount: %.2f", amount)
	}
	// Mask card number for display
	maskedCard := "XXXX-XXXX-XXXX-" + cc.CardNumber[len(cc.CardNumber)-4:]
	return fmt.Sprintf("Processed Credit Card payment of $%.2f for %s (%s)", amount, cc.CardHolder, maskedCard), nil
}

// PayPalProcessor implements PaymentProcessor.
type PayPalProcessor struct {
	Email string
}

func (pp PayPalProcessor) ProcessPayment(amount float64) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("invalid paypal payment amount: %.2f", amount)
	}
	return fmt.Sprintf("Processed PayPal payment of $%.2f from account %s", amount, pp.Email), nil
}

// 38. Logger interface defines a simple logging contract.
type Logger interface {
	Log(message string)
}

// ConsoleLogger implements Logger by printing to standard output.
type ConsoleLogger struct {
	Prefix string
}

func (cl ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", cl.Prefix, message)
}

// BufferLogger implements Logger by storing log messages in memory for verification.
type BufferLogger struct {
	Logs []string
}

func (bl *BufferLogger) Log(message string) {
	bl.Logs = append(bl.Logs, message)
}

// 39. IdentifyShape uses type assertions to identify the concrete type of a Shape.
func IdentifyShape(s Shape) string {
	// Try assertion to Circle
	if c, ok := s.(Circle); ok {
		return fmt.Sprintf("Concrete Type: Circle (Radius: %.2f)", c.Radius)
	}

	// Try assertion to Rectangle
	if r, ok := s.(Rectangle); ok {
		return fmt.Sprintf("Concrete Type: Rectangle (Width: %.2f, Height: %.2f)", r.Width, r.Height)
	}

	return "Concrete Type: Unknown Shape"
}

// 40. HandleType uses a type switch to format details depending on the input type.
func HandleType(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("Type: int, Value: %d", v)
	case string:
		return fmt.Sprintf("Type: string, Value: %q", v)
	case bool:
		return fmt.Sprintf("Type: bool, Value: %t", v)
	case Circle:
		return fmt.Sprintf("Type: Circle, Value: Circle{Radius: %.2f}", v.Radius)
	case Rectangle:
		return fmt.Sprintf("Type: Rectangle, Value: Rectangle{Width: %.2f, Height: %.2f}", v.Width, v.Height)
	default:
		return fmt.Sprintf("Type: Unknown (%T), Value: %v", v, v)
	}
}
