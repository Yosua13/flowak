package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"backend/config"
	"backend/db"
	"backend/middleware"
	"backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Email regex validation
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// GenerateUUID formats a secure UUIDv4 string
func GenerateUUID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp + random string if entropy fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

// RegisterHandler registers a new user
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request body"}`))
		return
	}

	// 1. Validation & Sanitization
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password
	role := strings.TrimSpace(req.Role)

	if name == "" || email == "" || password == "" || role == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "All fields are required"}`))
		return
	}

	if len(name) < 2 || len(name) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Name must be between 2 and 100 characters"}`))
		return
	}

	if !emailRegex.MatchString(email) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid email address format"}`))
		return
	}

	if len(password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Password must be at least 6 characters"}`))
		return
	}

	// Allowed roles validation
	if role != "pm" && role != "uiux" && role != "frontend" && role != "backend" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid role. Must be one of: pm, uiux, frontend, backend"}`))
		return
	}

	// 2. Check if email already exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database error checking user"}`))
		return
	}
	if exists {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Email address already registered"}`))
		return
	}

	// 3. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to encrypt password"}`))
		return
	}

	// 4. Save to Database
	userID := "usr_" + GenerateUUID()
	_, err = db.DB.Exec("INSERT INTO users (id, name, email, password_hash, role) VALUES (?, ?, ?, ?, ?)",
		userID, name, email, string(hashedPassword), role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to save user to database"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"success": true, "message": "User registered successfully", "user_id": "%s"}`, userID)))
}

// LoginHandler authenticates user and returns JWT
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid request body"}`))
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if email == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Email and password are required"}`))
		return
	}

	// 1. Find user in database
	var user models.User
	var passwordHash string
	err := db.DB.QueryRow("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Name, &user.Email, &passwordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid email or password"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database query error"}`))
		return
	}

	// 2. Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid email or password"}`))
		return
	}

	// 3. Generate JWT Token (Expired in 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.ActiveConfig.JWTSecret))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to sign authentication token"}`))
		return
	}

	resp := models.UserLoginResponse{
		Token: tokenString,
		User:  user,
	}

	json.NewEncoder(w).Encode(resp)
}
