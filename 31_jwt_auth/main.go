package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ==================================================
// TUTORIAL 31: JWT AUTHENTICATION MIDDLEWARE
// ==================================================
//
// PROBLEM STATEMENT:
// JSON Web Tokens (JWTs) are a compact, URL-safe way to represent claims
// between two parties. This tutorial implements JWT generation and
// validation from scratch using HMAC-SHA256, without external libraries.
//
// JWT STRUCTURE:
//
//   Header.Payload.Signature
//
//   Header:    {"alg": "HS256", "typ": "JWT"}  (base64url encoded)
//   Payload:   {"sub": "user123", "exp": ...}   (base64url encoded)
//   Signature: HMAC-SHA256(header.payload, secret)
//
// FLOW:
//
//   Login → Server generates JWT → Client stores it
//   Request → Client sends JWT in Authorization header
//          → Middleware validates JWT → Allow/Deny
//

// --- JWT Implementation ---

var jwtSecret = []byte("super-secret-key-change-in-production")

// JWTHeader is the first part of a JWT.
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTPayload is the claims portion of a JWT.
type JWTPayload struct {
	Subject   string `json:"sub"`   // User ID
	Name      string `json:"name"`  // User's name
	Role      string `json:"role"`  // User's role
	IssuedAt  int64  `json:"iat"`   // Issued at (unix timestamp)
	ExpiresAt int64  `json:"exp"`   // Expires at (unix timestamp)
}

// base64URLEncode encodes bytes to base64url (no padding).
func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

// base64URLDecode decodes a base64url string.
func base64URLDecode(s string) ([]byte, error) {
	// Add padding back
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// GenerateJWT creates a new JWT token for the given user.
func GenerateJWT(userID, name, role string, ttl time.Duration) (string, error) {
	// 1. Create header
	header := JWTHeader{Algorithm: "HS256", Type: "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64URLEncode(headerJSON)

	// 2. Create payload with claims
	now := time.Now()
	payload := JWTPayload{
		Subject:   userID,
		Name:      name,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64URLEncode(payloadJSON)

	// 3. Create signature: HMAC-SHA256(header.payload, secret)
	signingInput := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	signature := base64URLEncode(mac.Sum(nil))

	// 4. Combine: header.payload.signature
	token := signingInput + "." + signature
	return token, nil
}

// ValidateJWT parses and validates a JWT token.
// Returns the payload if valid, or an error if invalid/expired.
func ValidateJWT(token string) (*JWTPayload, error) {
	// 1. Split into 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	headerEncoded, payloadEncoded, signatureEncoded := parts[0], parts[1], parts[2]

	// 2. Verify signature
	signingInput := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(signingInput))
	expectedSignature := base64URLEncode(mac.Sum(nil))

	if !hmac.Equal([]byte(signatureEncoded), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// 3. Decode and parse payload
	payloadJSON, err := base64URLDecode(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	var payload JWTPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload: %w", err)
	}

	// 4. Check expiration
	if time.Now().Unix() > payload.ExpiresAt {
		return nil, fmt.Errorf("token expired at %s",
			time.Unix(payload.ExpiresAt, 0).Format(time.RFC3339))
	}

	return &payload, nil
}

// --- HTTP Middleware ---

// JWTAuthMiddleware validates the JWT token in the Authorization header.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid Authorization format, use: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		// Validate the token
		payload, err := ValidateJWT(parts[1])
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Token is valid — add user info to request headers for downstream handlers
		r.Header.Set("X-User-ID", payload.Subject)
		r.Header.Set("X-User-Name", payload.Name)
		r.Header.Set("X-User-Role", payload.Role)

		fmt.Printf("[JWT Auth] ✅ User: %s (%s), Role: %s\n", payload.Name, payload.Subject, payload.Role)
		next.ServeHTTP(w, r)
	})
}

// RoleMiddleware restricts access to specific roles.
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")
			for _, role := range allowedRoles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error": "forbidden: insufficient role"}`, http.StatusForbidden)
		})
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("   TUTORIAL 31: JWT AUTHENTICATION MIDDLEWARE      ")
	fmt.Println("==================================================")

	mux := http.NewServeMux()

	// --- Public: Login endpoint (generates JWT) ---
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// In production, validate credentials against a database
		username := r.URL.Query().Get("user")
		role := r.URL.Query().Get("role")
		if username == "" {
			username = "alice"
		}
		if role == "" {
			role = "user"
		}

		token, err := GenerateJWT(username, username, role, 1*time.Hour)
		if err != nil {
			http.Error(w, `{"error": "failed to generate token"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token": "%s"}`, token)
		fmt.Printf("[Login] Generated token for user '%s' (role: %s)\n", username, role)
	})

	// --- Protected: User profile (any authenticated user) ---
	profileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user_id": "%s", "name": "%s", "role": "%s"}`,
			r.Header.Get("X-User-ID"),
			r.Header.Get("X-User-Name"),
			r.Header.Get("X-User-Role"))
	})
	mux.Handle("/profile", JWTAuthMiddleware(profileHandler))

	// --- Protected: Admin-only endpoint ---
	adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"message": "Welcome, admin! You have full access."}`)
	})
	mux.Handle("/admin", JWTAuthMiddleware(RoleMiddleware("admin")(adminHandler)))

	fmt.Println("\nServer starting on :8080")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET /login?user=alice&role=admin  → Get JWT token")
	fmt.Println("  GET /profile                      → Protected (any user)")
	fmt.Println("  GET /admin                        → Protected (admin only)")
	fmt.Println("\nTest flow:")
	fmt.Println("  1. curl http://localhost:8080/login?user=alice&role=admin")
	fmt.Println(`  2. curl -H "Authorization: Bearer <token>" http://localhost:8080/profile`)
	fmt.Println(`  3. curl -H "Authorization: Bearer <token>" http://localhost:8080/admin`)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
