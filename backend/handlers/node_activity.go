package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"backend/db"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// ListNodeActivityHandler returns server-recorded activity for one node. The
// response intentionally excludes before/after payloads because those may
// contain fields that are not appropriate to expose in a compact activity feed.
func ListNodeActivityHandler(c *gin.Context) {
	nodeID := c.Param("id")
	var projectID string
	if err := db.DB.QueryRow(`SELECT m.project_id FROM workflow_nodes n JOIN modules m ON m.id=n.module_id WHERE n.id=$1 AND n.deleted_at IS NULL`, nodeID).Scan(&projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}

	rows, err := db.DB.Query(`
		SELECT id, action, actor_id, created_at FROM (
			SELECT 'comment:' || id AS id, 'comment_created' AS action, author_id AS actor_id, created_at
			FROM comments
			WHERE node_id=$1 AND deleted_at IS NULL
			UNION ALL
			SELECT a.id, a.action, a.actor_id, a.created_at
			FROM activity_logs a
			JOIN work_items w ON w.id=a.entity_id AND w.deleted_at IS NULL
			WHERE a.project_id=$2 AND a.entity_type='work_item' AND w.node_id=$1
		) AS node_activity
		ORDER BY created_at DESC
		LIMIT 100`, nodeID, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list node activity"})
		return
	}
	defer rows.Close()

	activity := []gin.H{}
	for rows.Next() {
		var id, action string
		var actorID sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &action, &actorID, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse node activity"})
			return
		}
		entry := gin.H{"id": id, "action": action, "created_at": createdAt}
		if actorID.Valid {
			entry["actor_id"] = actorID.String
		}
		activity = append(activity, entry)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list node activity"})
		return
	}
	c.JSON(http.StatusOK, activity)
}
