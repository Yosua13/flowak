-- Migration: 08_domain_specifications.sql
-- Typed specification fields are filterable; repeatable business facts stay normalized.

ALTER TABLE workflow_nodes
    ADD COLUMN IF NOT EXISTS outcome TEXT,
    ADD COLUMN IF NOT EXISTS trigger_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS preconditions TEXT,
    ADD COLUMN IF NOT EXISTS reference_links TEXT;

ALTER TABLE node_role_tasks ADD COLUMN IF NOT EXISTS readiness VARCHAR(20) NOT NULL DEFAULT 'planned';
ALTER TABLE node_uiux_specs
    ADD COLUMN IF NOT EXISTS user_goal TEXT, ADD COLUMN IF NOT EXISTS surface VARCHAR(80),
    ADD COLUMN IF NOT EXISTS figma_frame_url TEXT, ADD COLUMN IF NOT EXISTS design_version VARCHAR(80),
    ADD COLUMN IF NOT EXISTS screen_states TEXT, ADD COLUMN IF NOT EXISTS interactions TEXT,
    ADD COLUMN IF NOT EXISTS content_messages TEXT, ADD COLUMN IF NOT EXISTS responsive_intent TEXT;
ALTER TABLE node_frontend_specs
    ADD COLUMN IF NOT EXISTS experience_name VARCHAR(200), ADD COLUMN IF NOT EXISTS entry_exit_behavior TEXT,
    ADD COLUMN IF NOT EXISTS input_requirements TEXT, ADD COLUMN IF NOT EXISTS api_references TEXT,
    ADD COLUMN IF NOT EXISTS analytics_intent TEXT, ADD COLUMN IF NOT EXISTS feature_availability TEXT;
ALTER TABLE node_api_contracts
    ADD COLUMN IF NOT EXISTS service_capability TEXT, ADD COLUMN IF NOT EXISTS api_references TEXT,
    ADD COLUMN IF NOT EXISTS business_validation TEXT, ADD COLUMN IF NOT EXISTS dependency_references TEXT,
    ADD COLUMN IF NOT EXISTS idempotency_notes TEXT, ADD COLUMN IF NOT EXISTS caching_notes TEXT,
    ADD COLUMN IF NOT EXISTS security_notes TEXT, ADD COLUMN IF NOT EXISTS observability_intent TEXT,
    ADD COLUMN IF NOT EXISTS sla_value NUMERIC, ADD COLUMN IF NOT EXISTS sla_unit VARCHAR(30);

CREATE TABLE IF NOT EXISTS node_business_rules (
    id VARCHAR(160) PRIMARY KEY, node_id VARCHAR(120) NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    rule_code VARCHAR(80), severity VARCHAR(20) NOT NULL DEFAULT 'medium', description TEXT NOT NULL, sort_order INT NOT NULL DEFAULT 0,
    CHECK (severity IN ('low', 'medium', 'high', 'critical'))
);
CREATE TABLE IF NOT EXISTS node_decision_outcomes (
    id VARCHAR(160) PRIMARY KEY, node_id VARCHAR(120) NOT NULL REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    edge_id VARCHAR(120) NULL REFERENCES workflow_edges(id) ON DELETE SET NULL, outcome TEXT NOT NULL, sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_business_rules_node ON node_business_rules(node_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_decision_outcomes_node ON node_decision_outcomes(node_id, sort_order);

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'workflow_nodes_trigger_type_check') THEN
    ALTER TABLE workflow_nodes ADD CONSTRAINT workflow_nodes_trigger_type_check CHECK (trigger_type IS NULL OR trigger_type IN ('manual','event','schedule','api'));
  END IF;
END $$;
