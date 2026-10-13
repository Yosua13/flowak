package workitem

import (
	"database/sql"
	"errors"
	"net/http"

	"backend/db"
	domainWorkItem "backend/internal/domain/workitem"
	repoWorkItem "backend/internal/repository/postgres/workitem"
	usecaseWorkItem "backend/internal/usecase/workitem"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// HandleGetDetail handles requests for reading a single work item by its key.
func HandleGetDetail(c *gin.Context) {
	key := c.Param("key")

	var projectID string
	err := db.DB.QueryRowContext(c.Request.Context(), `SELECT project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, key).Scan(&projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}

	repo := repoWorkItem.NewPostgresWorkItemRepository(db.DB)
	uc := usecaseWorkItem.NewGetWorkItemDetailUseCase(repo)

	item, err := uc.Execute(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, domainWorkItem.ErrWorkItemNotFound) || errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read work item"})
		return
	}

	c.JSON(http.StatusOK, item)
}
