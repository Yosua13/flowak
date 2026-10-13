package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"backend/db"
	workitemHandler "backend/internal/transport/http/handler/workitem"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

var workItemTypes = map[string]bool{"Story": true, "Task": true, "Bug": true, "Review": true, "Research": true, "Subtask": true}
var workItemStatuses = map[string]bool{"Backlog": true, "Ready": true, "In Progress": true, "In Review": true, "Blocked": true, "Done": true, "Canceled": true}
var workItemPriorities = map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
var allowedPoints = map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true, 13: true}

type workItemCursor struct {
	Project  string `json:"project"`
	Sort     string `json:"sort"`
	Filter   string `json:"filter"`
	Sequence int64  `json:"sequence"`
	Ceiling  int64  `json:"ceiling"`
}

func workItemFilterKey(status, assignee, nodeID string) string {
	data, _ := json.Marshal([]string{status, assignee, nodeID})
	return string(data)
}

func encodeWorkItemCursor(cursor workItemCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeWorkItemCursor(raw string, cursor *workItemCursor) error {
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(data) > 1024 {
		return fmt.Errorf("invalid cursor")
	}
	if err := json.Unmarshal(data, cursor); err != nil {
		return err
	}
	if cursor.Sequence < 1 || cursor.Ceiling < cursor.Sequence {
		return fmt.Errorf("invalid cursor sequence")
	}
	return nil
}

func validationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message, "error_code": "validation_error"})
}

func validWorkItemInput(req models.WorkItemRequest, creating bool) error {
	if !workItemTypes[req.Type] {
		return fmt.Errorf("invalid work item type")
	}
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if req.Priority != "" && !workItemPriorities[req.Priority] {
		return fmt.Errorf("invalid priority")
	}
	if req.Points != nil && !allowedPoints[*req.Points] {
		return fmt.Errorf("points must be one of 1, 2, 3, 5, 8, or 13")
	}
	if req.Status != "" && !workItemStatuses[req.Status] {
		return fmt.Errorf("invalid status")
	}
	status := req.Status
	if status == "" && creating {
		status = "Backlog"
	}
	if status == "Blocked" && (req.BlockedReason == nil || strings.TrimSpace(*req.BlockedReason) == "") {
		return fmt.Errorf("blocked status requires blocked_reason")
	}
	if status == "Done" && (req.Resolution == nil || strings.TrimSpace(*req.Resolution) == "") {
		return fmt.Errorf("done status requires resolution")
	}
	if req.NodeID == nil && (req.Description == nil || strings.TrimSpace(*req.Description) == "") {
		return fmt.Errorf("project-level work item requires description as project scope")
	}
	for _, date := range []*string{req.StartDate, req.DueDate} {
		if date != nil && *date != "" {
			if _, err := time.Parse("2006-01-02", *date); err != nil {
				return fmt.Errorf("dates must use YYYY-MM-DD")
			}
		}
	}
	return nil
}

func statusTransitionAllowed(from, to string) bool {
	if from == to {
		return true
	}
	return map[string]map[string]bool{
		"Backlog": {"Ready": true, "Canceled": true}, "Ready": {"In Progress": true, "Blocked": true, "Canceled": true},
		"In Progress": {"In Review": true, "Blocked": true, "Ready": true, "Canceled": true},
		"In Review":   {"Done": true, "In Progress": true, "Blocked": true, "Canceled": true},
		"Blocked":     {"Ready": true, "In Progress": true, "Canceled": true}, "Done": {}, "Canceled": {},
	}[from][to]
}

