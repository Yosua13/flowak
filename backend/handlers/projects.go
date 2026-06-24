package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// hasProjectAccess checks if a user is the owner or a member of a project
func hasProjectAccess(userID, projectID string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects WHERE id = $1 AND owner_id = $2
			UNION
			SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&exists)
	return exists, err
}

// GetProjectsHandler handles GET /api/projects
func GetProjectsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve all projects owned by the user or where the user is a member
	rows, err := db.DB.Query(`
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at 
		FROM projects p 
		LEFT JOIN project_members pm ON p.id = pm.project_id 
		WHERE p.owner_id = $1 OR pm.user_id = $1 
		ORDER BY p.created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
		return
	}
	defer rows.Close()

	projects := []models.Project{}
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse projects"})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, projects)
}

// CreateProjectHandler handles POST /api/projects
func CreateProjectHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Only PM can create new projects
	role, err := middleware.GetUserRole(c)
	if err != nil || role != "pm" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Project Manager (PM) yang dapat membuat proyek baru"})
		return
	}

	var req models.ProjectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name is required"})
		return
	}

	projectID := "proj_" + GenerateUUID()

	// Start a transaction to insert both project and a default module
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO projects (id, name, description, owner_id) VALUES ($1, $2, $3, $4)",
		projectID, name, description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project record"})
		return
	}

	// Seed a default workflow module inside the project
	moduleID := "mod_" + GenerateUUID()
	defaultNodesJSON := "[]"
	defaultEdgesJSON := "[]"
	_, err = tx.Exec("INSERT INTO modules (id, project_id, name, description, nodes, edges, schema_version) VALUES ($1, $2, $3, $4, $5, $6, 1)",
		moduleID, projectID, "Alur Kerja Utama", "Modul alur kerja default untuk proyek baru.", defaultNodesJSON, defaultEdgesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create default workflow module"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"project_id": projectID,
		"module_id":  moduleID,
	})
}

// GetProjectDetailHandler handles GET /api/projects/:id
func GetProjectDetailHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Authorization check
	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this project"})
		return
	}

	// Retrieve project
	var p models.Project
	err = db.DB.QueryRow("SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1", projectID).
		Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project info"})
		return
	}

	// Get project modules
	rows, err := db.DB.Query("SELECT id, project_id, name, description, nodes, edges, schema_version, created_at FROM modules WHERE project_id = $1", projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error for modules"})
		return
	}
	defer rows.Close()

	type ProjectResponse struct {
		models.Project
		Modules []models.Module `json:"modules"`
	}

	modulesList := []models.Module{}
	for rows.Next() {
		var m models.Module
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Description, &m.Nodes, &m.Edges, &m.SchemaVersion, &m.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse modules"})
			return
		}
		modulesList = append(modulesList, m)
	}

	resp := ProjectResponse{
		Project: p,
		Modules: modulesList,
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteProjectHandler handles DELETE /api/projects/:id
func DeleteProjectHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Authorization check
	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this project"})
		return
	}

	// Verify ownership: only owner can delete the project
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking ownership"})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the project owner can delete this project"})
		return
	}

	// Delete the project (cascades automatically)
	_, err = db.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Project deleted successfully"})
}

// GetProjectModulesHandler handles GET /api/projects/:id/modules
func GetProjectModulesHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")

	// Authorization check
	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this project"})
		return
	}

	rows, err := db.DB.Query("SELECT id, project_id, name, description, nodes, edges, schema_version, created_at FROM modules WHERE project_id = $1 ORDER BY created_at ASC", projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error for modules"})
		return
	}
	defer rows.Close()

	modulesList := []models.Module{}
	for rows.Next() {
		var m models.Module
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Description, &m.Nodes, &m.Edges, &m.SchemaVersion, &m.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse modules"})
			return
		}
		modulesList = append(modulesList, m)
	}
	c.JSON(http.StatusOK, modulesList)
}

// CreateProjectModuleHandler handles POST /api/projects/:id/modules
func CreateProjectModuleHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")

	// Authorization check
	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this project"})
		return
	}

	// Only project owner can create modules
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking project owner"})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the project owner can create modules"})
		return
	}

	var req models.ModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Module name is required"})
		return
	}

	moduleID := "mod_" + GenerateUUID()

	// Default canvas values
	nodesJSON := "[]"
	edgesJSON := "[]"

	if req.Nodes != nil {
		nb, _ := json.Marshal(req.Nodes)
		nodesJSON = string(nb)
	}
	if req.Edges != nil {
		eb, _ := json.Marshal(req.Edges)
		edgesJSON = string(eb)
	}

	_, err = db.DB.Exec("INSERT INTO modules (id, project_id, name, description, nodes, edges, schema_version) VALUES ($1, $2, $3, $4, $5, $6, 1)",
		moduleID, projectID, name, description, nodesJSON, edgesJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save module"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "module_id": moduleID})
}

// UpdateModuleHandler handles PUT /api/modules/:id
func UpdateModuleHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	moduleID := c.Param("id")

	// Authorization
	var ownerID string
	var projectID string
	err = db.DB.QueryRow("SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1", moduleID).
		Scan(&ownerID, &projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking ownership"})
		return
	}

	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this module"})
		return
	}

	var req models.ModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		_, err = db.DB.Exec("UPDATE modules SET name = $1 WHERE id = $2", req.Name, moduleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module name"})
			return
		}
	}

	if req.Description != "" {
		_, err = db.DB.Exec("UPDATE modules SET description = $1 WHERE id = $2", req.Description, moduleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module description"})
			return
		}
	}

	if req.Nodes != nil {
		nodesBytes, err := json.Marshal(req.Nodes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid nodes format"})
			return
		}
		_, err = db.DB.Exec("UPDATE modules SET nodes = $1 WHERE id = $2", string(nodesBytes), moduleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update nodes data"})
			return
		}
	}

	if req.Edges != nil {
		edgesBytes, err := json.Marshal(req.Edges)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid edges format"})
			return
		}
		_, err = db.DB.Exec("UPDATE modules SET edges = $1 WHERE id = $2", string(edgesBytes), moduleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update edges data"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Module updated successfully"})
}

// DeleteModuleHandler handles DELETE /api/modules/:id
func DeleteModuleHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	moduleID := c.Param("id")

	// Authorization
	var ownerID string
	var projectID string
	err = db.DB.QueryRow("SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1", moduleID).
		Scan(&ownerID, &projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking ownership"})
		return
	}

	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this module"})
		return
	}

	// Only project owner (PM) can delete modules
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the project owner can delete modules"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM modules WHERE id = $1", moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Module deleted successfully"})
}

// GetProjectMembersHandler handles GET /api/projects/:id/members
func GetProjectMembersHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Verify project exists
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking project"})
		return
	}

	// Check access
	hasAccess, err := hasProjectAccess(userID, projectID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view members of this project"})
		return
	}

	rows, err := db.DB.Query(`
		SELECT u.id, u.name, u.email, u.role, u.created_at
		FROM users u
		JOIN project_members pm ON u.id = pm.user_id
		WHERE pm.project_id = $1
		ORDER BY u.name ASC`, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error for members"})
		return
	}
	defer rows.Close()

	members := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse members"})
			return
		}
		members = append(members, u)
	}

	c.JSON(http.StatusOK, members)
}

// AddProjectMemberHandler handles POST /api/projects/:id/members
func AddProjectMemberHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// Get owner
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking project"})
		return
	}

	// Only project owner (PM) can add members
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the project owner can add members"})
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	targetUserID := strings.TrimSpace(req.UserID)
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Ensure target user exists
	var userExists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", targetUserID).Scan(&userExists)
	if err != nil || !userExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Check if already member
	var isMember bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)", projectID, targetUserID).Scan(&isMember)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking membership"})
		return
	}
	if isMember {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this project"})
		return
	}

	_, err = db.DB.Exec("INSERT INTO project_members (project_id, user_id) VALUES ($1, $2)", projectID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Member added successfully"})
}

// RemoveProjectMemberHandler handles DELETE /api/projects/:id/members/:userId
func RemoveProjectMemberHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	targetUserID := c.Param("userId")
	if projectID == "" || targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and User ID are required"})
		return
	}

	// Only project owner (PM) can remove members
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking project"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the project owner can remove members"})
		return
	}

	_, err = db.DB.Exec("DELETE FROM project_members WHERE project_id = $1 AND user_id = $2", projectID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member from project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Member removed successfully"})
}
