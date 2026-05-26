package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ==================================================
// TUTORIAL 35: URL SHORTENER
// ==================================================
//
// PROBLEM STATEMENT:
// Build a mini URL shortening service (like bit.ly) that:
//   - Accepts long URLs and returns short codes
//   - Redirects short codes to original URLs
//   - Tracks click statistics
//   - Validates URLs before shortening
//
// This exercise combines API design, storage, business logic, and
// HTTP redirects into a cohesive system.
//
// ARCHITECTURE:
//
//   POST /shorten {"url": "https://..."} → {"short_url": "http://localhost/abc123"}
//   GET  /abc123                          → 302 Redirect to original URL
//   GET  /stats/abc123                    → {"clicks": 42, "created_at": "..."}
//

// URLEntry represents a shortened URL record.
type URLEntry struct {
	ShortCode  string    `json:"short_code"`
	OriginalURL string   `json:"original_url"`
	ShortURL   string    `json:"short_url"`
	CreatedAt  time.Time `json:"created_at"`
	Clicks     int64     `json:"clicks"`
	LastAccess time.Time `json:"last_access,omitempty"`
}

// URLStore manages the mapping between short codes and URLs.
type URLStore struct {
	mu       sync.RWMutex
	urls     map[string]*URLEntry // shortCode → entry
	baseURL  string
}

// NewURLStore creates a new URL store.
func NewURLStore(baseURL string) *URLStore {
	return &URLStore{
		urls:    make(map[string]*URLEntry),
		baseURL: baseURL,
	}
}

// generateShortCode creates a random 6-character hex string.
func generateShortCode() string {
	bytes := make([]byte, 3) // 3 bytes = 6 hex characters
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Shorten creates a short URL for the given long URL.
func (s *URLStore) Shorten(originalURL string) (*URLEntry, error) {
	// Validate URL
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	// Check if URL already exists
	s.mu.RLock()
	for _, entry := range s.urls {
		if entry.OriginalURL == originalURL {
			s.mu.RUnlock()
			return entry, nil
		}
	}
	s.mu.RUnlock()

	// Generate a unique short code
	var shortCode string
	for {
		shortCode = generateShortCode()
		s.mu.RLock()
		_, exists := s.urls[shortCode]
		s.mu.RUnlock()
		if !exists {
			break
		}
	}

	entry := &URLEntry{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.urls[shortCode] = entry
	s.mu.Unlock()

	return entry, nil
}

// Resolve looks up a short code and increments the click counter.
func (s *URLStore) Resolve(shortCode string) (string, bool) {
	s.mu.RLock()
	entry, ok := s.urls[shortCode]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}

	// Atomically increment clicks
	atomic.AddInt64(&entry.Clicks, 1)

	s.mu.Lock()
	entry.LastAccess = time.Now()
	s.mu.Unlock()

	return entry.OriginalURL, true
}

// GetStats returns the stats for a short code.
func (s *URLStore) GetStats(shortCode string) (*URLEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.urls[shortCode]
	if !ok {
		return nil, false
	}

	// Return a copy
	copy := *entry
	copy.Clicks = atomic.LoadInt64(&entry.Clicks)
	return &copy, true
}

// GetAll returns all URL entries.
func (s *URLStore) GetAll() []*URLEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*URLEntry, 0, len(s.urls))
	for _, entry := range s.urls {
		copy := *entry
		copy.Clicks = atomic.LoadInt64(&entry.Clicks)
		result = append(result, &copy)
	}
	return result
}

// --- Validation ---

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL is required")
	}
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}
	if len(rawURL) > 2048 {
		return fmt.Errorf("URL is too long (max 2048 characters)")
	}
	// Check for a valid-looking domain
	withoutScheme := strings.TrimPrefix(strings.TrimPrefix(rawURL, "https://"), "http://")
	if !strings.Contains(withoutScheme, ".") {
		return fmt.Errorf("URL must contain a valid domain")
	}
	return nil
}

// --- HTTP Handlers ---

type URLShortenerApp struct {
	store *URLStore
}

// handleShorten creates a short URL.
func (app *URLShortenerApp) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	entry, err := app.store.Shorten(req.URL)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Printf("[Shorten] %s → %s\n", entry.OriginalURL, entry.ShortURL)
	respondWithJSON(w, http.StatusCreated, entry)
}

// handleRedirect resolves a short code and redirects.
func (app *URLShortenerApp) handleRedirect(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" || shortCode == "favicon.ico" {
		return
	}

	// Skip known routes
	if shortCode == "shorten" || strings.HasPrefix(shortCode, "stats/") || shortCode == "all" {
		return
	}

	originalURL, ok := app.store.Resolve(shortCode)
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("short code '%s' not found", shortCode))
		return
	}

	fmt.Printf("[Redirect] %s → %s\n", shortCode, originalURL)
	http.Redirect(w, r, originalURL, http.StatusFound) // 302 redirect
}

// handleStats returns click statistics for a short code.
func (app *URLShortenerApp) handleStats(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/stats/")
	if shortCode == "" {
		respondWithError(w, http.StatusBadRequest, "short code required")
		return
	}

	entry, ok := app.store.GetStats(shortCode)
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("short code '%s' not found", shortCode))
		return
	}

	respondWithJSON(w, http.StatusOK, entry)
}

// handleListAll returns all shortened URLs.
func (app *URLShortenerApp) handleListAll(w http.ResponseWriter, r *http.Request) {
	entries := app.store.GetAll()
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(entries),
		"urls":  entries,
	})
}

// --- HTTP Helpers ---

func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, status int, message string) {
	respondWithJSON(w, status, map[string]string{"error": message})
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("         TUTORIAL 35: URL SHORTENER                ")
	fmt.Println("==================================================")

	store := NewURLStore("http://localhost:8080")
	app := &URLShortenerApp{store: store}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/shorten", app.handleShorten)   // POST — create short URL
	mux.HandleFunc("/stats/", app.handleStats)       // GET  — view stats
	mux.HandleFunc("/all", app.handleListAll)         // GET  — list all URLs

	// Redirect route (catch-all for short codes)
	mux.HandleFunc("/", app.handleRedirect) // GET /{code} → redirect

	fmt.Println("\nServer starting on :8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  POST /shorten         — Shorten a URL")
	fmt.Println("  GET  /{code}          — Redirect to original URL")
	fmt.Println("  GET  /stats/{code}    — View click statistics")
	fmt.Println("  GET  /all             — List all shortened URLs")
	fmt.Println("\nTest commands:")
	fmt.Println(`  curl -X POST -d '{"url":"https://golang.org"}' http://localhost:8080/shorten`)
	fmt.Println("  curl -L http://localhost:8080/{code}")
	fmt.Println("  curl http://localhost:8080/stats/{code}")
	fmt.Println("  curl http://localhost:8080/all")

	log.Fatal(http.ListenAndServe(":8080", mux))
}
