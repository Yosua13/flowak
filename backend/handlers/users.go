package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUsersHandler handles GET /api/users
func GetUsersHandler(c *gin.Context) {
	_, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch all users in the system to list as team contributors
	rows, err := db.DB.Query("SELECT id, name, email, role FROM users WHERE status = 'active' ORDER BY name ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
		return
	}
	defer rows.Close()

	var usersList []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse users"})
			return
		}
		usersList = append(usersList, u)
	}

	c.JSON(http.StatusOK, usersList)
}

// PostUsersHandler handles POST /api/users (Add a contributor)
func PostUsersHandler(c *gin.Context) {
	_, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	role := strings.TrimSpace(req.Role)

	if name == "" || email == "" || role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	if !emailRegex.MatchString(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
		return
	}

	if role != "pm" && role != "uiux" && role != "frontend" && role != "backend" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check duplicate
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email availability"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered"})
		return
	}

	tempPassword := GenerateTempPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption error"})
		return
	}

	newUserID := "usr_" + GenerateUUID()
	_, err = db.DB.Exec("INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)",
		newUserID, name, email, string(hashedPassword), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	_, _ = db.DB.Exec(`
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ('org_default', $1, $2)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role, status = 'active'
	`, newUserID, role)

	c.JSON(http.StatusCreated, gin.H{"success": true, "user_id": newUserID, "temporary_password": tempPassword})
}

// DeleteUserHandler handles DELETE /api/users/:id
func DeleteUserHandler(c *gin.Context) {
	currentUserID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// 1. Safety check: prevent self deletion
	if currentUserID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda tidak dapat menghapus akun Anda sendiri"})
		return
	}

	_, err = db.DB.Exec("UPDATE users SET status = 'inactive', updated_at = CURRENT_TIMESTAMP WHERE id = $1", targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted successfully"})
}

// UserDashboardStatsHandler retrieves tasks stats for the user's dashboard
func UserDashboardStatsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 1. Get user name
	var name string
	err = db.DB.QueryRow("SELECT name FROM users WHERE id = $1", userID).Scan(&name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user profile"})
		return
	}

	var normalizedMyTasks int
	var normalizedCompletedTasks int
	var normalizedTotalTasks int
	err = db.DB.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE t.assignee_id = $1 AND t.status <> 'done') AS my_tasks,
			COUNT(*) FILTER (WHERE t.status = 'done') AS completed_tasks,
			COUNT(*) AS total_tasks
		FROM node_role_tasks t
		JOIN workflow_nodes n ON n.id = t.node_id
		JOIN modules m ON m.id = n.module_id
		JOIN projects p ON p.id = m.project_id
		LEFT JOIN project_members pm ON pm.project_id = p.id
		WHERE p.status = 'active'
		  AND m.status = 'active'
		  AND (p.owner_id = $1 OR pm.user_id = $1)
	`, userID).Scan(&normalizedMyTasks, &normalizedCompletedTasks, &normalizedTotalTasks)
	if err == nil && normalizedTotalTasks > 0 {
		completionRate := (normalizedCompletedTasks * 100) / normalizedTotalTasks
		c.JSON(http.StatusOK, gin.H{
			"my_tasks_count":  normalizedMyTasks,
			"completion_rate": completionRate,
		})
		return
	}

	// 2. Query all modules of projects user has access to
	rows, err := db.DB.Query(`
		SELECT m.nodes 
		FROM modules m
		JOIN projects p ON m.project_id = p.id
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.status = 'active'
		  AND m.status = 'active'
		  AND (p.owner_id = $1 OR pm.user_id = $1)
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query workflow tasks"})
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

	c.JSON(http.StatusOK, resp)
}
