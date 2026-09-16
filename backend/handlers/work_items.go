package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
)

var workItemTypes = map[string]bool{"Story": true, "Task": true, "Bug": true, "Review": true, "Research": true, "Subtask": true}
var workItemStatuses = map[string]bool{"Backlog": true, "Ready": true, "In Progress": true, "In Review": true, "Blocked": true, "Done": true, "Canceled": true}
var workItemPriorities = map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
var allowedPoints = map[int]bool{1: true, 2: true, 3: true, 5: true, 8: true, 13: true}

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

func getWorkItem(key string) (models.WorkItem, error) {
	return scanWorkItem(db.DB.QueryRow(`SELECT `+workItemFields+` FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, key))
}

func validateWorkItemReferences(tx *sql.Tx, projectID string, req *models.WorkItemRequest) error {
	if req.ModuleID != nil {
		var exists bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM modules WHERE id=$1 AND project_id=$2 AND status='active')`, *req.ModuleID, projectID).Scan(&exists); err != nil || !exists {
			return fmt.Errorf("module does not belong to project")
		}
	}
	if req.NodeID != nil {
		var moduleID string
		if err := tx.QueryRow(`SELECT n.module_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL AND m.project_id=$2`, *req.NodeID, projectID).Scan(&moduleID); err != nil {
			return fmt.Errorf("node does not belong to project")
		}
		if req.ModuleID != nil && *req.ModuleID != moduleID {
			return fmt.Errorf("node does not belong to module")
		}
		req.ModuleID = &moduleID
	}
	if req.ParentID != nil {
		var parentProject string
		if err := tx.QueryRow(`SELECT project_id FROM work_items WHERE id=$1 AND deleted_at IS NULL`, *req.ParentID).Scan(&parentProject); err != nil || parentProject != projectID {
			return fmt.Errorf("parent must belong to the same project")
		}
	}
	if req.AssigneeID != nil {
		var member bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`, projectID, *req.AssigneeID).Scan(&member); err != nil || !member {
			return fmt.Errorf("assignee must be a project member")
		}
	}
	return nil
}

func CreateWorkItemHandler(c *gin.Context) {
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
		validationError(c, "invalid request body")
		return
	}
	if err := validWorkItemInput(req, true); err != nil {
		validationError(c, err.Error())
		return
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}
	if req.Status == "" {
		req.Status = "Backlog"
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()
	if err := validateWorkItemReferences(tx, projectID, &req); err != nil {
		validationError(c, err.Error())
		return
	}
	var sequence int64
	if err := tx.QueryRow(`INSERT INTO project_work_item_sequences(project_id,next_value) VALUES($1,2) ON CONFLICT(project_id) DO UPDATE SET next_value=project_work_item_sequences.next_value+1 RETURNING next_value-1`, projectID).Scan(&sequence); err != nil {
		c.JSON(500, gin.H{"error": "failed to allocate work item key"})
		return
	}
	var prefix string
	if err := tx.QueryRow(`SELECT work_item_prefix FROM projects WHERE id=$1`, projectID).Scan(&prefix); err != nil {
		c.JSON(500, gin.H{"error": "project not found"})
		return
	}
	id, key := "wi_"+GenerateUUID(), fmt.Sprintf("%s-%d", prefix, sequence)
	_, err = tx.Exec(`INSERT INTO work_items(id,work_key,project_id,module_id,node_id,facet_key,parent_id,sequence,type,title,description,priority,points,status,assignee_id,reporter_id,start_date,due_date,blocked_reason,resolution) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NULLIF($17,'')::date,NULLIF($18,'')::date,$19,$20)`, id, key, projectID, req.ModuleID, req.NodeID, req.FacetKey, req.ParentID, sequence, req.Type, strings.TrimSpace(req.Title), req.Description, req.Priority, req.Points, req.Status, req.AssigneeID, userID, req.StartDate, req.DueDate, req.BlockedReason, req.Resolution)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create work item"})
		return
	}
	_, err = tx.Exec(`INSERT INTO work_item_status_history(id,work_item_id,to_status,changed_by,note) VALUES($1,$2,$3,$4,'created')`, "wih_"+GenerateUUID(), id, req.Status, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record work item history"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit work item"})
		return
	}
	item, err := getWorkItem(key)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read created work item"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func ListWorkItemsHandler(c *gin.Context) {
	projectID := c.Param("id")
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	status, assignee, nodeID := c.Query("status"), c.Query("assignee_id"), c.Query("node_id")
	if status != "" && !workItemStatuses[status] {
		validationError(c, "invalid status filter")
		return
	}
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			validationError(c, "limit must be 1-100")
			return
		}
		limit = n
	}
	args := []any{projectID}
	where := "project_id=$1 AND deleted_at IS NULL"
	for _, filter := range []struct{ value, column string }{{status, "status"}, {assignee, "assignee_id"}, {nodeID, "node_id"}} {
		if filter.value != "" {
			args = append(args, filter.value)
			where += fmt.Sprintf(" AND %s=$%d", filter.column, len(args))
		}
	}
	if cursor := c.Query("cursor"); cursor != "" {
		n, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			validationError(c, "invalid cursor")
			return
		}
		args = append(args, n)
		where += fmt.Sprintf(" AND sequence < $%d", len(args))
	}
	args = append(args, limit)
	rows, err := db.DB.Query(`SELECT `+workItemFields+` FROM work_items WHERE `+where+` ORDER BY sequence DESC LIMIT $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list work items"})
		return
	}
	defer rows.Close()
	items := []models.WorkItem{}
	for rows.Next() {
		var i models.WorkItem
		if err := rows.Scan(&i.ID, &i.Key, &i.ProjectID, &i.ModuleID, &i.NodeID, &i.FacetKey, &i.ParentID, &i.Type, &i.Title, &i.Description, &i.Priority, &i.Points, &i.Status, &i.AssigneeID, &i.ReporterID, &i.StartDate, &i.DueDate, &i.BlockedReason, &i.Resolution, &i.RowVersion, &i.CreatedAt, &i.UpdatedAt); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse work item"})
			return
		}
		items = append(items, i)
	}
	c.JSON(200, items)
}

