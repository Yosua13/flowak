package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"backend/db"
	"backend/middleware"
	"backend/models"
)

// ProjectsHandler handles GET /api/projects and POST /api/projects
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Retrieve all projects owned by the user
		rows, err := db.DB.Query("SELECT id, name, description, owner_id, created_at FROM projects WHERE owner_id = $1 ORDER BY created_at DESC", userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Database query error"}`))
			return
		}
		defer rows.Close()

		projects := []models.Project{}
		for rows.Next() {
			var p models.Project
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to parse projects"}`))
				return
			}
			projects = append(projects, p)
		}

		json.NewEncoder(w).Encode(projects)

	case http.MethodPost:
		// Create a new project
		var req models.ProjectCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid request body"}`))
			return
		}

		name := strings.TrimSpace(req.Name)
		description := strings.TrimSpace(req.Description)

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Project name is required"}`))
			return
		}

		projectID := "proj_" + GenerateUUID()

		// Start a transaction to insert both project and a default module
		tx, err := db.DB.Begin()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to start transaction"}`))
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO projects (id, name, description, owner_id) VALUES ($1, $2, $3, $4)",
			projectID, name, description, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to create project record"}`))
			return
		}

		// Seed a default workflow module inside the project
		moduleID := "mod_" + GenerateUUID()
		defaultNodesJSON := "[]"
		defaultEdgesJSON := "[]"
		_, err = tx.Exec("INSERT INTO modules (id, project_id, name, description, nodes, edges, schema_version) VALUES ($1, $2, $3, $4, $5, $6, 1)",
			moduleID, projectID, "Alur Kerja Utama", "Modul alur kerja default untuk proyek baru.", defaultNodesJSON, defaultEdgesJSON)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to create default workflow module"}`))
			return
		}

		if err := tx.Commit(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to commit transaction"}`))
			return
		}

		// Return the created project
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"success": true, "project_id": "%s", "module_id": "%s"}`, projectID, moduleID)))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}

// ProjectDetailHandler handles GET /api/projects/{id} and DELETE /api/projects/{id}
func ProjectDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	projectID := r.PathValue("id")
	if projectID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Project ID is required"}`))
		return
	}

	// 1. Authorization check: Verify user owns the project
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Project not found"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database error checking ownership"}`))
		return
	}

	if ownerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "You do not have permission to access this project"}`))
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Retrieve project with its modules
		var p models.Project
		err = db.DB.QueryRow("SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1", projectID).
			Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to get project info"}`))
			return
		}

		// Get project modules
		rows, err := db.DB.Query("SELECT id, project_id, name, description, nodes, edges, schema_version, created_at FROM modules WHERE project_id = $1", projectID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Database query error for modules"}`))
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
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to parse modules"}`))
				return
			}
			modulesList = append(modulesList, m)
		}

		resp := ProjectResponse{
			Project: p,
			Modules: modulesList,
		}
		json.NewEncoder(w).Encode(resp)

	case http.MethodDelete:
		// Delete the project (cascades to modules automatically due to foreign key)
		_, err = db.DB.Exec("DELETE FROM projects WHERE id = $1", projectID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to delete project"}`))
			return
		}
		w.Write([]byte(`{"success": true, "message": "Project deleted successfully"}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}

// ProjectModulesHandler handles GET /api/projects/{id}/modules and POST /api/projects/{id}/modules
func ProjectModulesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	projectID := r.PathValue("id")

	// 1. Authorization check
	var ownerID string
	err = db.DB.QueryRow("SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Project not found"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database error checking ownership"}`))
		return
	}

	if ownerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "You do not have permission to access this project"}`))
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query("SELECT id, project_id, name, description, nodes, edges, schema_version, created_at FROM modules WHERE project_id = $1 ORDER BY created_at ASC", projectID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Database query error for modules"}`))
			return
		}
		defer rows.Close()

		modulesList := []models.Module{}
		for rows.Next() {
			var m models.Module
			if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Description, &m.Nodes, &m.Edges, &m.SchemaVersion, &m.CreatedAt); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to parse modules"}`))
				return
			}
			modulesList = append(modulesList, m)
		}
		json.NewEncoder(w).Encode(modulesList)

	case http.MethodPost:
		// Create new module inside project
		var req models.ModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid request body"}`))
			return
		}

		name := strings.TrimSpace(req.Name)
		description := strings.TrimSpace(req.Description)

		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Module name is required"}`))
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
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to save module"}`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"success": true, "module_id": "%s"}`, moduleID)))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}

// ModuleDetailHandler handles PUT /api/modules/{id} and DELETE /api/modules/{id}
func ModuleDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := middleware.GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	moduleID := r.PathValue("id")

	// 1. Authorization: verify this module belongs to a project owned by the user
	var ownerID string
	var projectID string
	err = db.DB.QueryRow("SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1", moduleID).
		Scan(&ownerID, &projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Module not found"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Database error checking ownership"}`))
		return
	}

	if ownerID != userID {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "You do not have permission to access this module"}`))
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req models.ModuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Invalid request body"}`))
			return
		}

		// Update fields if provided
		if req.Name != "" {
			_, err = db.DB.Exec("UPDATE modules SET name = $1 WHERE id = $2", req.Name, moduleID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to update module name"}`))
				return
			}
		}

		if req.Description != "" {
			_, err = db.DB.Exec("UPDATE modules SET description = $1 WHERE id = $2", req.Description, moduleID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to update module description"}`))
				return
			}
		}

		if req.Nodes != nil {
			nodesBytes, err := json.Marshal(req.Nodes)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid nodes format"}`))
				return
			}
			_, err = db.DB.Exec("UPDATE modules SET nodes = $1 WHERE id = $2", string(nodesBytes), moduleID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to update nodes data"}`))
				return
			}
		}

		if req.Edges != nil {
			edgesBytes, err := json.Marshal(req.Edges)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid edges format"}`))
				return
			}
			_, err = db.DB.Exec("UPDATE modules SET edges = $1 WHERE id = $2", string(edgesBytes), moduleID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Failed to update edges data"}`))
				return
			}
		}

		w.Write([]byte(`{"success": true, "message": "Module updated successfully"}`))

	case http.MethodDelete:
		_, err = db.DB.Exec("DELETE FROM modules WHERE id = $1", moduleID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to delete module"}`))
			return
		}
		w.Write([]byte(`{"success": true, "message": "Module deleted successfully"}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed"}`))
	}
}
