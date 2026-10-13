package module

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"backend/db"
	domainModule "backend/internal/domain/module"
	pgModule "backend/internal/repository/postgres/module"
	usecaseModule "backend/internal/usecase/module"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// HandleSyncGraph handles PUT /api/modules/:id to update module metadata and graph entities.
func HandleSyncGraph(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	moduleID := c.Param("id")

	// Verify project ownership and capability
	var ownerID, projectID string
	err = db.DB.QueryRowContext(c.Request.Context(), "SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1", moduleID).
		Scan(&ownerID, &projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking ownership"})
		return
	}

	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}

	var req models.ModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Convert nodes & edges to domain entities
	var domainNodes []domainModule.Node
	if req.Nodes != nil {
		bytes, err := json.Marshal(req.Nodes)
		if err == nil {
			_ = json.Unmarshal(bytes, &domainNodes)
		}
	}

	var domainEdges []domainModule.Edge
	if req.Edges != nil {
		bytes, err := json.Marshal(req.Edges)
		if err == nil {
			_ = json.Unmarshal(bytes, &domainEdges)
		}
	}

	var domainDelNodes []domainModule.GraphDelete
	for _, d := range req.DeletedNodes {
		domainDelNodes = append(domainDelNodes, domainModule.GraphDelete{ID: d.ID, RowVersion: d.RowVersion})
	}

	var domainDelEdges []domainModule.GraphDelete
	for _, d := range req.DeletedEdges {
		domainDelEdges = append(domainDelEdges, domainModule.GraphDelete{ID: d.ID, RowVersion: d.RowVersion})
	}

	repo := pgModule.NewModulePostgresRepo(db.DB)
	syncReq := usecaseModule.SyncModuleGraphRequest{
		Name:         req.Name,
		Description:  req.Description,
		Nodes:        domainNodes,
		Edges:        domainEdges,
		DeletedNodes: domainDelNodes,
		DeletedEdges: domainDelEdges,
	}

	graph, err := usecaseModule.ExecuteSyncModuleGraph(c.Request.Context(), repo, middleware.GlobalRedisClient, moduleID, userID, syncReq)
	if err != nil {
		if errors.Is(err, domainModule.ErrGraphConflict) {
			c.JSON(http.StatusConflict, gin.H{"error": "node version conflict", "error_code": "GRAPH_VERSION_CONFLICT"})
			return
		}
		if errors.Is(err, domainModule.ErrInvalidGraphReference) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "error_code": "INVALID_GRAPH_REFERENCE"})
			return
		}
		if errors.Is(err, domainModule.ErrCycleDetected) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "error_code": "GRAPH_CYCLE_DETECTED"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Module updated successfully",
		"nodes":   graph.Nodes,
		"edges":   graph.Edges,
	})
}