func workItemProjectByKey(key string) (string, error) {
	var projectID string
	err := db.DB.QueryRow(`SELECT project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, key).Scan(&projectID)
	return projectID, err
}

func scanWorkItem(row *sql.Row) (models.WorkItem, error) {
	var item models.WorkItem
	err := row.Scan(&item.ID, &item.Key, &item.ProjectID, &item.ModuleID, &item.NodeID, &item.FacetKey, &item.ParentID, &item.Type, &item.Title, &item.Description, &item.Priority, &item.Points, &item.Status, &item.AssigneeID, &item.ReporterID, &item.StartDate, &item.DueDate, &item.BlockedReason, &item.Resolution, &item.RowVersion, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

const workItemFields = `id, work_key, project_id, module_id, node_id, facet_key, parent_id, type, title, description, priority, points, status, assignee_id, reporter_id, start_date, due_date, blocked_reason, resolution, row_version, created_at, updated_at`

// PostgreSQL prepared statements must receive one stable type for each
// placeholder. $1 participates in both the status assignment and CASE checks.
const transitionWorkItemSQL = `UPDATE work_items SET status=$1::varchar,blocked_reason=CASE WHEN $1::varchar='Blocked' THEN $2 ELSE blocked_reason END,resolution=CASE WHEN $1::varchar='Done' THEN $3 ELSE resolution END,row_version=row_version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$4 AND row_version=$5 AND deleted_at IS NULL`

func getWorkItem(key string) (models.WorkItem, error) {
	return scanWorkItem(db.DB.QueryRow(`SELECT `+workItemFields+` FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, key))
}

func workItemDeliveryChanges(before models.WorkItem, after models.WorkItemRequest) (map[string]any, map[string]any) {
	oldValues, newValues := map[string]any{}, map[string]any{}
	compare := func(field string, oldValue, newValue any) {
		oldJSON, _ := json.Marshal(oldValue)
		newJSON, _ := json.Marshal(newValue)
		if string(oldJSON) != string(newJSON) {
			oldValues[field], newValues[field] = oldValue, newValue
		}
	}
	compare("assignee_id", before.AssigneeID, after.AssigneeID)
	compare("points", before.Points, after.Points)
	compare("parent_id", before.ParentID, after.ParentID)
	var dueDate *string
	if before.DueDate != nil {
		value := before.DueDate.Format("2006-01-02")
		dueDate = &value
	}
	compare("due_date", dueDate, after.DueDate)
	return oldValues, newValues
}

// CreateWorkItemHandler delegates to the Clean Architecture work item creation handler.
func CreateWorkItemHandler(c *gin.Context) {
	workitemHandler.HandleCreate(c)
}

// ListWorkItemsHandler delegates to the Clean Architecture work items list handler.
func ListWorkItemsHandler(c *gin.Context) {
	workitemHandler.HandleList(c)
}

// GetWorkItemHandler delegates to the Clean Architecture work item detail handler.
func GetWorkItemHandler(c *gin.Context) {
	workitemHandler.HandleGetDetail(c)
}

// UpdateWorkItemHandler delegates to the Clean Architecture work item update handler.
func UpdateWorkItemHandler(c *gin.Context) {
	workitemHandler.HandleUpdate(c)
}

// TransitionWorkItemHandler delegates to the Clean Architecture work item transition handler.
func TransitionWorkItemHandler(c *gin.Context) {
	workitemHandler.HandleTransition(c)
}


func createComment(c *gin.Context, projectID, nodeID, workItemID string) {
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityComment) {
		return
	}
	userID, _ := middleware.GetUserID(c)
	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		validationError(c, "comment body is required")
		return
	}
	mentions, _ := json.Marshal(req.Mentions)
	if req.ParentID != nil {
		var sameTarget bool
		err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM comments WHERE id=$1 AND project_id=$2 AND deleted_at IS NULL AND (($3<>'' AND node_id=$3) OR ($4<>'' AND work_item_id=$4)))`, *req.ParentID, projectID, nodeID, workItemID).Scan(&sameTarget)
		if err != nil || !sameTarget {
			validationError(c, "parent comment must target the same resource")
			return
		}
	}
	id := "com_" + GenerateUUID()
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start comment transaction"})
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO comments(id,project_id,module_id,node_id,work_item_id,parent_id,author_id,body,mentions) VALUES($1,$2,NULL,$3,NULLIF($4,''),$5,$6,$7,$8)`, id, projectID, nilIfEmpty(nodeID), workItemID, req.ParentID, userID, strings.TrimSpace(req.Body), string(mentions))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create comment"})
		return
	}
	if err := CreateMentionNotifications(tx, projectID, id, userID, req.Mentions); err != nil {
		c.JSON(500, gin.H{"error": "failed to create mention notifications"})
		return
	}
	event, err := writeCollaborationEvent(tx, projectID, "", userID, "comment.created", "comment:"+id, map[string]any{"comment_id": id, "work_item_id": workItemID, "node_id": nodeID})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record comment event"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit comment"})
		return
	}
	emitProjectEvent(projectID, event)
	c.JSON(201, gin.H{"id": id, "body": strings.TrimSpace(req.Body)})
}
func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func CreateNodeCommentHandler(c *gin.Context) {
	nodeID := c.Param("id")
	var projectID string
	err := db.DB.QueryRow(`SELECT m.project_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL`, nodeID).Scan(&projectID)
	if err != nil {
		c.JSON(404, gin.H{"error": "node not found"})
		return
	}
	createComment(c, projectID, nodeID, "")
}
func CreateWorkItemCommentHandler(c *gin.Context) {
	projectID, err := workItemProjectByKey(c.Param("key"))
	if err != nil {
		c.JSON(404, gin.H{"error": "work item not found"})
		return
	}
	item, _ := getWorkItem(c.Param("key"))
	createComment(c, projectID, "", item.ID)
}

