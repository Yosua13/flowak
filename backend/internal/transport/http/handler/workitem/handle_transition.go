package workitem

import (
	"errors"
	"net/http"

	"backend/db"
	domainWorkItem "backend/internal/domain/workitem"
	repoWorkItem "backend/internal/repository/postgres/workitem"
	usecaseWorkItem "backend/internal/usecase/workitem"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// GlobalEventPublisher is the event publisher instance used by HTTP transition handlers.
var GlobalEventPublisher domainWorkItem.EventPublisher

// SetEventPublisher configures the global event publisher for work items.
func SetEventPublisher(pub domainWorkItem.EventPublisher) {
	GlobalEventPublisher = pub
}

// HandleTransition processes a status transition request with state-machine validation and optimistic locking.
func HandleTransition(c *gin.Context) {
	key := c.Param("key")

	var projectID string
	err := db.DB.QueryRowContext(c.Request.Context(), `SELECT project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, key).Scan(&projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityWorkItem) {
		return
	}

	userID, _ := middleware.GetUserID(c)

	var req models.WorkItemTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "error_code": "validation_error"})
		return
	}

	repo := repoWorkItem.NewPostgresWorkItemRepository(db.DB)
	uc := usecaseWorkItem.NewTransitionWorkItemUseCase(repo, GlobalEventPublisher)

	updated, err := uc.Execute(c.Request.Context(), userID, key, usecaseWorkItem.TransitionWorkItemInput{
		Status:     req.Status,
		RowVersion: req.RowVersion,
		Note:       req.Note,
		Resolution: req.Resolution,
	})
	if err != nil {
		if errors.Is(err, domainWorkItem.ErrOptimisticLockConflict) {
			c.JSON(http.StatusConflict, gin.H{
				"error":      "work item has changed",
				"error_code": "work_item_version_conflict",
			})
			return
		}
		if errors.Is(err, domainWorkItem.ErrWorkItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"error_code": "validation_error",
		})
		return
	}

	c.JSON(http.StatusOK, updated)
}
