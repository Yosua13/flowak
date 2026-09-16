package handlers

import (
	"backend/db"
	"backend/middleware"
	"backend/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"sync"
	"time"
)

const eventSchemaVersion = 1

var subscribers = struct {
	sync.Mutex
	byProject map[string]map[chan []byte]struct{}
}{byProject: map[string]map[chan []byte]struct{}{}}

func safeEventPayload(payload map[string]any) []byte {
	for _, key := range []string{"authorization", "cookie", "password", "secret", "token", "response", "body"} {
		delete(payload, key)
	}
	value, _ := json.Marshal(payload)
	return value
}
func emitProjectEvent(projectID string, event []byte) {
	subscribers.Lock()
	defer subscribers.Unlock()
	for ch := range subscribers.byProject[projectID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// writeCollaborationEvent writes the audit/outbox payload within the caller's transaction.
func writeCollaborationEvent(tx *sql.Tx, projectID, moduleID, actorID, name, key string, payload map[string]any) ([]byte, error) {
	raw := safeEventPayload(payload)
	eventID := "evt_" + GenerateUUID()
	_, err := tx.Exec(`INSERT INTO event_outbox(id,project_id,module_id,event_name,payload,idempotency_key) VALUES($1,$2,NULLIF($3,''),$4,$5::jsonb,$6) ON CONFLICT(idempotency_key) DO NOTHING`, eventID, projectID, moduleID, name, string(raw), key)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`INSERT INTO activity_logs(id,project_id,module_id,actor_id,action,entity_type,entity_id,after_data) VALUES($1,$2,NULLIF($3,''),$4,$5,'event',$6,$7::jsonb)`, "act_"+GenerateUUID(), projectID, moduleID, actorID, name, eventID, string(raw))
	if err != nil {
		return nil, err
	}
	envelope, _ := json.Marshal(gin.H{"schema_version": eventSchemaVersion, "id": eventID, "name": name, "project_id": projectID, "module_id": moduleID, "payload": json.RawMessage(raw)})
	return envelope, nil
}

func ListNotificationsHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	rows, err := db.DB.Query(`SELECT id,title,body,type,read_at,created_at,event_name FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read notifications"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, title, body, kind string
		var read *time.Time
		var created time.Time
		var event *string
		if err := rows.Scan(&id, &title, &body, &kind, &read, &created, &event); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse notifications"})
			return
		}
		out = append(out, gin.H{"id": id, "title": title, "message": body, "type": kind, "read": read != nil, "timestamp": created, "event_name": event})
	}
	c.JSON(200, out)
}
func MarkNotificationsReadHandler(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	_, err := db.DB.Exec(`UPDATE notifications SET read_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND read_at IS NULL`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update notifications"})
		return
	}
	c.JSON(200, gin.H{"success": true})
}
func ProjectEventsHandler(c *gin.Context) {
	projectID := c.Param("id")
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	ch := make(chan []byte, 16)
	subscribers.Lock()
	if subscribers.byProject[projectID] == nil {
		subscribers.byProject[projectID] = map[chan []byte]struct{}{}
	}
	subscribers.byProject[projectID][ch] = struct{}{}
	subscribers.Unlock()
	defer func() { subscribers.Lock(); delete(subscribers.byProject[projectID], ch); subscribers.Unlock() }()
	fmt.Fprint(c.Writer, "event: ready\ndata: {\"schema_version\":1}\n\n")
	c.Writer.Flush()
	select {
	case event := <-ch:
		fmt.Fprintf(c.Writer, "event: change\ndata: %s\n\n", event)
		c.Writer.Flush()
	case <-time.After(25 * time.Second):
		fmt.Fprint(c.Writer, "event: ping\ndata: {}\n\n")
		c.Writer.Flush()
	}
}

func CreateMentionNotifications(tx *sql.Tx, projectID, commentID, authorID string, mentions []string) error {
	for _, userID := range mentions {
		if userID == authorID {
			continue
		}
		var member bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id=$1 AND user_id=$2)`, projectID, userID).Scan(&member); err != nil || !member {
			continue
		}
		notificationID := "ntf_" + GenerateUUID()
		dedup := strings.Join([]string{"mention", commentID, userID}, ":")
		_, err := tx.Exec(`INSERT INTO notifications(id,user_id,project_id,title,body,type,event_name,payload) VALUES($1,$2,$3,'Anda disebut dalam komentar','Buka work item untuk meninjau komentar.','info','comment.mentioned',jsonb_build_object('comment_id',$4)) ON CONFLICT DO NOTHING`, notificationID, userID, projectID, commentID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO notification_outbox(id,notification_id,recipient_id,dedup_key) VALUES($1,$2,$3,$4) ON CONFLICT(dedup_key) DO NOTHING`, "nout_"+GenerateUUID(), notificationID, userID, dedup)
		if err != nil {
			return err
		}
	}
	return nil
}

func PublishModuleBaselineHandler(c *gin.Context) {
	moduleID := c.Param("id")
	var projectID string
	if err := db.DB.QueryRow(`SELECT project_id FROM modules WHERE id=$1 AND status='active'`, moduleID).Scan(&projectID); err != nil {
		c.JSON(404, gin.H{"error": "module not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}
	actorID, _ := middleware.GetUserID(c)
	module := models.Module{ID: moduleID}
	if err := hydrateModuleGraph(&module); err != nil {
		c.JSON(500, gin.H{"error": "failed to read normalized graph"})
		return
	}
	snapshot, _ := json.Marshal(gin.H{"schemaVersion": module.SchemaVersion, "nodes": json.RawMessage(module.Nodes), "edges": json.RawMessage(module.Edges)})
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start baseline transaction"})
		return
	}
	defer tx.Rollback()
	var version int
	if err = tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM module_versions WHERE module_id=$1`, moduleID).Scan(&version); err != nil {
		c.JSON(500, gin.H{"error": "failed to allocate baseline version"})
		return
	}
	versionID := "mv_" + GenerateUUID()
	if _, err = tx.Exec(`INSERT INTO module_versions(id,module_id,version,graph_snapshot,created_by,status,diff_summary,published_at) VALUES($1,$2,$3,$4::jsonb,$5,'published',jsonb_build_object('node_count',$6,'edge_count',$7),CURRENT_TIMESTAMP)`, versionID, moduleID, version, string(snapshot), actorID, len(json.RawMessage(module.Nodes)), len(json.RawMessage(module.Edges))); err != nil {
		c.JSON(500, gin.H{"error": "failed to publish baseline"})
		return
	}
	event, err := writeCollaborationEvent(tx, projectID, moduleID, actorID, "module.baseline_published", fmt.Sprintf("baseline:%s:%d", moduleID, version), map[string]any{"module_id": moduleID, "version": version, "baseline_id": versionID})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record baseline event"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit baseline"})
		return
	}
	emitProjectEvent(projectID, event)
	c.JSON(201, gin.H{"id": versionID, "version": version, "status": "published"})
}

