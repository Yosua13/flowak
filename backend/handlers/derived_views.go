package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// DerivedViewData is a read model for project views. It deliberately joins only
// normalized records so derived screens never use the deprecated graph snapshot.
type DerivedViewData struct {
	WorkItems    []models.WorkItem              `json:"work_items"`
	History      []models.WorkItemStatusHistory `json:"history"`
	FacetReviews []models.FacetReviewDue        `json:"facet_reviews"`
	Baselines    []models.ModuleBaseline        `json:"baselines"`
	Comments     []models.DecisionComment       `json:"comments"`
	Evidence     []models.DerivedEvidence       `json:"evidence"`
	GeneratedAt  time.Time                      `json:"generated_at"`
}

func derivedItems(c *gin.Context, projectID, moduleID, facetKey string) ([]models.WorkItem, error) {
	args := []any{projectID}
	where := "project_id=$1 AND deleted_at IS NULL"
	if moduleID != "" {
		args = append(args, moduleID)
		where += " AND module_id=$" + strconv.Itoa(len(args))
	}
	if facetKey != "" {
		args = append(args, facetKey)
		where += " AND facet_key=$" + strconv.Itoa(len(args))
	}
	rows, err := db.DB.Query(`SELECT `+workItemFields+` FROM work_items WHERE `+where+` ORDER BY sequence DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.WorkItem{}
	for rows.Next() {
		var item models.WorkItem
		if err := rows.Scan(&item.ID, &item.Key, &item.ProjectID, &item.ModuleID, &item.NodeID, &item.FacetKey, &item.ParentID, &item.Type, &item.Title, &item.Description, &item.Priority, &item.Points, &item.Status, &item.AssigneeID, &item.ReporterID, &item.StartDate, &item.DueDate, &item.BlockedReason, &item.Resolution, &item.RowVersion, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func placeholders(ids []string, start int) (string, []any) {
	parts, args := make([]string, len(ids)), make([]any, len(ids))
	for index, id := range ids {
		parts[index], args[index] = "$"+strconv.Itoa(start+index), id
	}
	return strings.Join(parts, ","), args
}

func validDerivedFacet(value string) bool {
	return value == "" || value == "business" || value == "uiux" || value == "frontend" || value == "backend"
}

// GetDerivedViewDataHandler returns the common read model used by status,
// calendar, analytics, documents, and exports. It intentionally excludes API
// response bodies and environment values from evidence.
func GetDerivedViewDataHandler(c *gin.Context) {
	projectID, moduleID, facetKey := c.Param("id"), c.Query("module_id"), c.Query("facet_key")
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	if !validDerivedFacet(facetKey) {
		validationError(c, "invalid facet_key")
		return
	}
	if moduleID != "" {
		var exists bool
		if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM modules WHERE id=$1 AND project_id=$2 AND status='active')`, moduleID, projectID).Scan(&exists); err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "module not found"})
			return
		}
	}
	items, err := derivedItems(c, projectID, moduleID, facetKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read work items"})
		return
	}
	data := DerivedViewData{WorkItems: items, History: []models.WorkItemStatusHistory{}, FacetReviews: []models.FacetReviewDue{}, Baselines: []models.ModuleBaseline{}, Comments: []models.DecisionComment{}, Evidence: []models.DerivedEvidence{}, GeneratedAt: time.Now().UTC()}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	if len(ids) > 0 {
		marks, args := placeholders(ids, 1)
		rows, err := db.DB.Query(`SELECT work_item_id,from_status,to_status,note,created_at FROM work_item_status_history WHERE work_item_id IN (`+marks+`) ORDER BY created_at`, args...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read work item history"})
			return
		}
		for rows.Next() {
			var entry models.WorkItemStatusHistory
			if err := rows.Scan(&entry.WorkItemID, &entry.FromStatus, &entry.ToStatus, &entry.Note, &entry.CreatedAt); err != nil {
				rows.Close()
				c.JSON(500, gin.H{"error": "failed to parse work item history"})
				return
			}
			data.History = append(data.History, entry)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			c.JSON(500, gin.H{"error": "failed to read work item history"})
			return
		}
		rows.Close()
		rows, err = db.DB.Query(`SELECT id,work_item_id,body,resolved_at,created_at FROM comments WHERE deleted_at IS NULL AND work_item_id IN (`+marks+`) ORDER BY created_at`, args...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read work item comments"})
			return
		}
		for rows.Next() {
			var entry models.DecisionComment
			if err := rows.Scan(&entry.ID, &entry.WorkItemID, &entry.Body, &entry.ResolvedAt, &entry.CreatedAt); err != nil {
				rows.Close()
				c.JSON(500, gin.H{"error": "failed to parse work item comments"})
				return
			}
			data.Comments = append(data.Comments, entry)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			c.JSON(500, gin.H{"error": "failed to read work item comments"})
			return
		}
		rows.Close()
		rows, err = db.DB.Query(`SELECT id,file_name,work_item_id,node_id,created_at FROM attachments WHERE work_item_id IN (`+marks+`) ORDER BY created_at DESC`, args...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read evidence"})
			return
		}
		for rows.Next() {
			var entry models.DerivedEvidence
			if err := rows.Scan(&entry.ID, &entry.Label, &entry.WorkItemID, &entry.NodeID, &entry.CreatedAt); err != nil {
				rows.Close()
				c.JSON(500, gin.H{"error": "failed to parse evidence"})
				return
			}
			entry.Kind = "attachment"
			data.Evidence = append(data.Evidence, entry)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			c.JSON(500, gin.H{"error": "failed to read evidence"})
			return
		}
		rows.Close()
	}
	if moduleID != "" {
		roleFilter, roleArgs := "", []any{moduleID}
		if facetKey == "uiux" || facetKey == "frontend" || facetKey == "backend" {
			roleFilter = " AND r.role_key=$2"
			roleArgs = append(roleArgs, facetKey)
		}
		rows, err := db.DB.Query(`SELECT n.id,n.label,r.role_key,r.readiness,r.status,r.due_date FROM node_role_tasks r JOIN workflow_nodes n ON n.id=r.node_id WHERE n.module_id=$1 AND n.deleted_at IS NULL AND r.due_date IS NOT NULL`+roleFilter+` ORDER BY r.due_date`, roleArgs...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read facet reviews"})
			return
		}
		for rows.Next() {
			var review models.FacetReviewDue
			if err := rows.Scan(&review.NodeID, &review.NodeLabel, &review.RoleKey, &review.Readiness, &review.Status, &review.DueDate); err != nil {
				rows.Close()
				c.JSON(500, gin.H{"error": "failed to parse facet reviews"})
				return
			}
			data.FacetReviews = append(data.FacetReviews, review)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			c.JSON(500, gin.H{"error": "failed to read facet reviews"})
			return
		}
		rows.Close()
		rows, err = db.DB.Query(`SELECT module_id,version,created_at FROM module_versions WHERE module_id=$1 ORDER BY version DESC`, moduleID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read module baseline"})
			return
		}
		for rows.Next() {
			var baseline models.ModuleBaseline
			if err := rows.Scan(&baseline.ModuleID, &baseline.Version, &baseline.CreatedAt); err != nil {
				rows.Close()
				c.JSON(500, gin.H{"error": "failed to parse module baseline"})
				return
			}
			data.Baselines = append(data.Baselines, baseline)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			c.JSON(500, gin.H{"error": "failed to read module baseline"})
			return
		}
		rows.Close()
	}
	c.JSON(http.StatusOK, data)
}
