package workitem

import (
	"net/http"

	"backend/db"
	repoWorkItem "backend/internal/repository/postgres/workitem"
	usecaseWorkItem "backend/internal/usecase/workitem"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// HandleCreate handles work item creation requests under a project.
func HandleCreate(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	projectID := c.Param("id")
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityWorkItem) {
		return
	}

	var req models.WorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "error_code": "validation_error"})
		return
	}

	repo := repoWorkItem.NewPostgresWorkItemRepository(db.DB)
	uc := usecaseWorkItem.NewCreateWorkItemUseCase(repo)

	item, err := uc.Execute(c.Request.Context(), usecaseWorkItem.CreateWorkItemInput{
		ProjectID:     projectID,
		ReporterID:    userID,
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
	})
	if err != nil {
		// Differentiate validation errors from system errors
		msg := err.Error()
		if msg == "project not found" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg, "error_code": "validation_error"})
		return
	}

	c.JSON(http.StatusCreated, item)
}
