package workitem

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/db"
	domainWorkItem "backend/internal/domain/workitem"
	repoWorkItem "backend/internal/repository/postgres/workitem"
	usecaseWorkItem "backend/internal/usecase/workitem"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// HandleUpdate handles partial work item update requests with optimistic locking.
func HandleUpdate(c *gin.Context) {
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

	var req models.WorkItemRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "error_code": "validation_error"})
		return
	}

	var supplied map[string]json.RawMessage
	if cached, ok := c.Get(gin.BodyBytesKey); ok {
		if raw, ok := cached.([]byte); ok {
			_ = json.Unmarshal(raw, &supplied)
		}
	}

	wasCleared := func(field string) bool {
		raw, ok := supplied[field]
		value := strings.TrimSpace(string(raw))
		return ok && (value == "null" || value == `""`)
	}

	repo := repoWorkItem.NewPostgresWorkItemRepository(db.DB)
	uc := usecaseWorkItem.NewUpdateWorkItemUseCase(repo)

	updated, err := uc.Execute(c.Request.Context(), userID, key, usecaseWorkItem.UpdateWorkItemInput{
		RowVersion:    req.RowVersion,
		Type:          req.Type,
		Title:         req.Title,
		Description:   req.Description,
		Priority:      req.Priority,
		Points:        req.Points,
		Status:        req.Status,
		ModuleID:      req.ModuleID,
		NodeID:        req.NodeID,
		FacetKey:      req.FacetKey,
		ParentID:      req.ParentID,
		AssigneeID:    req.AssigneeID,
		StartDate:     req.StartDate,
		DueDate:       req.DueDate,
		BlockedReason: req.BlockedReason,
		Resolution:    req.Resolution,
		WasCleared:    wasCleared,
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
