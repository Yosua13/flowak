package module

import (
	"context"
	"database/sql"
)

// ModuleRepository defines persistence operations for module workflow graphs and baselines.
type ModuleRepository interface {
	// FindGraphByModuleID loads the normalized workflow nodes, edges, doc, and facets.
	FindGraphByModuleID(ctx context.Context, moduleID string) (*ModuleGraph, error)

	// UpsertNodes synchronizes nodes, business rules, decision outcomes, and role specifications.
	UpsertNodes(ctx context.Context, tx *sql.Tx, moduleID string, nodes []Node) error

	// UpsertEdges synchronizes directed workflow transitions.
	UpsertEdges(ctx context.Context, tx *sql.Tx, moduleID string, edges []Edge) error

	// DeleteNodes tombstones specified nodes after validating edge disconnects.
	DeleteNodes(ctx context.Context, tx *sql.Tx, moduleID string, deletedNodes []GraphDelete) error

	// DeleteEdges tombstones specified edges.
	DeleteEdges(ctx context.Context, tx *sql.Tx, moduleID string, deletedEdges []GraphDelete) error

	// PublishBaseline creates an immutable versioned snapshot in module_versions.
	PublishBaseline(ctx context.Context, moduleID, actorID string) (*ModuleBaseline, error)

	// GetProjectIDByModuleID retrieves the parent project ID for access authorization.
	GetProjectIDByModuleID(ctx context.Context, moduleID string) (string, error)

	// UpdateModuleMeta updates module name, description, and increments version.
	UpdateModuleMeta(ctx context.Context, tx *sql.Tx, moduleID, name, description, userID string) error

	// BeginTx starts a database transaction.
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
