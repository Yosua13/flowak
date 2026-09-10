-- Migration: 06_tenant_auth_rbac.sql
-- Tenant-scoped invitations and opaque server-side session rotation.

CREATE TABLE IF NOT EXISTS organization_invitations (
    id VARCHAR(80) PRIMARY KEY,
    organization_id VARCHAR(80) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(150) NOT NULL,
    project_id VARCHAR(80) NULL REFERENCES projects(id) ON DELETE CASCADE,
    project_role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    functional_role VARCHAR(20) NULL,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP NULL,
    invited_by VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (project_role IN ('owner', 'editor', 'commenter', 'viewer')),
    CHECK (functional_role IS NULL OR functional_role IN ('pm', 'uiux', 'frontend', 'backend'))
);

CREATE INDEX IF NOT EXISTS idx_organization_invitations_lookup
    ON organization_invitations (organization_id, email, expires_at)
    WHERE accepted_at IS NULL;

CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(80) PRIMARY KEY,
    user_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id VARCHAR(80) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    refresh_token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_active
    ON user_sessions (user_id, organization_id, expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id VARCHAR(80) PRIMARY KEY,
    user_id VARCHAR(80) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_project_role_check;
ALTER TABLE project_members ADD CONSTRAINT project_members_project_role_check
    CHECK (project_role IN ('owner', 'editor', 'commenter', 'viewer'));
