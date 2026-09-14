-- Migration: 09_legacy_organization_ownership.sql
-- Repair tenant roles created by the pre-RBAC migration and ensure every
-- legacy project owner has an explicit owner membership.

UPDATE organization_members om
SET role = CASE
    WHEN EXISTS (
        SELECT 1
        FROM projects p
        WHERE p.organization_id = om.organization_id
          AND p.owner_id = om.user_id
    ) OR EXISTS (
        SELECT 1 FROM users u WHERE u.id = om.user_id AND u.role = 'pm'
    ) THEN 'owner'
    ELSE 'member'
END
WHERE om.role NOT IN ('owner', 'member');

INSERT INTO organization_members (organization_id, user_id, role, status)
SELECT DISTINCT p.organization_id, p.owner_id, 'owner', 'active'
FROM projects p
ON CONFLICT (organization_id, user_id)
DO UPDATE SET role = 'owner', status = 'active';

INSERT INTO project_members (project_id, user_id, project_role, functional_role, added_by)
SELECT p.id, p.owner_id, 'owner', 'pm', p.owner_id
FROM projects p
ON CONFLICT (project_id, user_id)
DO UPDATE SET project_role = 'owner';

ALTER TABLE organization_members
    DROP CONSTRAINT IF EXISTS organization_members_role_check;

ALTER TABLE organization_members
    ADD CONSTRAINT organization_members_role_check
    CHECK (role IN ('owner', 'member'));
