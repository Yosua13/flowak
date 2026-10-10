package module

import (
	"context"
	"database/sql"
	"strings"

	"backend/internal/domain/module"
)

// ModulePostgresRepo implements module.ModuleRepository using PostgreSQL.
type ModulePostgresRepo struct {
	db *sql.DB
}

// NewModulePostgresRepo creates a new ModulePostgresRepo.
func NewModulePostgresRepo(db *sql.DB) *ModulePostgresRepo {
	return &ModulePostgresRepo{db: db}
}

// FindGraphByModuleID loads the normalized workflow nodes, edges, doc, and facets.
func (r *ModulePostgresRepo) FindGraphByModuleID(ctx context.Context, moduleID string) (*module.ModuleGraph, error) {
	return FindGraphByModuleID(ctx, r.db, moduleID)
}

// UpsertNodes synchronizes nodes, business rules, decision outcomes, and role specifications.
func (r *ModulePostgresRepo) UpsertNodes(ctx context.Context, tx *sql.Tx, moduleID string, nodes []module.Node) error {
	return UpsertNodes(ctx, tx, moduleID, nodes)
}

// UpsertEdges synchronizes directed workflow transitions.
func (r *ModulePostgresRepo) UpsertEdges(ctx context.Context, tx *sql.Tx, moduleID string, edges []module.Edge) error {
	return UpsertEdges(ctx, tx, moduleID, edges)
}

// DeleteNodes tombstones specified nodes after validating edge disconnects.
func (r *ModulePostgresRepo) DeleteNodes(ctx context.Context, tx *sql.Tx, moduleID string, deletedNodes []module.GraphDelete) error {
	return DeleteNodes(ctx, tx, moduleID, deletedNodes)
}

// DeleteEdges tombstones specified edges.
func (r *ModulePostgresRepo) DeleteEdges(ctx context.Context, tx *sql.Tx, moduleID string, deletedEdges []module.GraphDelete) error {
	return DeleteEdges(ctx, tx, moduleID, deletedEdges)
}

// PublishBaseline creates an immutable versioned snapshot in module_versions.
func (r *ModulePostgresRepo) PublishBaseline(ctx context.Context, moduleID, actorID string) (*module.ModuleBaseline, error) {
	return PublishBaseline(ctx, r.db, moduleID, actorID)
}

// GetProjectIDByModuleID retrieves parent project ID for authorization.
func (r *ModulePostgresRepo) GetProjectIDByModuleID(ctx context.Context, moduleID string) (string, error) {
	var projectID string
	err := r.db.QueryRowContext(ctx, "SELECT project_id FROM modules WHERE id = $1 AND status = 'active'", moduleID).Scan(&projectID)
	if err == sql.ErrNoRows {
		return "", module.ErrModuleNotFound
	}
	return projectID, err
}

// UpdateModuleMeta updates module name, description, and increments version.
func (r *ModulePostgresRepo) UpdateModuleMeta(ctx context.Context, tx *sql.Tx, moduleID, name, description, userID string) error {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name != "" || description != "" {
		_, err := tx.ExecContext(ctx, `
			UPDATE modules
			SET name = CASE WHEN $1 <> '' THEN $1 ELSE name END,
				description = CASE WHEN $2 <> '' THEN $2 ELSE description END,
				updated_by = $3,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
		`, name, description, userID, moduleID)
		if err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, "UPDATE modules SET version = version + 1, updated_by = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", userID, moduleID)
	return err
}

// BeginTx starts a database transaction.
func (r *ModulePostgresRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
