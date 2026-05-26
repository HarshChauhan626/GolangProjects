package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ==================================================
// TUTORIAL 32: CRUD REST API
// ==================================================
//
// PROBLEM STATEMENT:
// Implement a full RESTful API for a "Book" resource with:
//   - CREATE:  POST   /books      — add a new book
//   - READ:    GET    /books      — list all books
//   - READ:    GET    /books/{id} — get a single book
//   - UPDATE:  PUT    /books/{id} — update a book
//   - DELETE:  DELETE /books/{id} — delete a book
//
// Includes input validation, proper HTTP status codes, and JSON responses.
//
// ARCHITECTURE:
//
//   Client → HTTP Request → Router → Handler → Store → JSON Response
//

// Book represents the domain model.
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	ISBN      string    `json:"isbn"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BookInput is the expected request body for creating/updating a book.
type BookInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Year   int    `json:"year"`
}

// Validate checks the input for required fields and constraints.
func (b BookInput) Validate() []string {
	var errors []string
	if strings.TrimSpace(b.Title) == "" {
		errors = append(errors, "title is required")
	}
	if strings.TrimSpace(b.Author) == "" {
		errors = append(errors, "author is required")
	}
	if b.Year != 0 && (b.Year < 1000 || b.Year > time.Now().Year()+1) {
		errors = append(errors, fmt.Sprintf("year must be between 1000 and %d", time.Now().Year()+1))
	}
	return errors
}

// --- In-Memory Store ---

// BookStore is a thread-safe in-memory storage for books.
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

// NewBookStore creates a store with some seed data.
func NewBookStore() *BookStore {
	store := &BookStore{
		books:  make(map[int]Book),
		nextID: 1,
	}

	// Seed data
	now := time.Now()
	seedBooks := []BookInput{
		{"The Go Programming Language", "Alan Donovan", "978-0134190440", 2015},
		{"Concurrency in Go", "Katherine Cox-Buday", "978-1491941195", 2017},
		{"Learning Go", "Jon Bodner", "978-1492077213", 2021},
	}
	for _, b := range seedBooks {
		store.books[store.nextID] = Book{
			ID: store.nextID, Title: b.Title, Author: b.Author,
			ISBN: b.ISBN, Year: b.Year, CreatedAt: now, UpdatedAt: now,
		}
		store.nextID++
	}

	return store
}

func (s *BookStore) GetAll() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		result = append(result, b)
	}
	return result
}

func (s *BookStore) GetByID(id int) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.books[id]
	return b, ok
}

func (s *BookStore) Create(input BookInput) Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	book := Book{
		ID: s.nextID, Title: input.Title, Author: input.Author,
		ISBN: input.ISBN, Year: input.Year, CreatedAt: now, UpdatedAt: now,
	}
	s.books[s.nextID] = book
	s.nextID++
	return book
}

func (s *BookStore) Update(id int, input BookInput) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, ok := s.books[id]
	if !ok {
		return Book{}, false
	}
	book.Title = input.Title
	book.Author = input.Author
	book.ISBN = input.ISBN
	book.Year = input.Year
	book.UpdatedAt = time.Now()
	s.books[id] = book
	return book, true
}

func (s *BookStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.books[id]; !ok {
		return false
	}
	delete(s.books, id)
	return true
}

// --- HTTP Helpers ---

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// extractID parses the ID from URL path like "/books/123".
func extractID(path, prefix string) (int, error) {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, "/")
	return strconv.Atoi(idStr)
}

// --- Handlers ---

// BookHandler handles all /books routes using method-based routing.
type BookHandler struct {
	store *BookStore
}

func (h *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route: /books (collection) vs /books/{id} (single resource)
	if r.URL.Path == "/books" || r.URL.Path == "/books/" {
		switch r.Method {
		case http.MethodGet:
			h.listBooks(w, r)
		case http.MethodPost:
			h.createBook(w, r)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// Route: /books/{id}
	if strings.HasPrefix(r.URL.Path, "/books/") {
		id, err := extractID(r.URL.Path, "/books/")
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid book ID")
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getBook(w, r, id)
		case http.MethodPut:
			h.updateBook(w, r, id)
		case http.MethodDelete:
			h.deleteBook(w, r, id)
		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	respondError(w, http.StatusNotFound, "not found")
}

func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	books := h.store.GetAll()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(books),
		"books": books,
	})
}

func (h *BookHandler) getBook(w http.ResponseWriter, r *http.Request, id int) {
	book, ok := h.store.GetByID(id)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("book %d not found", id))
		return
	}
	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var input BookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Validate
	if errs := input.Validate(); len(errs) > 0 {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":   "validation failed",
			"details": errs,
		})
		return
	}

	book := h.store.Create(input)
	respondJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) updateBook(w http.ResponseWriter, r *http.Request, id int) {
	var input BookInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":   "validation failed",
			"details": errs,
		})
		return
	}

	book, ok := h.store.Update(id, input)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("book %d not found", id))
		return
	}
	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	if !h.store.Delete(id) {
		respondError(w, http.StatusNotFound, fmt.Sprintf("book %d not found", id))
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204 — deleted successfully, no body
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 32: CRUD REST API                ")
	fmt.Println("==================================================")

	store := NewBookStore()
	handler := &BookHandler{store: store}

	mux := http.NewServeMux()
	mux.Handle("/books", handler)
	mux.Handle("/books/", handler)

	fmt.Println("\nServer starting on :8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET    /books       — List all books")
	fmt.Println("  POST   /books       — Create a book")
	fmt.Println("  GET    /books/{id}  — Get a book by ID")
	fmt.Println("  PUT    /books/{id}  — Update a book")
	fmt.Println("  DELETE /books/{id}  — Delete a book")
	fmt.Println("\nTest commands:")
	fmt.Println("  curl http://localhost:8080/books")
	fmt.Println("  curl http://localhost:8080/books/1")
	fmt.Println(`  curl -X POST -d '{"title":"New Book","author":"Author","year":2024}' http://localhost:8080/books`)
	fmt.Println(`  curl -X PUT -d '{"title":"Updated","author":"Author","year":2024}' http://localhost:8080/books/1`)
	fmt.Println("  curl -X DELETE http://localhost:8080/books/1")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
