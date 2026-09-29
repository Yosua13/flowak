package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"backend/db"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

type checklistEntry struct {
	ID         string    `json:"id"`
	Body       string    `json:"body"`
	IsComplete bool      `json:"is_complete"`
	SortOrder  int       `json:"sort_order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type watcherEntry struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
type linkEntry struct {
	ID               string    `json:"id"`
	LinkedWorkItemID string    `json:"linked_work_item_id"`
	LinkedKey        string    `json:"linked_key"`
	LinkType         string    `json:"link_type"`
	CreatedAt        time.Time `json:"created_at"`
}
type attachmentEntry struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType *string   `json:"content_type,omitempty"`
	StorageKey  string    `json:"storage_key"`
	SizeBytes   *int64    `json:"size_bytes,omitempty"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
}
type artifactSummary struct {
	Checklist       []checklistEntry  `json:"checklist"`
	Watchers        []watcherEntry    `json:"watchers"`
	Links           []linkEntry       `json:"links"`
	Attachments     []attachmentEntry `json:"attachments"`
	CommentCount    int               `json:"comment_count"`
	AttachmentCount int               `json:"attachment_count"`
}

func authorizedArtifactItem(c *gin.Context, capability middleware.ProjectCapability) (string, string, bool) {
	var id, projectID string
	err := db.DB.QueryRow(`SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, c.Param("key")).Scan(&id, &projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return "", "", false
	}
	if !middleware.AuthorizeProject(c, projectID, capability) {
		return "", "", false
	}
	return id, projectID, true
}

// GetWorkItemArtifactsHandler returns one tenant-scoped snapshot for the detail modal.
func GetWorkItemArtifactsHandler(c *gin.Context) {
	itemID, _, ok := authorizedArtifactItem(c, middleware.CapabilityView)
	if !ok {
		return
	}
	summary := artifactSummary{Checklist: []checklistEntry{}, Watchers: []watcherEntry{}, Links: []linkEntry{}, Attachments: []attachmentEntry{}}
	rows, err := db.DB.Query(`SELECT id,body,is_complete,sort_order,created_at,updated_at FROM work_item_acceptance_criteria WHERE work_item_id=$1 ORDER BY sort_order,id`, itemID)
	if err != nil {
		artifactServerError(c)
		return
	}
	for rows.Next() {
		var entry checklistEntry
		if rows.Scan(&entry.ID, &entry.Body, &entry.IsComplete, &entry.SortOrder, &entry.CreatedAt, &entry.UpdatedAt) != nil {
			rows.Close()
			artifactServerError(c)
			return
		}
		summary.Checklist = append(summary.Checklist, entry)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		artifactServerError(c)
		return
	}
	rows, err = db.DB.Query(`SELECT w.user_id,u.name,w.created_at FROM work_item_watchers w JOIN users u ON u.id=w.user_id WHERE w.work_item_id=$1 ORDER BY w.created_at,w.user_id`, itemID)
	if err != nil {
		artifactServerError(c)
		return
	}
	for rows.Next() {
		var entry watcherEntry
		if rows.Scan(&entry.UserID, &entry.Name, &entry.CreatedAt) != nil {
			rows.Close()
			artifactServerError(c)
			return
		}
		summary.Watchers = append(summary.Watchers, entry)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		artifactServerError(c)
		return
	}
	rows, err = db.DB.Query(`SELECT l.id,l.linked_work_item_id,w.work_key,l.link_type,l.created_at FROM work_item_links l JOIN work_items w ON w.id=l.linked_work_item_id AND w.deleted_at IS NULL WHERE l.work_item_id=$1 ORDER BY l.created_at,l.id`, itemID)
	if err != nil {
		artifactServerError(c)
		return
	}
	for rows.Next() {
		var entry linkEntry
		if rows.Scan(&entry.ID, &entry.LinkedWorkItemID, &entry.LinkedKey, &entry.LinkType, &entry.CreatedAt) != nil {
			rows.Close()
			artifactServerError(c)
			return
		}
		summary.Links = append(summary.Links, entry)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		artifactServerError(c)
		return
	}
	rows, err = db.DB.Query(`SELECT id,file_name,content_type,storage_key,size_bytes,uploaded_by,created_at FROM attachments WHERE work_item_id=$1 ORDER BY created_at DESC,id DESC`, itemID)
	if err != nil {
		artifactServerError(c)
		return
	}
	for rows.Next() {
		var entry attachmentEntry
		if rows.Scan(&entry.ID, &entry.FileName, &entry.ContentType, &entry.StorageKey, &entry.SizeBytes, &entry.UploadedBy, &entry.CreatedAt) != nil {
			rows.Close()
			artifactServerError(c)
			return
		}
		summary.Attachments = append(summary.Attachments, entry)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		artifactServerError(c)
		return
	}
	err = db.DB.QueryRow(`SELECT (SELECT count(*) FROM comments WHERE work_item_id=$1 AND deleted_at IS NULL),(SELECT count(*) FROM attachments WHERE work_item_id=$1)`, itemID).Scan(&summary.CommentCount, &summary.AttachmentCount)
	if err != nil {
		artifactServerError(c)
		return
	}
	c.JSON(http.StatusOK, summary)
}

type artifactInput struct {
	Body        string `json:"body"`
	IsComplete  *bool  `json:"is_complete"`
	SortOrder   *int   `json:"sort_order"`
	UserID      string `json:"user_id"`
	LinkedKey   string `json:"linked_key"`
	LinkType    string `json:"link_type"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	StorageKey  string `json:"storage_key"`
	SizeBytes   *int64 `json:"size_bytes"`
}

func validArtifactInput(kind, method string, input artifactInput) error {
	switch kind {
	case "checklist":
		if method == http.MethodPost && strings.TrimSpace(input.Body) == "" {
			return errors.New("checklist body is required")
		}
		if method == http.MethodPatch && strings.TrimSpace(input.Body) == "" && input.IsComplete == nil && input.SortOrder == nil {
			return errors.New("checklist change is required")
		}
		if len(input.Body) > 4000 {
			return errors.New("checklist body is too long")
		}
	case "watchers":
		if method != http.MethodPost || input.UserID == "" {
			return errors.New("watcher user_id is required")
		}
	case "links":
		if method == http.MethodPost && strings.TrimSpace(input.LinkedKey) == "" {
			return errors.New("linked_key is required")
		}
		if method == http.MethodPatch && input.LinkType == "" {
			return errors.New("link_type is required")
		}
		if input.LinkType != "" && input.LinkType != "relates_to" && input.LinkType != "blocks" && input.LinkType != "depends_on" {
			return errors.New("invalid link_type")
		}
	case "attachments":
		if method == http.MethodPost && (strings.TrimSpace(input.FileName) == "" || strings.TrimSpace(input.StorageKey) == "") {
			return errors.New("file_name and storage_key are required")
		}
		if method == http.MethodPatch && strings.TrimSpace(input.FileName) == "" && input.SizeBytes == nil && input.ContentType == "" {
			return errors.New("attachment change is required")
		}
		if len(input.FileName) > 255 || len(input.ContentType) > 120 {
			return errors.New("attachment metadata is too long")
		}
		if input.SizeBytes != nil && *input.SizeBytes < 0 {
			return errors.New("size_bytes must be non-negative")
		}
		// With no storage provider, the metadata points only to a pre-existing HTTPS artifact.
		if method == http.MethodPost {
			u, err := url.Parse(input.StorageKey)
			if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
				return errors.New("storage_key must be an HTTPS artifact URL without credentials or query")
			}
		}
	default:
		return errors.New("unknown artifact type")
	}
	return nil
}

func artifactServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process work item artifact"})
}

func writeArtifactActivity(tx *sql.Tx, itemID, projectID, actorID, kind, action, artifactID string) error {
	data, _ := json.Marshal(gin.H{"kind": kind, "artifact_id": artifactID})
	_, err := tx.Exec(`INSERT INTO activity_logs(id,project_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,$3,$4,'work_item',$5,$6::jsonb)`, "act_"+GenerateUUID(), projectID, actorID, action, itemID, string(data))
	return err
}

// MutateWorkItemArtifactHandler handles create, update and delete with one authorization and audit path.
func MutateWorkItemArtifactHandler(c *gin.Context) {
	itemID, projectID, ok := authorizedArtifactItem(c, middleware.CapabilityWorkItem)
	if !ok {
		return
	}
	kind, method := c.Param("kind"), c.Request.Method
	if kind != "checklist" && kind != "watchers" && kind != "links" && kind != "attachments" {
		c.JSON(404, gin.H{"error": "artifact type not found"})
		return
	}
	if method == http.MethodPatch && kind != "checklist" && kind != "attachments" && kind != "links" {
		c.JSON(405, gin.H{"error": "artifact cannot be patched"})
		return
	}
	var input artifactInput
	if method != http.MethodDelete {
		if c.ShouldBindJSON(&input) != nil {
			validationError(c, "invalid artifact input")
			return
		}
		if err := validArtifactInput(kind, method, input); err != nil {
			validationError(c, err.Error())
			return
		}
	}
	actorID, _ := middleware.GetUserID(c)
	tx, err := db.DB.Begin()
	if err != nil {
		artifactServerError(c)
		return
	}
	defer tx.Rollback()
	id := c.Param("artifactId")
	if method == http.MethodPost {
		id = "artifact_" + GenerateUUID()
	}
	var result sql.Result
	switch kind {
	case "checklist":
		switch method {
		case http.MethodPost:
			result, err = tx.Exec(`INSERT INTO work_item_acceptance_criteria(id,work_item_id,body,sort_order) VALUES($1,$2,$3,$4)`, id, itemID, strings.TrimSpace(input.Body), intValue(input.SortOrder))
		case http.MethodPatch:
			result, err = tx.Exec(`UPDATE work_item_acceptance_criteria SET body=COALESCE(NULLIF($1::text,''),body),is_complete=COALESCE($2::boolean,is_complete),sort_order=COALESCE($3::int,sort_order),updated_at=CURRENT_TIMESTAMP WHERE id=$4 AND work_item_id=$5`, strings.TrimSpace(input.Body), input.IsComplete, input.SortOrder, id, itemID)
		case http.MethodDelete:
			result, err = tx.Exec(`DELETE FROM work_item_acceptance_criteria WHERE id=$1 AND work_item_id=$2`, id, itemID)
		}
	case "watchers":
		if method == http.MethodPost {
			var member bool
			err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM projects p WHERE p.id=$1 AND p.owner_id=$2 UNION SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`, projectID, input.UserID).Scan(&member)
			if err != nil || !member {
				validationError(c, "watcher must be a project member")
				return
			}
			id = input.UserID
			result, err = tx.Exec(`INSERT INTO work_item_watchers(work_item_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, itemID, id)
		} else {
			result, err = tx.Exec(`DELETE FROM work_item_watchers WHERE work_item_id=$1 AND user_id=$2`, itemID, id)
		}
	case "links":
		if method == http.MethodPost {
			var linkedID string
			err = tx.QueryRow(`SELECT id FROM work_items WHERE work_key=$1 AND project_id=$2 AND deleted_at IS NULL`, input.LinkedKey, projectID).Scan(&linkedID)
			if err != nil || linkedID == itemID {
				validationError(c, "linked item must be distinct and in the same project")
				return
			}
			linkType := input.LinkType
			if linkType == "" {
				linkType = "relates_to"
			}
			result, err = tx.Exec(`INSERT INTO work_item_links(id,work_item_id,linked_work_item_id,link_type) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, id, itemID, linkedID, linkType)
		} else if method == http.MethodPatch {
			result, err = tx.Exec(`UPDATE work_item_links SET link_type=$1 WHERE id=$2 AND work_item_id=$3`, input.LinkType, id, itemID)
		} else {
			result, err = tx.Exec(`DELETE FROM work_item_links WHERE id=$1 AND work_item_id=$2`, id, itemID)
		}
	case "attachments":
		switch method {
		case http.MethodPost:
			result, err = tx.Exec(`INSERT INTO attachments(id,project_id,work_item_id,file_name,content_type,storage_key,size_bytes,uploaded_by) VALUES($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8)`, id, projectID, itemID, strings.TrimSpace(input.FileName), input.ContentType, input.StorageKey, input.SizeBytes, actorID)
		case http.MethodPatch:
			result, err = tx.Exec(`UPDATE attachments SET file_name=COALESCE(NULLIF($1::text,''),file_name),content_type=COALESCE(NULLIF($2::text,''),content_type),size_bytes=COALESCE($3::bigint,size_bytes) WHERE id=$4 AND work_item_id=$5 AND project_id=$6`, strings.TrimSpace(input.FileName), input.ContentType, input.SizeBytes, id, itemID, projectID)
		case http.MethodDelete:
			result, err = tx.Exec(`DELETE FROM attachments WHERE id=$1 AND work_item_id=$2 AND project_id=$3`, id, itemID, projectID)
		}
	}
	if err != nil {
		artifactServerError(c)
		return
	}
	changed, err := result.RowsAffected()
	if err != nil {
		artifactServerError(c)
		return
	}
	if changed == 0 && method != http.MethodPost {
		c.JSON(404, gin.H{"error": "artifact not found"})
		return
	}
	if changed == 0 && kind == "links" {
		c.JSON(409, gin.H{"error": "link already exists"})
		return
	}
	if changed > 0 {
		action := map[string]string{http.MethodPost: "artifact_created", http.MethodPatch: "artifact_updated", http.MethodDelete: "artifact_deleted"}[method]
		if writeArtifactActivity(tx, itemID, projectID, actorID, kind, action, id) != nil {
			artifactServerError(c)
			return
		}
	}
	if tx.Commit() != nil {
		artifactServerError(c)
		return
	}
	if method == http.MethodDelete {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(map[string]int{http.MethodPost: http.StatusCreated, http.MethodPatch: http.StatusOK}[method], gin.H{"id": id})
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
