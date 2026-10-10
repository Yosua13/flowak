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

// ExecuteEnqueueAPIRun enqueues an API contract execution job to RabbitMQ when async mode
// is requested; otherwise it falls back immediately to synchronous execution.
func ExecuteEnqueueAPIRun(ctx context.Context, database *sql.DB, publisher *producer.APIRunnerJobPublisher, apiRequestID string, input models.APIRunRequest, actorID string, isAsync bool) (*models.APIRunResult, error) {
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
	`, apiRequestID, input.EnvironmentID).Scan(&method, &contractPath, &baseURL)
	if err != nil {
		return nil, fmt.Errorf("request or environment not found: %w", err)
	}

	relative := input.RelativePath
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

	if input.Method != "" {
		method = input.Method
	}
	method = strings.ToUpper(method)

	runID := "run_" + handlers.GenerateUUID()
	requestID := "req_" + handlers.GenerateUUID()
	now := time.Now().UTC()

	payload := producer.APIRunnerJobPayload{
		RunID:         runID,
		APIRequestID:  apiRequestID,
		EnvironmentID: input.EnvironmentID,
		ActorID:       actorID,
		Method:        method,
		RelativePath:  relative,
		Headers:       input.Headers,
		Body:          input.Body,
		Timestamp:     now,
	}

	// Asynchronous mode: Enqueue to RabbitMQ if publisher is active
	if isAsync && publisher != nil {
		insertQuery := `
			INSERT INTO api_runs (
				id, api_request_id, environment_id, actor_id, status, duration_ms,
				response_size, request_id, target_host, policy_decision, redacted_metadata,
				retained_body, body_truncated
			)
			VALUES ($1, $2, $3, $4, 'pending', 0, 0, $5, $6, 'allowed', '{}'::jsonb, NULL, false)
		`
		_, _ = database.ExecContext(ctx, insertQuery,
			runID,
			apiRequestID,
			input.EnvironmentID,
			actorID,
			requestID,
			base.Hostname(),
		)

		if err := publisher.PublishAPIRunnerJob(ctx, payload); err == nil {
			return &models.APIRunResult{
				ID:             runID,
				Status:         "pending",
				RequestID:      requestID,
				PolicyDecision: "allowed",
			}, nil
		}
		// If publishing fails unexpectedly, fall back to synchronous execution
	}

	// Synchronous fallback mode: Execute directly and return completed metrics
	return ExecuteProcessAPIRunJob(ctx, database, payload)
}
