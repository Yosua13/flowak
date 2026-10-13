package module

import (
	"database/sql"
	"errors"
	"net/http"

	"backend/db"
	pgModule "backend/internal/repository/postgres/module"
	usecaseModule "backend/internal/usecase/module"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// HandleGetGraph handles GET /api/modules/:id/graph to fetch a module's canvas graph.
func HandleGetGraph(c *gin.Context) {
	_, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	moduleID := c.Param("id")

	var projectID string
	err = db.DB.QueryRowContext(c.Request.Context(), "SELECT project_id FROM modules WHERE id = $1 AND status = 'active'", moduleID).Scan(&projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking module"})
		return
	}

	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}

	repo := pgModule.NewModulePostgresRepo(db.DB)
	graph, err := usecaseModule.ExecuteGetModuleGraph(c.Request.Context(), repo, middleware.GlobalRedisClient, moduleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load module graph"})
		return
	}

	c.JSON(http.StatusOK, graph)
}
