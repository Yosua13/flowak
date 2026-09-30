-- Durable collaboration cursors, idempotent notification delivery, and bounded retention.
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS dedup_key VARCHAR(220) NULL;
CREATE UNIQUE INDEX IF NOT EXISTS notifications_dedup_key_unique ON notifications(dedup_key) WHERE dedup_key IS NOT NULL;

ALTER TABLE activity_logs ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP NULL;
UPDATE activity_logs SET expires_at = created_at + INTERVAL '180 days' WHERE expires_at IS NULL;
UPDATE api_runs SET expires_at = created_at + INTERVAL '30 days' WHERE expires_at IS NULL AND is_evidence = FALSE;

CREATE INDEX IF NOT EXISTS idx_event_outbox_project_cursor ON event_outbox(project_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_cursor ON notifications(user_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_expiry ON activity_logs(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_runs_expiry ON api_runs(expires_at) WHERE expires_at IS NOT NULL;
