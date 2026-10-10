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

// HandlePublishBaseline handles POST /api/modules/:id/publish to snapshot current graph into an immutable baseline.
func HandlePublishBaseline(c *gin.Context) {
	moduleID := c.Param("id")

	var projectID string
	if err := db.DB.QueryRowContext(c.Request.Context(), `SELECT project_id FROM modules WHERE id=$1 AND status='active'`, moduleID).Scan(&projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error reading module"})
		return
	}

	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	repo := pgModule.NewModulePostgresRepo(db.DB)

	baseline, err := usecaseModule.ExecutePublishBaseline(c.Request.Context(), repo, moduleID, actorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish baseline"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      baseline.ID,
		"version": baseline.Version,
		"status":  "published",
	})
}
