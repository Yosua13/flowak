package workitem

import (
	"net/http"
	"strconv"

	"backend/db"
	domainWorkItem "backend/internal/domain/workitem"
	repoWorkItem "backend/internal/repository/postgres/workitem"
	usecaseWorkItem "backend/internal/usecase/workitem"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// HandleList handles listing and paginating work items for a specific project.
func HandleList(c *gin.Context) {
	projectID := c.Param("id")
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}

	status := c.Query("status")
	assignee := c.Query("assignee_id")
	nodeID := c.Query("node_id")

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1-100", "error_code": "validation_error"})
			return
		}
		limit = n
	}

	sortOrder := c.DefaultQuery("sort", "newest")
	if sortOrder != "newest" && sortOrder != "oldest" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be newest or oldest", "error_code": "validation_error"})
		return
	}

	pageMode := c.Query("page") == "1"
	cursor := c.Query("cursor")

	repo := repoWorkItem.NewPostgresWorkItemRepository(db.DB)
	uc := usecaseWorkItem.NewListWorkItemsUseCase(repo)

	result, err := uc.Execute(c.Request.Context(), domainWorkItem.ListFilter{
		ProjectID:  projectID,
		Status:     status,
		AssigneeID: assignee,
		NodeID:     nodeID,
		Limit:      limit,
		Sort:       sortOrder,
		PageMode:   pageMode,
		Cursor:     cursor,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "error_code": "validation_error"})
		return
	}

	if pageMode {
		c.JSON(http.StatusOK, gin.H{
			"items":       result.Items,
			"next_cursor": result.NextCursor,
			"sort":        result.Sort,
		})
		return
	}

	c.JSON(http.StatusOK, result.Items)
}