func GetWorkItemHandler(c *gin.Context) {
	projectID, err := workItemProjectByKey(c.Param("key"))
	if err != nil {
		c.JSON(404, gin.H{"error": "work item not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	item, err := getWorkItem(c.Param("key"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read work item"})
		return
	}
	c.JSON(200, item)
}

func UpdateWorkItemHandler(c *gin.Context) {
	key := c.Param("key")
	item, err := getWorkItem(key)
	if err != nil {
		c.JSON(404, gin.H{"error": "work item not found"})
		return
	}
	if !middleware.AuthorizeProject(c, item.ProjectID, middleware.CapabilityWorkItem) {
		return
	}
	var req models.WorkItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "invalid request body")
		return
	}
	if req.RowVersion < 1 {
		validationError(c, "row_version is required")
		return
	}
	if req.Status != "" && req.Status != item.Status {
		validationError(c, "use the transition endpoint to change status")
		return
	}
	// Patch fields inherit existing values so all status invariants are validated consistently.
	if req.Type == "" {
		req.Type = item.Type
	}
	if req.Title == "" {
		req.Title = item.Title
	}
	if req.Priority == "" {
		req.Priority = item.Priority
	}
	if req.Status == "" {
		req.Status = item.Status
	}
	if req.ModuleID == nil {
		req.ModuleID = item.ModuleID
	}
	if req.NodeID == nil {
		req.NodeID = item.NodeID
	}
	if req.ParentID == nil {
		req.ParentID = item.ParentID
	}
	if req.Description == nil {
		req.Description = item.Description
	}
	if req.Points == nil {
		req.Points = item.Points
	}
	if req.AssigneeID == nil {
		req.AssigneeID = item.AssigneeID
	}
	if req.BlockedReason == nil {
		req.BlockedReason = item.BlockedReason
	}
	if req.Resolution == nil {
		req.Resolution = item.Resolution
	}
	if req.StartDate == nil && item.StartDate != nil {
		value := item.StartDate.Format("2006-01-02")
		req.StartDate = &value
	}
	if req.DueDate == nil && item.DueDate != nil {
		value := item.DueDate.Format("2006-01-02")
		req.DueDate = &value
	}
	if err := validWorkItemInput(req, false); err != nil {
		validationError(c, err.Error())
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()
	if err := validateWorkItemReferences(tx, item.ProjectID, &req); err != nil {
		validationError(c, err.Error())
		return
	}
	if req.ParentID != nil && *req.ParentID == item.ID {
		validationError(c, "work item cannot be its own parent")
		return
	}
	if req.ParentID != nil {
		var cycle bool
		err = tx.QueryRow(`WITH RECURSIVE ancestors AS (SELECT parent_id FROM work_items WHERE id=$1 UNION ALL SELECT w.parent_id FROM work_items w JOIN ancestors a ON w.id=a.parent_id WHERE a.parent_id IS NOT NULL) SELECT EXISTS(SELECT 1 FROM ancestors WHERE parent_id=$2)`, *req.ParentID, item.ID).Scan(&cycle)
		if err != nil || cycle {
			validationError(c, "parent would create a hierarchy cycle")
			return
		}
	}
	result, err := tx.Exec(`UPDATE work_items SET module_id=$1,node_id=$2,facet_key=$3,parent_id=$4,type=$5,title=$6,description=$7,priority=$8,points=$9,status=$10,assignee_id=$11,start_date=NULLIF($12,'')::date,due_date=NULLIF($13,'')::date,blocked_reason=$14,resolution=$15,row_version=row_version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$16 AND row_version=$17 AND deleted_at IS NULL`, req.ModuleID, req.NodeID, req.FacetKey, req.ParentID, req.Type, strings.TrimSpace(req.Title), req.Description, req.Priority, req.Points, req.Status, req.AssigneeID, req.StartDate, req.DueDate, req.BlockedReason, req.Resolution, item.ID, req.RowVersion)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update work item"})
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		c.JSON(409, gin.H{"error": "work item has changed", "error_code": "work_item_version_conflict"})
		return
	}
	_, err = tx.Exec(`INSERT INTO activity_logs(id,project_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,'updated','work_item',$4,jsonb_build_object('row_version',$5))`, "act_"+GenerateUUID(), item.ProjectID, func() string { id, _ := middleware.GetUserID(c); return id }(), item.ID, req.RowVersion+1)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record work item activity"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit work item"})
		return
	}
	updated, _ := getWorkItem(key)
	c.JSON(200, updated)
}

