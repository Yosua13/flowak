package handlers

import (
	"backend/db"
	"backend/middleware"
	"backend/models"
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
	"time"
)

func RunAPIRequestHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var input models.APIRunRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid runner input"})
		return
	}
	var projectID, method, contractPath, baseURL string
	err = db.DB.QueryRow(`SELECT m.project_id,r.method,r.relative_path,e.approved_base_url FROM api_requests r JOIN workflow_nodes n ON n.id=r.node_id JOIN modules m ON m.id=n.module_id JOIN environments e ON e.id=$2 WHERE r.id=$1`, c.Param("id"), input.EnvironmentID).Scan(&projectID, &method, &contractPath, &baseURL)
	if err != nil {
		c.JSON(404, gin.H{"error": "request or environment not found"})
		return
	}
	if !middleware.AuthorizeProject(c, projectID, middleware.CapabilityView) {
		return
	}
	relative := input.RelativePath
	if relative == "" {
		relative = contractPath
	}
	if !strings.HasPrefix(relative, "/") {
		c.JSON(400, gin.H{"error": "relative path required"})
		return
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "invalid environment"})
		return
	}
	target := base.ResolveReference(&url.URL{Path: relative}).String()
	if input.Method != "" {
		method = input.Method
	}
	method = strings.ToUpper(method)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	result, runErr := (APIRunner{AllowedHosts: map[string]bool{base.Hostname(): true}}).Run(ctx, method, target, input.Body, input.Headers)
	status, decision := "succeeded", "allowed"
	if runErr != nil {
		status, decision = "blocked", "denied"
	}
	runID, requestID := "run_"+GenerateUUID(), "req_"+GenerateUUID()
	_, _ = db.DB.Exec(`INSERT INTO api_runs(id,api_request_id,environment_id,actor_id,status,duration_ms,response_size,request_id,target_host,policy_decision,redacted_metadata,retained_body,body_truncated) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'{}'::jsonb,$11,$12)`, runID, c.Param("id"), input.EnvironmentID, userID, status, result.Duration.Milliseconds(), result.Size, requestID, base.Hostname(), decision, nullIfEmpty(result.Body), result.Truncated)
	if runErr != nil {
		c.JSON(400, gin.H{"error": runErr.Error(), "policy_decision": decision, "request_id": requestID})
		return
	}
	c.JSON(200, models.APIRunResult{ID: runID, Status: status, StatusCode: result.StatusCode, DurationMS: result.Duration.Milliseconds(), ResponseSize: result.Size, Headers: result.Headers, Body: result.Body, Truncated: result.Truncated, RequestID: requestID, PolicyDecision: decision})
}

func nullIfEmpty(value string) any {
	if value == "" {
		return sql.NullString{}
	}
	return value
}
