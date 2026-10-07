package handlers

import (
	"strings"
	"time"

	"backend/db"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

func ListAPIRequestsHandler(c *gin.Context) {
	projectID, err := contractProjectForNode(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,name,method,relative_path,body_type,body_template,timeout_ms,version,created_at,updated_at FROM api_requests WHERE node_id=$1 ORDER BY name`, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list API requests"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, name, method, path, bodyType string
		var body *string
		var timeout, version int
		var created, updated any
		if rows.Scan(&id, &name, &method, &path, &bodyType, &body, &timeout, &version, &created, &updated) != nil {
			c.JSON(500, gin.H{"error": "failed to parse API requests"})
			return
		}
		out = append(out, gin.H{"id": id, "name": name, "method": method, "relative_path": path, "body_type": bodyType, "body_template": body, "timeout_ms": timeout, "version": version, "created_at": created, "updated_at": updated})
	}
	c.JSON(200, out)
}
func CreateAPIRequestHandler(c *gin.Context) {
	projectID, err := contractProjectForNode(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}
	var in struct {
		Name         string `json:"name"`
		Method       string `json:"method"`
		RelativePath string `json:"relative_path"`
		BodyType     string `json:"body_type"`
		BodyTemplate string `json:"body_template"`
		Timeout      int    `json:"timeout_ms"`
	}
	if c.ShouldBindJSON(&in) != nil || strings.TrimSpace(in.Name) == "" || !strings.HasPrefix(in.RelativePath, "/") {
		contractBadRequest(c, "request name and relative path are required")
		return
	}
	in.Method = strings.ToUpper(in.Method)
	if in.Method == "" {
		in.Method = "GET"
	}
	if in.Method != "GET" && in.Method != "POST" && in.Method != "PUT" && in.Method != "PATCH" && in.Method != "DELETE" {
		contractBadRequest(c, "unsupported request method")
		return
	}
	if in.BodyType == "" {
		in.BodyType = "json"
	}
	if in.Timeout == 0 {
		in.Timeout = 10000
	}
	if in.Timeout < 100 || in.Timeout > 30000 {
		contractBadRequest(c, "timeout must be 100-30000ms")
		return
	}
	id := "api_" + GenerateUUID()
	_, err = db.DB.Exec(`INSERT INTO api_requests(id,node_id,name,method,relative_path,body_type,body_template,timeout_ms) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, c.Param("id"), strings.TrimSpace(in.Name), in.Method, in.RelativePath, in.BodyType, in.BodyTemplate, in.Timeout)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create API request"})
		return
	}
	c.JSON(201, gin.H{"id": id})
}
func UpdateAPIRequestHandler(c *gin.Context) {
	projectID, err := contractProjectForRequest(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}
	var in struct {
		Name         *string `json:"name"`
		Method       *string `json:"method"`
		RelativePath *string `json:"relative_path"`
		BodyTemplate *string `json:"body_template"`
		Timeout      *int    `json:"timeout_ms"`
	}
	if c.ShouldBindJSON(&in) != nil {
		contractBadRequest(c, "invalid API request input")
		return
	}
	if in.RelativePath != nil && !strings.HasPrefix(*in.RelativePath, "/") {
		contractBadRequest(c, "relative path required")
		return
	}
	_, err = db.DB.Exec(`UPDATE api_requests SET name=COALESCE(NULLIF($1,''),name),method=COALESCE(NULLIF($2,''),method),relative_path=COALESCE(NULLIF($3,''),relative_path),body_template=COALESCE($4,body_template),timeout_ms=COALESCE($5,timeout_ms),version=version+1,updated_at=CURRENT_TIMESTAMP WHERE id=$6`, in.Name, in.Method, in.RelativePath, in.BodyTemplate, in.Timeout, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update API request"})
		return
	}
	c.Status(204)
}
func DeleteAPIRequestHandler(c *gin.Context) {
	projectID, err := contractProjectForRequest(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityEditGraph) {
		return
	}
	_, err = db.DB.Exec(`DELETE FROM api_requests WHERE id=$1`, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete API request"})
		return
	}
	c.Status(204)
}
func ListAPIRunsHandler(c *gin.Context) {
	cursor, ok := decodeTimelineCursor(c)
	if !ok {
		return
	}
	limit, ok := timelineLimit(c)
	if !ok {
		return
	}
	projectID, err := contractProjectForRequest(c.Param("id"))
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityView) {
		return
	}
	rows, err := db.DB.Query(`SELECT id,status,duration_ms,response_size,request_id,target_host,policy_decision,body_truncated,is_evidence,created_at FROM api_runs WHERE api_request_id=$1 AND ($2::timestamp IS NULL OR (created_at,id) < ($2,$3)) ORDER BY created_at DESC,id DESC LIMIT $4`, c.Param("id"), nullableCursorTime(cursor), nullableCursorID(cursor), limit+1)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list API runs"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, status, requestID, host, decision string
		var duration, size *int
		var truncated, evidence bool
		var created any
		if rows.Scan(&id, &status, &duration, &size, &requestID, &host, &decision, &truncated, &evidence, &created) != nil {
			c.JSON(500, gin.H{"error": "failed to parse API runs"})
			return
		}
		out = append(out, gin.H{"id": id, "status": status, "duration_ms": duration, "response_size": size, "request_id": requestID, "target_host": host, "policy_decision": decision, "body_truncated": truncated, "is_evidence": evidence, "created_at": created})
	}
	next := ""
	if len(out) > limit {
		last := out[limit-1]
		created, valid := last["created_at"].(time.Time)
		if !valid {
			c.JSON(500, gin.H{"error": "failed to parse API run cursor"})
			return
		}
		next = encodeTimelineCursor(created, last["id"].(string))
		out = out[:limit]
	}
	c.JSON(200, gin.H{"items": out, "next_cursor": next})
}
func SaveAPIRunEvidenceHandler(c *gin.Context) {
	var projectID string
	err := db.DB.QueryRow(`SELECT m.project_id FROM api_runs x JOIN api_requests r ON r.id=x.api_request_id JOIN workflow_nodes n ON n.id=r.node_id JOIN modules m ON m.id=n.module_id WHERE x.id=$1`, c.Param("id")).Scan(&projectID)
	if err != nil {
		contractNotFound(c)
		return
	}
	if !authorizedContractProject(c, projectID, middleware.CapabilityWorkItem) {
		return
	}
	_, err = db.DB.Exec(`UPDATE api_runs SET is_evidence=TRUE WHERE id=$1`, c.Param("id"))
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to save evidence"})
		return
	}
	c.Status(204)
}
