package module

import "time"

// Node represents a workflow canvas node entity.
type Node struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Label        string         `json:"label"`
	X            float64        `json:"x"`
	Y            float64        `json:"y"`
	RowVersion   int            `json:"rowVersion"`
	Doc          map[string]any `json:"doc,omitempty"`
	Roles        map[string]any `json:"roles,omitempty"`
	Completeness any            `json:"completeness,omitempty"`
	LegacyNotes  map[string]any `json:"legacyNotes,omitempty"`
}

// Edge represents a directed workflow transition between two nodes.
type Edge struct {
	ID         string `json:"id"`
	From       string `json:"from"`
	To         string `json:"to"`
	Label      string `json:"label,omitempty"`
	Condition  string `json:"condition,omitempty"`
	RowVersion int    `json:"rowVersion"`
}

// GraphDelete represents an entity deletion request with concurrency version check.
type GraphDelete struct {
	ID         string `json:"id"`
	RowVersion int    `json:"rowVersion"`
}

// ModuleGraph represents the full workflow canvas graph data for a module.
type ModuleGraph struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Nodes         []Node         `json:"nodes"`
	Edges         []Edge         `json:"edges"`
	Version       int            `json:"version"`
	SchemaVersion int            `json:"schema_version"`
	UpdatedAt     *time.Time     `json:"updated_at,omitempty"`
}

// ModuleBaseline represents an immutable versioned snapshot of a module graph.
type ModuleBaseline struct {
	ID            string    `json:"id"`
	ModuleID      string    `json:"module_id"`
	Version       int       `json:"version"`
	GraphSnapshot string    `json:"graph_snapshot"`
	CreatedBy     string    `json:"created_by"`
	Status        string    `json:"status"`
	DiffSummary   string    `json:"diff_summary"`
	PublishedAt   time.Time `json:"published_at"`
}

// GraphReconciliationReport compares legacy JSON snapshots with active normalized graph rows.
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
