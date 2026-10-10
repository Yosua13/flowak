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

type collaborationEvent struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Project  string          `json:"project_id"`
	ModuleID string          `json:"module_id,omitempty"`
	Payload  json.RawMessage `json:"payload"`
}

func collaborationEnvelope(eventID, projectID, moduleID, name string, payload []byte) []byte {
	envelope, _ := json.Marshal(collaborationEvent{ID: eventID, Name: name, Project: projectID, ModuleID: moduleID, Payload: payload})
	return envelope
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

// BroadcastProjectEvent forwards an event payload to active project SSE subscriber channels.
func BroadcastProjectEvent(projectID string, event []byte) {
	emitProjectEvent(projectID, event)
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
	return collaborationEnvelope(eventID, projectID, moduleID, name, raw), nil
}

func ListNotificationsHandler(c *gin.Context) {
	cursor, ok := decodeTimelineCursor(c)
	if !ok {
		return
	}
	limit, ok := timelineLimit(c)
	if !ok {
		return
	}
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	rows, err := db.DB.Query(`SELECT id,title,body,type,read_at,created_at,event_name FROM notifications WHERE user_id=$1 AND ($2::timestamp IS NULL OR (created_at,id) < ($2,$3)) ORDER BY created_at DESC,id DESC LIMIT $4`, userID, nullableCursorTime(cursor), nullableCursorID(cursor), limit+1)
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
	next := ""
	if len(out) > limit {
		last := out[limit-1]
		next = encodeTimelineCursor(last["timestamp"].(time.Time), last["id"].(string))
		out = out[:limit]
	}
	c.JSON(200, gin.H{"items": out, "next_cursor": next})
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
	c.Header("X-Accel-Buffering", "no")
	ch := make(chan []byte, 16)
	subscribers.Lock()
	if subscribers.byProject[projectID] == nil {
		subscribers.byProject[projectID] = map[chan []byte]struct{}{}
	}
	subscribers.byProject[projectID][ch] = struct{}{}
	subscribers.Unlock()
	defer func() { subscribers.Lock(); delete(subscribers.byProject[projectID], ch); subscribers.Unlock() }()
	if lastID := strings.TrimSpace(c.GetHeader("Last-Event-ID")); lastID != "" {
		if err := replayProjectEvents(c, projectID, lastID); err != nil {
			fmt.Fprint(c.Writer, "event: invalidate\ndata: {\"schema_version\":1}\n\n")
		}
	}
	fmt.Fprint(c.Writer, "event: ready\ndata: {\"schema_version\":1}\n\n")
	c.Writer.Flush()
	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()
	for {
		select {
		case event := <-ch:
			writeSSEChange(c, event)
		case <-ping.C:
			fmt.Fprint(c.Writer, "event: ping\ndata: {}\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func writeSSEChange(c *gin.Context, raw []byte) {
	var event collaborationEvent
	if json.Unmarshal(raw, &event) != nil || event.ID == "" {
		fmt.Fprint(c.Writer, "event: invalidate\ndata: {\"schema_version\":1}\n\n")
	} else {
		fmt.Fprintf(c.Writer, "id: %s\nevent: change\ndata: %s\n\n", event.ID, raw)
	}
	c.Writer.Flush()
}

// replayProjectEvents resumes a tenant-scoped stream after a browser reconnect.
// Unknown or expired cursors deliberately return an error so the client reloads
// canonical server state instead of assuming that its local cache is current.
func replayProjectEvents(c *gin.Context, projectID, lastID string) error {
	var createdAt time.Time
	if err := db.DB.QueryRow(`SELECT created_at FROM event_outbox WHERE project_id=$1 AND id=$2`, projectID, lastID).Scan(&createdAt); err != nil {
		return err
	}
	rows, err := db.DB.Query(`SELECT id,COALESCE(module_id,''),event_name,payload FROM event_outbox WHERE project_id=$1 AND (created_at,id) > ($2,$3) ORDER BY created_at,id LIMIT 101`, projectID, createdAt, lastID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, moduleID, name string
		var payload []byte
		if err := rows.Scan(&id, &moduleID, &name, &payload); err != nil {
			return err
		}
		writeSSEChange(c, collaborationEnvelope(id, projectID, moduleID, name, payload))
	}
	return rows.Err()
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
		dedup := strings.Join([]string{"mention", commentID, userID}, ":")
		if err := createNotification(tx, userID, projectID, dedup, "Anda disebut dalam komentar", "Buka work item untuk meninjau komentar.", "comment.mentioned", map[string]any{"comment_id": commentID}); err != nil {
			return err
		}
	}
	return nil
}

func createNotification(tx *sql.Tx, userID, projectID, dedup, title, body, eventName string, payload map[string]any) error {
	notificationID := "ntf_" + GenerateUUID()
	raw := safeEventPayload(payload)
	err := tx.QueryRow(`INSERT INTO notifications(id,user_id,project_id,title,body,type,event_name,payload,dedup_key) VALUES($1,$2,$3,$4,$5,'info',$6,$7::jsonb,$8) ON CONFLICT (dedup_key) WHERE dedup_key IS NOT NULL DO NOTHING RETURNING id`, notificationID, userID, projectID, title, body, eventName, string(raw), dedup).Scan(&notificationID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO notification_outbox(id,notification_id,recipient_id,dedup_key) VALUES($1,$2,$3,$4) ON CONFLICT(dedup_key) DO NOTHING`, "nout_"+GenerateUUID(), notificationID, userID, dedup)
	return err
}

// CreateWatcherNotifications is called from the transaction that changes a
// work item. The transition key makes retries idempotent, while actors never
// notify themselves.
func CreateWatcherNotifications(tx *sql.Tx, projectID, workItemID, actorID, transitionKey string) error {
	rows, err := tx.Query(`SELECT user_id FROM work_item_watchers WHERE work_item_id=$1`, workItemID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		if userID == actorID {
			continue
		}
		if err := createNotification(tx, userID, projectID, "watcher:"+transitionKey+":"+userID, "Work item yang diikuti diperbarui", "Buka work item untuk meninjau perubahan terbaru.", "work_item.watcher_updated", map[string]any{"work_item_id": workItemID}); err != nil {
			return err
		}
	}
	return rows.Err()
}

// PruneExpiredCollaborationRecords is safe to invoke on process start. Version
// baselines are immutable and intentionally excluded from expiry.
func PruneExpiredCollaborationRecords() error {
	for _, statement := range []string{
		`DELETE FROM event_outbox WHERE expires_at < CURRENT_TIMESTAMP`,
		`DELETE FROM notification_outbox WHERE expires_at < CURRENT_TIMESTAMP`,
		`DELETE FROM notifications WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '180 days'`,
		`DELETE FROM activity_logs WHERE expires_at < CURRENT_TIMESTAMP`,
		`DELETE FROM api_runs WHERE expires_at < CURRENT_TIMESTAMP`,
	} {
		if _, err := db.DB.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

// ListModuleBaselinesHandler exposes immutable baseline metadata in pages. The
// snapshot body remains intentionally unavailable from a list response.
func ListModuleBaselinesHandler(c *gin.Context) {
	cursor, ok := decodeTimelineCursor(c)
	if !ok {
		return
	}
	limit, ok := timelineLimit(c)
	if !ok {
		return
	}
	moduleID := c.Param("id")
	var projectID string
	if err := db.DB.QueryRow(`SELECT project_id FROM modules WHERE id=$1 AND status='active'`, moduleID).Scan(&projectID); err != nil {
		c.JSON(404, gin.H{"error": "module not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,version,status,created_at,published_at FROM module_versions WHERE module_id=$1 AND ($2::timestamp IS NULL OR (created_at,id) < ($2,$3)) ORDER BY created_at DESC,id DESC LIMIT $4`, moduleID, nullableCursorTime(cursor), nullableCursorID(cursor), limit+1)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list module baselines"})
		return
	}
	defer rows.Close()
	items := []gin.H{}
	for rows.Next() {
		var id, status string
		var version int
		var created time.Time
		var published *time.Time
		if err := rows.Scan(&id, &version, &status, &created, &published); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse module baseline"})
			return
		}
		items = append(items, gin.H{"id": id, "version": version, "status": status, "created_at": created, "published_at": published})
	}
	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{"error": "failed to list module baselines"})
		return
	}
	next := ""
	if len(items) > limit {
		last := items[limit-1]
		next = encodeTimelineCursor(last["created_at"].(time.Time), last["id"].(string))
		items = items[:limit]
	}
	c.JSON(200, gin.H{"items": items, "next_cursor": next})
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
