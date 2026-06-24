-- Migration: 04_saas_normalized_schema.sql
-- Description: Add SaaS tenant model and normalized workflow storage.

CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(80) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(120) UNIQUE NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'starter',
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO organizations (id, name, slug, plan, status)
VALUES ('org_default', 'Default Workspace', 'default-workspace', 'starter', 'active')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'active';
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP NULL;

CREATE TABLE IF NOT EXISTS organization_members (
    organization_id VARCHAR(80) NOT NULL,
    user_id VARCHAR(80) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (organization_id, user_id),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO organization_members (organization_id, user_id, role, status)
SELECT 'org_default', id, role, status
FROM users
ON CONFLICT (organization_id, user_id) DO NOTHING;

ALTER TABLE projects ADD COLUMN IF NOT EXISTS organization_id VARCHAR(80) NOT NULL DEFAULT 'org_default';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'active';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'projects_organization_id_fkey'
          AND table_name = 'projects'
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_organization_id_fkey
            FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
    END IF;
END $$;

ALTER TABLE project_members ADD COLUMN IF NOT EXISTS project_role VARCHAR(50) NOT NULL DEFAULT 'editor';
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS functional_role VARCHAR(50) NULL;
ALTER TABLE project_members ADD COLUMN IF NOT EXISTS added_by VARCHAR(80) NULL;

ALTER TABLE modules ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'active';
ALTER TABLE modules ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;
ALTER TABLE modules ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;
ALTER TABLE modules ADD COLUMN IF NOT EXISTS created_by VARCHAR(80) NULL;
ALTER TABLE modules ADD COLUMN IF NOT EXISTS updated_by VARCHAR(80) NULL;
ALTER TABLE modules ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE IF NOT EXISTS workflow_nodes (
    id VARCHAR(120) PRIMARY KEY,
    module_id VARCHAR(80) NOT NULL,
    type VARCHAR(40) NOT NULL,
    label VARCHAR(200) NOT NULL,
    x NUMERIC NOT NULL DEFAULT 0,
    y NUMERIC NOT NULL DEFAULT 0,
    actor TEXT NULL,
    trigger TEXT NULL,
    input_desc TEXT NULL,
    process_desc TEXT NULL,
    output_desc TEXT NULL,
    business_rules TEXT NULL,
    exception_path TEXT NULL,
    system_context TEXT NULL,
    sla_value NUMERIC NULL,
    sla_unit VARCHAR(30) NULL,
    priority VARCHAR(30) NOT NULL DEFAULT 'medium',
    risk_level VARCHAR(30) NOT NULL DEFAULT 'medium',
    acceptance_criteria TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    CHECK (type IN ('terminator', 'process', 'decision', 'actor', 'system'))
);

CREATE TABLE IF NOT EXISTS workflow_edges (
    id VARCHAR(120) PRIMARY KEY,
    module_id VARCHAR(80) NOT NULL,
    from_node_id VARCHAR(120) NOT NULL,
    to_node_id VARCHAR(120) NOT NULL,
    label VARCHAR(200) NULL,
    condition_text TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (from_node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (to_node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    UNIQUE (module_id, from_node_id, to_node_id),
    CHECK (from_node_id <> to_node_id)
);

CREATE TABLE IF NOT EXISTS node_role_tasks (
    id VARCHAR(160) PRIMARY KEY,
    node_id VARCHAR(120) NOT NULL,
    role_key VARCHAR(40) NOT NULL,
    assignee_id VARCHAR(80) NULL,
    assignee_label VARCHAR(150) NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'planned',
    reviewer_id VARCHAR(80) NULL,
    due_date DATE NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    notes TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (node_id, role_key),
    CHECK (role_key IN ('uiux', 'frontend', 'backend')),
    CHECK (status IN ('planned', 'in_progress', 'review', 'done'))
);

CREATE TABLE IF NOT EXISTS node_uiux_specs (
    node_id VARCHAR(120) PRIMARY KEY,
    screen_name VARCHAR(200) NULL,
    persona VARCHAR(150) NULL,
    prototype_url TEXT NULL,
    wireframe_url TEXT NULL,
    state_notes TEXT NULL,
    accessibility_notes TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS node_frontend_specs (
    node_id VARCHAR(120) PRIMARY KEY,
    page_name VARCHAR(200) NULL,
    route_path VARCHAR(200) NULL,
    interaction_notes TEXT NULL,
    validation_notes TEXT NULL,
    state_handling TEXT NULL,
    handoff_url TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS node_api_contracts (
    node_id VARCHAR(120) PRIMARY KEY,
    method VARCHAR(12) NULL,
    endpoint_path VARCHAR(300) NULL,
    auth_policy TEXT NULL,
    request_example JSONB NULL,
    response_example JSONB NULL,
    status_code VARCHAR(20) NULL,
    error_codes JSONB NULL,
    curl_example TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS module_versions (
    id VARCHAR(120) PRIMARY KEY,
    module_id VARCHAR(80) NOT NULL,
    version INT NOT NULL,
    graph_snapshot JSONB NOT NULL,
    created_by VARCHAR(80) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE (module_id, version)
);

CREATE TABLE IF NOT EXISTS activity_logs (
    id VARCHAR(120) PRIMARY KEY,
    organization_id VARCHAR(80) NULL,
    project_id VARCHAR(80) NULL,
    module_id VARCHAR(80) NULL,
    actor_id VARCHAR(80) NULL,
    action VARCHAR(80) NOT NULL,
    entity_type VARCHAR(80) NOT NULL,
    entity_id VARCHAR(120) NOT NULL,
    before_data JSONB NULL,
    after_data JSONB NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE SET NULL,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS comments (
    id VARCHAR(120) PRIMARY KEY,
    project_id VARCHAR(80) NOT NULL,
    module_id VARCHAR(80) NULL,
    node_id VARCHAR(120) NULL,
    author_id VARCHAR(80) NOT NULL,
    body TEXT NOT NULL,
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (node_id) REFERENCES workflow_nodes(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS document_exports (
    id VARCHAR(120) PRIMARY KEY,
    module_id VARCHAR(80) NOT NULL,
    export_type VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ready',
    file_name VARCHAR(255) NULL,
    created_by VARCHAR(80) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS ai_runs (
    id VARCHAR(120) PRIMARY KEY,
    organization_id VARCHAR(80) NULL,
    project_id VARCHAR(80) NULL,
    module_id VARCHAR(80) NULL,
    user_id VARCHAR(80) NULL,
    run_type VARCHAR(80) NOT NULL,
    prompt TEXT NULL,
    result_summary TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'success',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (module_id) REFERENCES modules(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(120) PRIMARY KEY,
    user_id VARCHAR(80) NOT NULL,
    project_id VARCHAR(80) NULL,
    title VARCHAR(180) NOT NULL,
    body TEXT NOT NULL,
    type VARCHAR(40) NOT NULL DEFAULT 'info',
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_modules_project ON modules(project_id);
CREATE INDEX IF NOT EXISTS idx_workflow_nodes_module ON workflow_nodes(module_id);
CREATE INDEX IF NOT EXISTS idx_workflow_edges_module ON workflow_edges(module_id);
CREATE INDEX IF NOT EXISTS idx_node_role_tasks_assignee ON node_role_tasks(assignee_id, status);
CREATE INDEX IF NOT EXISTS idx_node_role_tasks_due_date ON node_role_tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_activity_logs_project ON activity_logs(project_id, created_at DESC);
