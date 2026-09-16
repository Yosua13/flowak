-- Collaboration outbox and immutable module baselines.
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'published';
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS diff_summary JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS restored_from_version INT NULL;
ALTER TABLE module_versions ADD COLUMN IF NOT EXISTS published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE IF NOT EXISTS event_outbox (
  id VARCHAR(120) PRIMARY KEY, project_id VARCHAR(80) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  module_id VARCHAR(80) NULL REFERENCES modules(id) ON DELETE CASCADE, event_name VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL, idempotency_key VARCHAR(180) NOT NULL UNIQUE, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  delivered_at TIMESTAMP NULL, expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days'
);
CREATE TABLE IF NOT EXISTS notification_outbox (
  id VARCHAR(120) PRIMARY KEY, notification_id VARCHAR(120) NOT NULL UNIQUE REFERENCES notifications(id) ON DELETE CASCADE,
  recipient_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE CASCADE, dedup_key VARCHAR(200) NOT NULL UNIQUE,
  attempts INT NOT NULL DEFAULT 0, delivered_at TIMESTAMP NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '30 days'
);
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS event_name VARCHAR(100) NULL;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}'::jsonb;
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_event_outbox_project_created ON event_outbox(project_id, created_at DESC);
