-- Migration: 03_project_sharing.sql
-- Description: Create project_members table for project access control (RBAC & Sharing)

CREATE TABLE IF NOT EXISTS project_members (
    project_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, user_id),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
