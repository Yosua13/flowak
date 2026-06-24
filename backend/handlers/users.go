package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"golang.org/x/crypto/bcrypt"
)

// UsersHandler handles GET /api/users and POST /api/users
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Fetch all users in the system to list as team contributors
		rows, err := db.DB.Query("SELECT id, name, email, role FROM users ORDER BY name ASC")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Database query error"}`))
			return
		}
		defer rows.Close()

		var usersList []models.User
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to parse users"}`))
				return
			}
			usersList = append(usersList, u)
		}

		json.NewEncoder(w).Encode(usersList)

	case http.MethodPost:
		// Add a new contributor
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid request body"}`))
			return
		}

		name := strings.TrimSpace(req.Name)
		email := strings.ToLower(strings.TrimSpace(req.Email))
		role := strings.TrimSpace(req.Role)

		if name == "" || email == "" || role == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "All fields are required"}`))
			return
		}

		if !emailRegex.MatchString(email) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid email address format"}`))
			return
		}

		if role != "pm" && role != "uiux" && role != "frontend" && role != "backend" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid role"}`))
			return
		}

		// Check duplicate
		var exists bool
		err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to check email availability"}`))
			return
		}
		if exists {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "Email is already registered"}`))
			return
		}

		// Create a random password since this is added by team, default password is 'password123'
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Encryption error"}`))
			return
		}

		newUserID := "usr_" + GenerateUUID()
		_, err = db.DB.Exec("INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)",
			newUserID, name, email, string(hashedPassword), role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to save user"}`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"success": true, "user_id": "%s"}`, newUserID)))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}

// UserDetailHandler handles DELETE /api/users/{id}
func UserDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	currentUserID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	targetUserID := r.PathValue("id")
	if targetUserID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "User ID is required"}`))
		return
	}

	// 1. Safety check: prevent self deletion
	if currentUserID == targetUserID {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Anda tidak dapat menghapus akun Anda sendiri"}`))
		return
	}

	switch r.Method {
	case http.MethodDelete:
		_, err = db.DB.Exec("DELETE FROM users WHERE id = $1", targetUserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to delete user"}`))
			return
		}
		w.Write([]byte(`{"success": true, "message": "User deleted successfully"}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}
