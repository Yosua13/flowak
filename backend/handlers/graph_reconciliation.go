package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"backend/db"
)

// GraphReconciliationReport compares legacy JSON snapshots with active normalized graph rows.
// It is diagnostic only; normalized rows remain the source of truth.
type GraphReconciliationReport struct {
	ModuleID  string                    `json:"module_id"`
	ProjectID string                    `json:"project_id"`
	Matches   bool                      `json:"matches"`
	Nodes     GraphEntityReconciliation `json:"nodes"`
	Edges     GraphEntityReconciliation `json:"edges"`
}

// GraphEntityReconciliation contains deterministic differences for one graph entity type.
type GraphEntityReconciliation struct {
	SnapshotCount         int                    `json:"snapshot_count"`
	NormalizedCount       int                    `json:"normalized_count"`
	MissingFromSnapshot   []string               `json:"missing_from_snapshot"`
	MissingFromNormalized []string               `json:"missing_from_normalized"`
	Changed               []GraphFieldDifference `json:"changed"`
	SnapshotIssues        []string               `json:"snapshot_issues"`
}

// GraphFieldDifference lists key graph fields with different values for a shared ID.
type GraphFieldDifference struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields"`
}

// ReconcileModuleGraph reads a module in a read-only transaction. It never repairs data.
func ReconcileModuleGraph(ctx context.Context, moduleID string) (GraphReconciliationReport, error) {
	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return GraphReconciliationReport{}, err
	}
	defer tx.Rollback()

	report, err := reconcileModuleGraphTx(ctx, tx, moduleID)
	if err != nil {
		return GraphReconciliationReport{}, err
	}
	if err := tx.Commit(); err != nil {
		return GraphReconciliationReport{}, err
	}
	return report, nil
}

