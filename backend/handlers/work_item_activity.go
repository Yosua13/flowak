package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"backend/db"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// GetWorkItemActivityHandler exposes a safe, tenant-scoped activity timeline
// for the Kanban detail modal. Change payloads remain server-side only.
func GetWorkItemActivityHandler(c *gin.Context) {
	cursor, ok := decodeTimelineCursor(c)
	if !ok {
		return
	}
	limit, ok := timelineLimit(c)
	if !ok {
		return
	}
	var itemID, projectID string
	if err := db.DB.QueryRow(`SELECT id,project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL`, c.Param("key")).Scan(&itemID, &projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,action,actor_id,created_at FROM activity_logs WHERE project_id=$1 AND entity_type='work_item' AND entity_id=$2 AND ($3::timestamp IS NULL OR (created_at,id) < ($3,$4)) ORDER BY created_at DESC,id DESC LIMIT $5`, projectID, itemID, nullableCursorTime(cursor), nullableCursorID(cursor), limit+1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list work item activity"})
		return
	}
	defer rows.Close()
	entries := []gin.H{}
	for rows.Next() {
		var id, action string
		var actorID sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &action, &actorID, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse work item activity"})
			return
		}
		entry := gin.H{"id": id, "action": action, "created_at": createdAt}
		if actorID.Valid {
			entry["actor_id"] = actorID.String
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list work item activity"})
		return
	}
	next := ""
	if len(entries) > limit {
		last := entries[limit-1]
		next = encodeTimelineCursor(last["created_at"].(time.Time), last["id"].(string))
		entries = entries[:limit]
	}
	c.JSON(http.StatusOK, gin.H{"items": entries, "next_cursor": next})
}

func nullableCursorTime(cursor timelineCursor) any {
	if cursor.ID == "" {
		return nil
	}
	return cursor.CreatedAt
}
func nullableCursorID(cursor timelineCursor) any {
	if cursor.ID == "" {
		return nil
	}
	return cursor.ID
}
