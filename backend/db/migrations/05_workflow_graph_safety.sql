-- Migration: 05_workflow_graph_safety.sql
-- Keep normalized workflow records as the source of truth while retaining tombstones for audit.

ALTER TABLE workflow_nodes
    ADD COLUMN IF NOT EXISTS row_version INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

ALTER TABLE workflow_edges
    ADD COLUMN IF NOT EXISTS row_version INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS idx_workflow_nodes_active_module
    ON workflow_nodes(module_id, sort_order)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workflow_edges_active_module
    ON workflow_edges(module_id, sort_order)
    WHERE deleted_at IS NULL;
