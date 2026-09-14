-- Secure API contract runner. Secrets are stored encrypted and are never returned in full.
CREATE TABLE IF NOT EXISTS environments (
  id VARCHAR(120) PRIMARY KEY, project_id VARCHAR(80) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name VARCHAR(120) NOT NULL, approved_base_url TEXT NOT NULL, is_default BOOLEAN NOT NULL DEFAULT FALSE,
  runner_policy JSONB NOT NULL DEFAULT '{}'::jsonb, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(project_id, name)
);
CREATE UNIQUE INDEX IF NOT EXISTS environments_one_default ON environments(project_id) WHERE is_default;
CREATE TABLE IF NOT EXISTS environment_variables (
  id VARCHAR(120) PRIMARY KEY, environment_id VARCHAR(120) NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  variable_key VARCHAR(160) NOT NULL, encrypted_value BYTEA NOT NULL, is_secret BOOLEAN NOT NULL DEFAULT TRUE,
  updated_by VARCHAR(80) NULL REFERENCES users(id) ON DELETE SET NULL, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(environment_id, variable_key)
);
CREATE TABLE IF NOT EXISTS api_requests (
  id VARCHAR(120) PRIMARY KEY, node_id VARCHAR(120) NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
  name VARCHAR(160) NOT NULL, method VARCHAR(10) NOT NULL, relative_path TEXT NOT NULL,
  body_type VARCHAR(30) NOT NULL DEFAULT 'json', body_template TEXT, timeout_ms INT NOT NULL DEFAULT 10000,
  version INT NOT NULL DEFAULT 1, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CHECK (method IN ('GET','POST','PUT','PATCH','DELETE')), CHECK (timeout_ms BETWEEN 100 AND 30000)
);
CREATE TABLE IF NOT EXISTS api_request_parameters (
  id VARCHAR(120) PRIMARY KEY, api_request_id VARCHAR(120) NOT NULL REFERENCES api_requests(id) ON DELETE CASCADE,
  location VARCHAR(12) NOT NULL, parameter_key VARCHAR(160) NOT NULL, value_template TEXT, description TEXT,
  enabled BOOLEAN NOT NULL DEFAULT TRUE, secret_variable_id VARCHAR(120) NULL REFERENCES environment_variables(id) ON DELETE SET NULL,
  CHECK (location IN ('path','query','header','auth'))
);
CREATE TABLE IF NOT EXISTS api_response_examples (
  id VARCHAR(120) PRIMARY KEY, api_request_id VARCHAR(120) NOT NULL REFERENCES api_requests(id) ON DELETE CASCADE,
  status_code INT NOT NULL, content_type VARCHAR(160), body TEXT, is_error BOOLEAN NOT NULL DEFAULT FALSE, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS api_runs (
  id VARCHAR(120) PRIMARY KEY, api_request_id VARCHAR(120) NOT NULL REFERENCES api_requests(id) ON DELETE CASCADE,
  environment_id VARCHAR(120) NOT NULL REFERENCES environments(id) ON DELETE RESTRICT, actor_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  status VARCHAR(24) NOT NULL, duration_ms INT, response_size INT, request_id VARCHAR(120) NOT NULL,
  target_host TEXT NOT NULL, policy_decision TEXT NOT NULL, redacted_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  retained_body TEXT, body_truncated BOOLEAN NOT NULL DEFAULT FALSE, is_evidence BOOLEAN NOT NULL DEFAULT FALSE,
  expires_at TIMESTAMP NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS api_requests_node ON api_requests(node_id);
CREATE INDEX IF NOT EXISTS api_runs_request_created ON api_runs(api_request_id, created_at DESC);
