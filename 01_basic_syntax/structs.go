package main

import (
	"errors"
	"fmt"
	"strings"
)

// ==================================================
// STRUCTS & METHODS (Questions 26 - 30)
// ==================================================

// 26. User represents a system user.
type User struct {
	ID       int
	Username string
	Email    string
	IsActive bool
}

// Activate activates the user.
func (u *User) Activate() {
	u.IsActive = true
}

// Deactivate deactivates the user.
func (u *User) Deactivate() {
	u.IsActive = false
}

// UpdateEmail updates the user's email address.
func (u *User) UpdateEmail(newEmail string) {
	u.Email = newEmail
}

// 27. Implement String() method (fmt.Stringer interface) for User.
func (u User) String() string {
	status := "Inactive"
	if u.IsActive {
		status = "Active"
	}
	return fmt.Sprintf("User #%d: %s <%s> [%s]", u.ID, u.Username, u.Email, status)
}

// 28. Rectangle represents a two-dimensional rectangle.
type Rectangle struct {
	Width  float64
	Height float64
}

// Area calculates the area of the rectangle.
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter calculates the perimeter of the rectangle.
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// 29. Demonstrate pointer receiver vs value receiver.

// ScaleVal (Value Receiver) returns a new Rectangle scaled by factor, leaving original unmodified.
func (r Rectangle) ScaleVal(factor float64) Rectangle {
	r.Width *= factor
	r.Height *= factor
	return r
}

// ScalePtr (Pointer Receiver) scales the Rectangle in-place.
func (r *Rectangle) ScalePtr(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// 30. Nested structs for Address Book.

// Address represents a physical address.
type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
}

// Contact represents a contact card.
type Contact struct {
	Name    string
	Phone   string
	Address Address // nested struct
}

// AddressBook holds a list of contacts.
type AddressBook struct {
	Contacts []Contact
}

// NewAddressBook creates a new AddressBook.
func NewAddressBook() *AddressBook {
	return &AddressBook{
		Contacts: make([]Contact, 0),
	}
}

// AddContact adds a contact to the address book.
func (ab *AddressBook) AddContact(c Contact) {
	ab.Contacts = append(ab.Contacts, c)
}

// SearchByName searches contacts by name (case-insensitive substring match).
func (ab *AddressBook) SearchByName(name string) ([]Contact, error) {
	results := make([]Contact, 0)
	query := strings.ToLower(name)
	for _, c := range ab.Contacts {
		if strings.Contains(strings.ToLower(c.Name), query) {
			results = append(results, c)
		}
	}
	if len(results) == 0 {
		return nil, errors.New("no contacts found matching name")
	}
	return results, nil
}

// ListContacts returns all contacts in the address book.
func (ab *AddressBook) ListContacts() []Contact {
	return ab.Contacts
}
