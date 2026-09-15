-- Migration: 07_work_items.sql
-- Work items are independent execution records; workflow nodes remain graph records.

ALTER TABLE projects ADD COLUMN IF NOT EXISTS work_item_prefix VARCHAR(20) NOT NULL DEFAULT '';
UPDATE projects
SET work_item_prefix = 'P' || UPPER(SUBSTRING(MD5(id) FROM 1 FOR 8))
WHERE work_item_prefix IN ('', 'PRJ');
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'projects_work_item_prefix_key') THEN
        ALTER TABLE projects ADD CONSTRAINT projects_work_item_prefix_key UNIQUE (work_item_prefix);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS project_work_item_sequences (
    project_id VARCHAR(80) PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    next_value BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS work_items (
    id VARCHAR(120) PRIMARY KEY,
    work_key VARCHAR(160) NOT NULL UNIQUE,
    project_id VARCHAR(80) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    module_id VARCHAR(80) NULL REFERENCES modules(id) ON DELETE SET NULL,
    node_id VARCHAR(120) NULL REFERENCES workflow_nodes(id) ON DELETE SET NULL,
    facet_key VARCHAR(80) NULL,
    parent_id VARCHAR(120) NULL REFERENCES work_items(id) ON DELETE SET NULL,
    sequence BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    points SMALLINT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'Backlog',
    assignee_id VARCHAR(80) NULL REFERENCES users(id) ON DELETE SET NULL,
    reporter_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    start_date DATE NULL,
    due_date DATE NULL,
    blocked_reason TEXT NULL,
    resolution TEXT NULL,
    row_version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, sequence),
    CHECK (type IN ('Story', 'Task', 'Bug', 'Review', 'Research', 'Subtask')),
    CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CHECK (points IS NULL OR points IN (1, 2, 3, 5, 8, 13)),
    CHECK (status IN ('Backlog', 'Ready', 'In Progress', 'In Review', 'Blocked', 'Done', 'Canceled')),
    CHECK (parent_id IS NULL OR parent_id <> id),
    CHECK (status <> 'Blocked' OR blocked_reason IS NOT NULL),
    CHECK (status <> 'Done' OR resolution IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS work_item_status_history (
    id VARCHAR(120) PRIMARY KEY,
    work_item_id VARCHAR(120) NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
    from_status VARCHAR(20) NULL,
    to_status VARCHAR(20) NOT NULL,
    changed_by VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    note TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS work_item_links (
    id VARCHAR(120) PRIMARY KEY,
    work_item_id VARCHAR(120) NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
    linked_work_item_id VARCHAR(120) NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
    link_type VARCHAR(30) NOT NULL DEFAULT 'relates_to',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (work_item_id, linked_work_item_id, link_type),
    CHECK (work_item_id <> linked_work_item_id)
);

CREATE TABLE IF NOT EXISTS attachments (
    id VARCHAR(120) PRIMARY KEY,
    project_id VARCHAR(80) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    work_item_id VARCHAR(120) NULL REFERENCES work_items(id) ON DELETE CASCADE,
    node_id VARCHAR(120) NULL REFERENCES workflow_nodes(id) ON DELETE SET NULL,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(120) NULL,
    storage_key TEXT NOT NULL,
    size_bytes BIGINT NULL,
    uploaded_by VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK ((work_item_id IS NOT NULL)::int + (node_id IS NOT NULL)::int = 1)
);

CREATE TABLE IF NOT EXISTS work_item_watchers (
    work_item_id VARCHAR(120) NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
    user_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (work_item_id, user_id)
);

CREATE TABLE IF NOT EXISTS work_item_acceptance_criteria (
    id VARCHAR(120) PRIMARY KEY,
    work_item_id VARCHAR(120) NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    is_complete BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE comments ADD COLUMN IF NOT EXISTS work_item_id VARCHAR(120) NULL REFERENCES work_items(id) ON DELETE CASCADE;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS parent_id VARCHAR(120) NULL REFERENCES comments(id) ON DELETE SET NULL;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS mentions JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE comments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'comments_one_target') THEN
        ALTER TABLE comments ADD CONSTRAINT comments_one_target
            CHECK ((node_id IS NOT NULL)::int + (work_item_id IS NOT NULL)::int = 1) NOT VALID;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_work_items_project_active ON work_items(project_id, sequence DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_items_node_active ON work_items(node_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_items_parent_active ON work_items(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_work_item_history_item ON work_item_status_history(work_item_id, created_at);
CREATE INDEX IF NOT EXISTS idx_comments_work_item ON comments(work_item_id, created_at) WHERE deleted_at IS NULL;
