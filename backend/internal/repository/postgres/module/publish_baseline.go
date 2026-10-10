package module

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/domain/module"
)

// PublishBaseline snapshots normalized graph nodes and edges into immutable module_versions.
func PublishBaseline(ctx context.Context, db *sql.DB, moduleID, actorID string) (*module.ModuleBaseline, error) {
	graph, err := FindGraphByModuleID(ctx, db, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to load graph for baseline: %w", err)
	}

	snapshotBytes, err := json.Marshal(map[string]any{
		"schemaVersion": graph.SchemaVersion,
		"nodes":         graph.Nodes,
		"edges":         graph.Edges,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graph snapshot: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin baseline tx: %w", err)
	}
	defer tx.Rollback()

	var version int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM module_versions WHERE module_id = $1`, moduleID).Scan(&version); err != nil {
		return nil, fmt.Errorf("failed to allocate baseline version: %w", err)
	}

	versionID := "mv_" + generateUUID()
	var publishedAt time.Time

	err = tx.QueryRowContext(ctx, `
		INSERT INTO module_versions (
			id, module_id, version, graph_snapshot, created_by, status, diff_summary, published_at
		) VALUES (
			$1, $2, $3, $4::jsonb, $5, 'published', jsonb_build_object('node_count', $6, 'edge_count', $7), CURRENT_TIMESTAMP
		) RETURNING published_at
	`, versionID, moduleID, version, string(snapshotBytes), actorID, len(graph.Nodes), len(graph.Edges)).Scan(&publishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert module version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit baseline: %w", err)
	}

	return &module.ModuleBaseline{
		ID:            versionID,
		ModuleID:      moduleID,
		Version:       version,
		GraphSnapshot: string(snapshotBytes),
		CreatedBy:     actorID,
		Status:        "published",
		DiffSummary:   fmt.Sprintf(`{"node_count":%d,"edge_count":%d}`, len(graph.Nodes), len(graph.Edges)),
		PublishedAt:   publishedAt,
	}, nil
}
