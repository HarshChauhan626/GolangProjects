package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 28: REPOSITORY PATTERN
// ==================================================
//
// PROBLEM STATEMENT:
// The Repository Pattern abstracts data access behind a clean interface.
// Business logic interacts with the repository interface, not the concrete
// database implementation. This enables:
//   - Swapping databases without changing business logic
//   - Easy unit testing with mock repositories
//   - Clean separation of concerns
//
// ARCHITECTURE:
//
//   Business Logic (Service)
//         │
//         ▼
//   UserRepository (interface)
//         │
//         ├──▶ InMemoryUserRepository (for testing)
//         └──▶ SQLUserRepository (for production)
//

// --- Domain Model ---

// User is the domain entity.
type User struct {
	ID        int
	Name      string
	Email     string
	Active    bool
	CreatedAt time.Time
}

// --- Repository Interface ---

// UserRepository defines the contract for user data access.
// Any implementation must satisfy these methods.
type UserRepository interface {
	Create(user *User) error
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]*User, error)
	Update(user *User) error
	Delete(id int) error
}

// Common errors that repositories can return.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUser       = errors.New("invalid user data")
)

// --- In-Memory Implementation (for testing and demonstration) ---

// InMemoryUserRepository stores users in a map (simulates a database).
type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewInMemoryUserRepository creates an empty in-memory repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) Create(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate email
	for _, u := range r.users {
		if strings.EqualFold(u.Email, user.Email) {
			return ErrUserAlreadyExists
		}
	}

	user.ID = r.nextID
	user.CreatedAt = time.Now()
	r.nextID++

	// Store a copy to prevent external modification
	stored := *user
	r.users[user.ID] = &stored
	return nil
}

func (r *InMemoryUserRepository) GetByID(id int) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	// Return a copy
	result := *user
	return &result, nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Email, email) {
			result := *user
			return &result, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *InMemoryUserRepository) GetAll() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		copy := *user
		result = append(result, &copy)
	}
	return result, nil
}

func (r *InMemoryUserRepository) Update(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return ErrUserNotFound
	}

	stored := *user
	r.users[user.ID] = &stored
	return nil
}

func (r *InMemoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

// --- Service Layer (uses the interface, not the implementation) ---

// UserService contains business logic and depends on the repository interface.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a service with the given repository.
// The service doesn't know or care which implementation it gets.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// RegisterUser validates and creates a new user.
func (s *UserService) RegisterUser(name, email string) (*User, error) {
	// Business logic: validation
	if name == "" || email == "" {
		return nil, ErrInvalidUser
	}
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: invalid email format", ErrInvalidUser)
	}

	user := &User{
		Name:   name,
		Email:  email,
		Active: true,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// DeactivateUser sets a user's Active flag to false.
func (s *UserService) DeactivateUser(id int) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	user.Active = false
	return s.repo.Update(user)
}

// GetActiveUsers returns only active users.
func (s *UserService) GetActiveUsers() ([]*User, error) {
	all, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	active := make([]*User, 0)
	for _, u := range all {
		if u.Active {
			active = append(active, u)
		}
	}
	return active, nil
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("     TUTORIAL 28: REPOSITORY PATTERN               ")
	fmt.Println("==================================================")

	// Create an in-memory repository (swap with SQL in production)
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// --- Demo 1: Register users ---
	fmt.Println("\n--- Demo 1: Register Users ---")
	users := []struct{ name, email string }{
		{"Alice", "alice@example.com"},
		{"Bob", "bob@example.com"},
		{"Charlie", "charlie@example.com"},
		{"Diana", "diana@example.com"},
	}

	for _, u := range users {
		user, err := service.RegisterUser(u.name, u.email)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("  ✅ Registered: ID=%d, Name=%s, Email=%s\n",
				user.ID, user.Name, user.Email)
		}
	}

	// --- Demo 2: Duplicate email ---
	fmt.Println("\n--- Demo 2: Duplicate Email ---")
	_, err := service.RegisterUser("Alice2", "alice@example.com")
	fmt.Printf("  Duplicate registration: %v\n", err)

	// --- Demo 3: Invalid user ---
	fmt.Println("\n--- Demo 3: Validation ---")
	_, err = service.RegisterUser("", "test@test.com")
	fmt.Printf("  Empty name: %v\n", err)
	_, err = service.RegisterUser("Test", "invalid-email")
	fmt.Printf("  Bad email: %v\n", err)

	// --- Demo 4: Get by ID ---
	fmt.Println("\n--- Demo 4: Get By ID ---")
	user, err := repo.GetByID(2)
	if err == nil {
		fmt.Printf("  User 2: %s (%s)\n", user.Name, user.Email)
	}

	_, err = repo.GetByID(999)
	fmt.Printf("  User 999: %v\n", err)

	// --- Demo 5: Deactivate and list active ---
	fmt.Println("\n--- Demo 5: Deactivate User ---")
	_ = service.DeactivateUser(2)
	fmt.Println("  Deactivated user 2 (Bob).")

	active, _ := service.GetActiveUsers()
	fmt.Printf("  Active users (%d):\n", len(active))
	for _, u := range active {
		fmt.Printf("    - %s (%s)\n", u.Name, u.Email)
	}

	// --- Demo 6: Delete ---
	fmt.Println("\n--- Demo 6: Delete User ---")
	_ = repo.Delete(3)
	fmt.Println("  Deleted user 3 (Charlie).")
	all, _ := repo.GetAll()
	fmt.Printf("  Remaining users: %d\n", len(all))

	fmt.Println("\nRepository pattern demo complete!")
	fmt.Println("Tutorial 28 complete!")
}