func RestoreModuleBaselineHandler(c *gin.Context) {
	moduleID, rawVersion := c.Param("id"), c.Param("version")
	version, err := strconv.Atoi(rawVersion)
	if err != nil || version < 1 {
		validationError(c, "invalid version")
		return
	}
	var projectID string
	if err = db.DB.QueryRow(`SELECT project_id FROM modules WHERE id=$1 AND status='active'`, moduleID).Scan(&projectID); err != nil {
		c.JSON(404, gin.H{"error": "module not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}
	actorID, _ := middleware.GetUserID(c)
	var snapshot []byte
	if err = db.DB.QueryRow(`SELECT graph_snapshot FROM module_versions WHERE module_id=$1 AND version=$2`, moduleID, version).Scan(&snapshot); err != nil {
		c.JSON(404, gin.H{"error": "baseline not found"})
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start restore transaction"})
		return
	}
	defer tx.Rollback()
	var draftVersion int
	if err = tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM module_versions WHERE module_id=$1`, moduleID).Scan(&draftVersion); err != nil {
		c.JSON(500, gin.H{"error": "failed to allocate draft version"})
		return
	}
	draftID := "mv_" + GenerateUUID()
	if _, err = tx.Exec(`INSERT INTO module_versions(id,module_id,version,graph_snapshot,created_by,status,restored_from_version,diff_summary) VALUES($1,$2,$3,$4::jsonb,$5,'draft',$6,jsonb_build_object('restored_from',$6))`, draftID, moduleID, draftVersion, string(snapshot), actorID, version); err != nil {
		c.JSON(500, gin.H{"error": "failed to create restore draft"})
		return
	}
	event, err := writeCollaborationEvent(tx, projectID, moduleID, actorID, "module.restore_draft", fmt.Sprintf("restore:%s:%d", moduleID, draftVersion), map[string]any{"module_id": moduleID, "version": draftVersion, "restored_from_version": version})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to record restore event"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "failed to commit restore draft"})
		return
	}
	emitProjectEvent(projectID, event)
	c.JSON(201, gin.H{"id": draftID, "version": draftVersion, "status": "draft", "graph_snapshot": json.RawMessage(snapshot)})
}