func TransitionWorkItemHandler(c *gin.Context) {
	key := c.Param("key")
	item, err := getWorkItem(key)
	if err != nil {
		c.JSON(404, gin.H{"error": "work item not found"})
		return
	}
	if !middleware.AuthorizeProject(c, item.ProjectID, middleware.CapabilityWorkItem) {
		return
	}
	userID, _ := middleware.GetUserID(c)
	var req models.WorkItemTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "invalid request body")
		return
	}
	if req.RowVersion < 1 || !workItemStatuses[req.Status] {
		validationError(c, "valid status and row_version are required")
		return
	}
	if !statusTransitionAllowed(item.Status, req.Status) {
		validationError(c, "invalid status transition")
		return
	}
	if req.Status == "Done" && strings.TrimSpace(req.Resolution) == "" {
		validationError(c, "done status requires resolution")
		return
	}
	if req.Status == "Blocked" && strings.TrimSpace(req.Note) == "" {
		validationError(c, "blocked status requires a reason in note")
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE work_items SET status=$1,blocked_reason=CASE WHEN $1='Blocked' THEN $2 ELSE blocked_reason END,resolution=CASE WHEN $1='Done' THEN $3 ELSE resolution END,row_version=row_version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$4 AND row_version=$5 AND deleted_at IS NULL`, req.Status, req.Note, req.Resolution, item.ID, req.RowVersion)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to transition work item"})
		return
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		c.JSON(409, gin.H{"error": "work item has changed", "error_code": "work_item_version_conflict"})
		return
	}
	_, err = tx.Exec(`INSERT INTO work_item_status_history(id,work_item_id,from_status,to_status,changed_by,note) VALUES($1,$2,$3,$4,$5,$6)`, "wih_"+GenerateUUID(), item.ID, item.Status, req.Status, userID, req.Note)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record transition"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit transition"})
		return
	}
	updated, _ := getWorkItem(key)
	c.JSON(200, updated)
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