func ListWorkItemCommentsHandler(c *gin.Context) {
	item, err := getWorkItem(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}
	if !middleware.AuthorizeProject(c, item.ProjectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,parent_id,author_id,body,mentions,resolved_at,created_at,updated_at FROM comments WHERE work_item_id=$1 AND deleted_at IS NULL ORDER BY created_at`, item.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}
	defer rows.Close()
	comments := []gin.H{}
	for rows.Next() {
		var id, author, body, mentions string
		var parent *string
		var resolved *time.Time
		var created, updated time.Time
		if err := rows.Scan(&id, &parent, &author, &body, &mentions, &resolved, &created, &updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse comments"})
			return
		}
		comments = append(comments, gin.H{"id": id, "parent_id": parent, "author_id": author, "body": body, "mentions": json.RawMessage(mentions), "resolved_at": resolved, "created_at": created, "updated_at": updated})
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list comments"})
		return
	}
	c.JSON(http.StatusOK, comments)
}
func ListNodeCommentsHandler(c *gin.Context) {
	nodeID := c.Param("id")
	var projectID string
	err := db.DB.QueryRow(`SELECT m.project_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL`, nodeID).Scan(&projectID)
	if err != nil {
		c.JSON(404, gin.H{"error": "node not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,parent_id,author_id,body,mentions,resolved_at,created_at,updated_at FROM comments WHERE node_id=$1 AND deleted_at IS NULL ORDER BY created_at`, nodeID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list comments"})
		return
	}
	defer rows.Close()
	comments := []gin.H{}
	for rows.Next() {
		var id, author, body, mentions string
		var parent *string
		var resolved *time.Time
		var created, updated time.Time
		if err := rows.Scan(&id, &parent, &author, &body, &mentions, &resolved, &created, &updated); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse comments"})
			return
		}
		comments = append(comments, gin.H{"id": id, "parent_id": parent, "author_id": author, "body": body, "mentions": json.RawMessage(mentions), "resolved_at": resolved, "created_at": created, "updated_at": updated})
	}
	c.JSON(200, comments)
}

// UpdateCommentHandler permits an author to edit or resolve their own comment.
func UpdateCommentHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var projectID, authorID string
	err = db.DB.QueryRow(`SELECT project_id, author_id FROM comments WHERE id=$1 AND deleted_at IS NULL`, c.Param("id")).Scan(&projectID, &authorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityComment) {
		return
	}
	if authorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the comment author can edit or resolve it"})
		return
	}
	var req models.CommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Body == nil && req.Resolved == nil) {
		validationError(c, "body or resolved is required")
		return
	}
	if req.Body != nil && strings.TrimSpace(*req.Body) == "" {
		validationError(c, "comment body cannot be empty")
		return
	}
	_, err = db.DB.Exec(`UPDATE comments SET body=COALESCE($1,body), resolved_at=CASE WHEN $2::boolean THEN COALESCE(resolved_at,CURRENT_TIMESTAMP) WHEN $2::boolean IS FALSE THEN NULL ELSE resolved_at END, updated_at=CURRENT_TIMESTAMP WHERE id=$3`, req.Body, req.Resolved, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update comment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
