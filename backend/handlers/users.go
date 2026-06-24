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

// UserDashboardStatsHandler retrieves tasks stats for the user's dashboard
func UserDashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
		return
	}

	// 1. Get user name
	var name string
	err = db.DB.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to query user profile"}`))
		return
	}

	// 2. Query all modules of projects user has access to
	rows, err := db.DB.Query(`
		SELECT m.nodes 
		FROM modules m
		JOIN projects p ON m.project_id = p.id
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.owner_id = $1 OR pm.user_id = $1
	`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Failed to query workflow tasks"}`))
		return
	}
	defer rows.Close()

	type NodeRole struct {
		Assignee string `json:"assignee"`
		Status   string `json:"status"`
	}
	type NodeFacet struct {
		Uiux     *NodeRole `json:"uiux"`
		Frontend *NodeRole `json:"frontend"`
		Backend  *NodeRole `json:"backend"`
	}
	type NodeJson struct {
		Roles NodeFacet `json:"roles"`
	}

	myTasksCount := 0
	completedTasksCount := 0
	totalTasksCount := 0

	for rows.Next() {
		var nodesStr string
		if err := rows.Scan(&nodesStr); err != nil {
			continue
		}
		var nodes []NodeJson
		if err := json.Unmarshal([]byte(nodesStr), &nodes); err != nil {
			continue
		}
		for _, node := range nodes {
			// Check UI/UX
			if node.Roles.Uiux != nil {
				totalTasksCount++
				if node.Roles.Uiux.Status == "done" {
					completedTasksCount++
				} else if node.Roles.Uiux.Assignee == name {
					myTasksCount++
				}
			}
			// Check Frontend
			if node.Roles.Frontend != nil {
				totalTasksCount++
				if node.Roles.Frontend.Status == "done" {
					completedTasksCount++
				} else if node.Roles.Frontend.Assignee == name {
					myTasksCount++
				}
			}
			// Check Backend
			if node.Roles.Backend != nil {
				totalTasksCount++
				if node.Roles.Backend.Status == "done" {
					completedTasksCount++
				} else if node.Roles.Backend.Assignee == name {
					myTasksCount++
				}
			}
		}
	}

	completionRate := 0
	if totalTasksCount > 0 {
		completionRate = (completedTasksCount * 100) / totalTasksCount
	}

	resp := struct {
		MyTasksCount   int `json:"my_tasks_count"`
		CompletionRate int `json:"completion_rate"`
	}{
		MyTasksCount:   myTasksCount,
		CompletionRate: completionRate,
	}

	json.NewEncoder(w).Encode(resp)
}
