package apicontract

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"backend/handlers"
	"backend/internal/infra/rabbitmq/producer"
	"backend/models"
)

func nullIfEmpty(value string) any {
	if value == "" {
		return sql.NullString{}
	}
	return value
}

// ExecuteProcessAPIRunJob runs the HTTP target runner for an API request contract
// and persists the resulting execution metrics into PostgreSQL.
func ExecuteProcessAPIRunJob(ctx context.Context, database *sql.DB, job producer.APIRunnerJobPayload) (*models.APIRunResult, error) {
	if database == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	var method, contractPath, baseURL string
	err := database.QueryRowContext(ctx, `
		SELECT r.method, r.relative_path, e.approved_base_url
		FROM api_requests r
		JOIN workflow_nodes n ON n.id = r.node_id
		JOIN modules m ON m.id = n.module_id
		JOIN environments e ON e.id = $2
		WHERE r.id = $1
	`, job.APIRequestID, job.EnvironmentID).Scan(&method, &contractPath, &baseURL)
	if err != nil {
		return nil, fmt.Errorf("request or environment not found: %w", err)
	}

	relative := job.RelativePath
	if relative == "" {
		relative = contractPath
	}
	if !strings.HasPrefix(relative, "/") {
		return nil, fmt.Errorf("relative path required")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid environment URL: %w", err)
	}

	target := base.ResolveReference(&url.URL{Path: relative}).String()
	if job.Method != "" {
		method = job.Method
	}
	method = strings.ToUpper(method)

	runnerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	runner := handlers.APIRunner{AllowedHosts: map[string]bool{base.Hostname(): true}}
	result, runErr := runner.Run(runnerCtx, method, target, job.Body, job.Headers)

	status, decision := "succeeded", "allowed"
	if runErr != nil {
		status, decision = "blocked", "denied"
	}

	requestID := "req_" + handlers.GenerateUUID()

	// Upsert or insert into api_runs table
	upsertQuery := `
		INSERT INTO api_runs (
			id, api_request_id, environment_id, actor_id, status, duration_ms,
			response_size, request_id, target_host, policy_decision, redacted_metadata,
			retained_body, body_truncated
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, '{}'::jsonb, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			duration_ms = EXCLUDED.duration_ms,
			response_size = EXCLUDED.response_size,
			policy_decision = EXCLUDED.policy_decision,
			retained_body = EXCLUDED.retained_body,
			body_truncated = EXCLUDED.body_truncated
	`
	_, _ = database.ExecContext(ctx, upsertQuery,
		job.RunID,
		job.APIRequestID,
		job.EnvironmentID,
		job.ActorID,
		status,
		result.Duration.Milliseconds(),
		result.Size,
		requestID,
		base.Hostname(),
		decision,
		nullIfEmpty(result.Body),
		result.Truncated,
	)

	apiResult := &models.APIRunResult{
		ID:             job.RunID,
		Status:         status,
		StatusCode:     result.StatusCode,
		DurationMS:     result.Duration.Milliseconds(),
		ResponseSize:   result.Size,
		Headers:        result.Headers,
		Body:           result.Body,
		Truncated:      result.Truncated,
		RequestID:      requestID,
		PolicyDecision: decision,
	}

	if runErr != nil {
		return apiResult, runErr
	}

	return apiResult, nil
}