// ReconcileProjectGraphs reads all modules in a project in one read-only transaction.
func ReconcileProjectGraphs(ctx context.Context, projectID string) ([]GraphReconciliationReport, error) {
	tx, err := db.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, "SELECT id FROM modules WHERE project_id = $1 ORDER BY id", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	moduleIDs := []string{}
	for rows.Next() {
		var moduleID string
		if err := rows.Scan(&moduleID); err != nil {
			return nil, err
		}
		moduleIDs = append(moduleIDs, moduleID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	reports := make([]GraphReconciliationReport, 0, len(moduleIDs))
	for _, moduleID := range moduleIDs {
		report, err := reconcileModuleGraphTx(ctx, tx, moduleID)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return reports, nil
}

func reconcileModuleGraphTx(ctx context.Context, tx *sql.Tx, moduleID string) (GraphReconciliationReport, error) {
	var projectID string
	var nodesSnapshot, edgesSnapshot sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT project_id, nodes, edges FROM modules WHERE id = $1", moduleID).Scan(&projectID, &nodesSnapshot, &edgesSnapshot)
	if err != nil {
		return GraphReconciliationReport{}, err
	}

	snapshotNodes, nodeCount, nodeIssues := parseSnapshotEntities(nodesSnapshot.String, "node")
	snapshotEdges, edgeCount, edgeIssues := parseSnapshotEntities(edgesSnapshot.String, "edge")
	normalizedNodes, err := readNormalizedNodes(ctx, tx, moduleID)
	if err != nil {
		return GraphReconciliationReport{}, err
	}
	normalizedEdges, err := readNormalizedEdges(ctx, tx, moduleID)
	if err != nil {
		return GraphReconciliationReport{}, err
	}

	report := GraphReconciliationReport{
		ModuleID:  moduleID,
		ProjectID: projectID,
		Nodes:     compareGraphEntities(snapshotNodes, normalizedNodes, nodeCount, nodeIssues, []string{"type", "label", "x", "y"}),
		Edges:     compareGraphEntities(snapshotEdges, normalizedEdges, edgeCount, edgeIssues, []string{"from", "to", "label", "condition"}),
	}
	report.Matches = graphEntitiesMatch(report.Nodes) && graphEntitiesMatch(report.Edges)
	return report, nil
}

func readNormalizedNodes(ctx context.Context, tx *sql.Tx, moduleID string) (map[string]map[string]any, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, type, label, x, y
		FROM workflow_nodes
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := map[string]map[string]any{}
	for rows.Next() {
		var id, nodeType, label string
		var x, y float64
		if err := rows.Scan(&id, &nodeType, &label, &x, &y); err != nil {
			return nil, err
		}
		nodes[id] = map[string]any{"id": id, "type": nodeType, "label": label, "x": x, "y": y}
	}
	return nodes, rows.Err()
}

func readNormalizedEdges(ctx context.Context, tx *sql.Tx, moduleID string) (map[string]map[string]any, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, from_node_id, to_node_id, label, condition_text
		FROM workflow_edges
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	edges := map[string]map[string]any{}
	for rows.Next() {
		var id, fromID, toID string
		var label, condition sql.NullString
		if err := rows.Scan(&id, &fromID, &toID, &label, &condition); err != nil {
			return nil, err
		}
		edges[id] = map[string]any{"id": id, "from": fromID, "to": toID, "label": nullableGraphValue(label), "condition": nullableGraphValue(condition)}
	}
	return edges, rows.Err()
}

func nullableGraphValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}

func parseSnapshotEntities(raw, entityName string) (map[string]map[string]any, int, []string) {
	entities := map[string]map[string]any{}
	if raw == "" {
		return entities, 0, []string{fmt.Sprintf("%s snapshot is empty", entityName)}
	}

	var values []any
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return entities, 0, []string{fmt.Sprintf("%s snapshot is not a JSON array", entityName)}
	}
	issues := []string{}
	for index, value := range values {
		entity, ok := value.(map[string]any)
		if !ok {
			issues = append(issues, fmt.Sprintf("%s snapshot item %d is not an object", entityName, index))
			continue
		}
		id, _ := entity["id"].(string)
		if id == "" {
			issues = append(issues, fmt.Sprintf("%s snapshot item %d has no id", entityName, index))
			continue
		}
		if _, exists := entities[id]; exists {
			issues = append(issues, fmt.Sprintf("%s snapshot contains duplicate id %q", entityName, id))
			continue
		}
		entities[id] = entity
	}
	return entities, len(values), issues
}

func compareGraphEntities(snapshot, normalized map[string]map[string]any, snapshotCount int, snapshotIssues, fields []string) GraphEntityReconciliation {
	reconciliation := GraphEntityReconciliation{
		SnapshotCount: snapshotCount, NormalizedCount: len(normalized),
		MissingFromSnapshot: []string{}, MissingFromNormalized: []string{},
		Changed: []GraphFieldDifference{}, SnapshotIssues: snapshotIssues,
	}
	for id, normalizedEntity := range normalized {
		snapshotEntity, exists := snapshot[id]
		if !exists {
			reconciliation.MissingFromSnapshot = append(reconciliation.MissingFromSnapshot, id)
			continue
		}
		changedFields := []string{}
		for _, field := range fields {
			if !graphValueEqual(snapshotEntity[field], normalizedEntity[field]) {
				changedFields = append(changedFields, field)
			}
		}
		if len(changedFields) > 0 {
			reconciliation.Changed = append(reconciliation.Changed, GraphFieldDifference{ID: id, Fields: changedFields})
		}
	}
	for id := range snapshot {
		if _, exists := normalized[id]; !exists {
			reconciliation.MissingFromNormalized = append(reconciliation.MissingFromNormalized, id)
		}
	}
	sort.Strings(reconciliation.MissingFromSnapshot)
	sort.Strings(reconciliation.MissingFromNormalized)
	sort.Slice(reconciliation.Changed, func(i, j int) bool { return reconciliation.Changed[i].ID < reconciliation.Changed[j].ID })
	sort.Strings(reconciliation.SnapshotIssues)
	return reconciliation
}

func graphValueEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func graphEntitiesMatch(entity GraphEntityReconciliation) bool {
	return entity.SnapshotCount == entity.NormalizedCount && len(entity.MissingFromSnapshot) == 0 && len(entity.MissingFromNormalized) == 0 && len(entity.Changed) == 0 && len(entity.SnapshotIssues) == 0
}
