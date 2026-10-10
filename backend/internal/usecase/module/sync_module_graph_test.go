package module_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"backend/internal/domain/module"
	usecaseModule "backend/internal/usecase/module"
	"github.com/DATA-DOG/go-sqlmock"
)

type mockRepo struct {
	findGraphFn        func(ctx context.Context, moduleID string) (*module.ModuleGraph, error)
	upsertNodesFn      func(ctx context.Context, tx *sql.Tx, moduleID string, nodes []module.Node) error
	upsertEdgesFn      func(ctx context.Context, tx *sql.Tx, moduleID string, edges []module.Edge) error
	deleteNodesFn      func(ctx context.Context, tx *sql.Tx, moduleID string, deletedNodes []module.GraphDelete) error
	deleteEdgesFn      func(ctx context.Context, tx *sql.Tx, moduleID string, deletedEdges []module.GraphDelete) error
	publishBaselineFn  func(ctx context.Context, moduleID, actorID string) (*module.ModuleBaseline, error)
	getProjectIDFn     func(ctx context.Context, moduleID string) (string, error)
	updateModuleMetaFn func(ctx context.Context, tx *sql.Tx, moduleID, name, description, userID string) error
	beginTxFn          func(ctx context.Context) (*sql.Tx, error)
}

func (m *mockRepo) FindGraphByModuleID(ctx context.Context, moduleID string) (*module.ModuleGraph, error) {
	if m.findGraphFn != nil {
		return m.findGraphFn(ctx, moduleID)
	}
	return &module.ModuleGraph{ID: moduleID}, nil
}

func (m *mockRepo) UpsertNodes(ctx context.Context, tx *sql.Tx, moduleID string, nodes []module.Node) error {
	if m.upsertNodesFn != nil {
		return m.upsertNodesFn(ctx, tx, moduleID, nodes)
	}
	return nil
}

func (m *mockRepo) UpsertEdges(ctx context.Context, tx *sql.Tx, moduleID string, edges []module.Edge) error {
	if m.upsertEdgesFn != nil {
		return m.upsertEdgesFn(ctx, tx, moduleID, edges)
	}
	return nil
}

func (m *mockRepo) DeleteNodes(ctx context.Context, tx *sql.Tx, moduleID string, deletedNodes []module.GraphDelete) error {
	if m.deleteNodesFn != nil {
		return m.deleteNodesFn(ctx, tx, moduleID, deletedNodes)
	}
	return nil
}

func (m *mockRepo) DeleteEdges(ctx context.Context, tx *sql.Tx, moduleID string, deletedEdges []module.GraphDelete) error {
	if m.deleteEdgesFn != nil {
		return m.deleteEdgesFn(ctx, tx, moduleID, deletedEdges)
	}
	return nil
}

func (m *mockRepo) PublishBaseline(ctx context.Context, moduleID, actorID string) (*module.ModuleBaseline, error) {
	if m.publishBaselineFn != nil {
		return m.publishBaselineFn(ctx, moduleID, actorID)
	}
	return &module.ModuleBaseline{ID: "base_1", ModuleID: moduleID}, nil
}

func (m *mockRepo) GetProjectIDByModuleID(ctx context.Context, moduleID string) (string, error) {
	if m.getProjectIDFn != nil {
		return m.getProjectIDFn(ctx, moduleID)
	}
	return "proj_1", nil
}

func (m *mockRepo) UpdateModuleMeta(ctx context.Context, tx *sql.Tx, moduleID, name, description, userID string) error {
	if m.updateModuleMetaFn != nil {
		return m.updateModuleMetaFn(ctx, tx, moduleID, name, description, userID)
	}
	return nil
}

func (m *mockRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if m.beginTxFn != nil {
		return m.beginTxFn(ctx)
	}
	return nil, nil
}

func TestExecuteSyncModuleGraph_CycleRejection(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}

	req := usecaseModule.SyncModuleGraphRequest{
		Nodes: []module.Node{{ID: "A"}, {ID: "B"}},
		Edges: []module.Edge{
			{ID: "e1", From: "A", To: "B"},
			{ID: "e2", From: "B", To: "A"},
		},
	}

	_, err := usecaseModule.ExecuteSyncModuleGraph(ctx, repo, nil, "mod_1", "usr_1", req)
	if !errors.Is(err, module.ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got: %v", err)
	}
}

func TestExecuteSyncModuleGraph_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	ctx := context.Background()
	repo := &mockRepo{
		beginTxFn: func(ctx context.Context) (*sql.Tx, error) {
			return db.BeginTx(ctx, nil)
		},
		findGraphFn: func(ctx context.Context, moduleID string) (*module.ModuleGraph, error) {
			return &module.ModuleGraph{
				ID:    moduleID,
				Nodes: []module.Node{{ID: "A"}, {ID: "B"}},
				Edges: []module.Edge{{ID: "e1", From: "A", To: "B"}},
			}, nil
		},
	}

	req := usecaseModule.SyncModuleGraphRequest{
		Nodes: []module.Node{{ID: "A"}, {ID: "B"}},
		Edges: []module.Edge{{ID: "e1", From: "A", To: "B"}},
	}

	g, err := usecaseModule.ExecuteSyncModuleGraph(ctx, repo, nil, "mod_1", "usr_1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g == nil || len(g.Nodes) != 2 || len(g.Edges) != 1 {
		t.Fatalf("unexpected graph result: %+v", g)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteGetModuleGraph_FallbackDB(t *testing.T) {
	ctx := context.Background()
	calledDB := false

	repo := &mockRepo{
		findGraphFn: func(ctx context.Context, moduleID string) (*module.ModuleGraph, error) {
			calledDB = true
			return &module.ModuleGraph{ID: moduleID, Name: "DB Flow"}, nil
		},
	}

	g, err := usecaseModule.ExecuteGetModuleGraph(ctx, repo, nil, "mod_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !calledDB || g.Name != "DB Flow" {
		t.Fatalf("expected fallback to DB repo, got: %+v", g)
	}
}

func TestExecutePublishBaseline_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{
		publishBaselineFn: func(ctx context.Context, moduleID, actorID string) (*module.ModuleBaseline, error) {
			return &module.ModuleBaseline{
				ID:        "mv_1",
				ModuleID:  moduleID,
				Version:   1,
				CreatedBy: actorID,
				Status:    "published",
			}, nil
		},
	}

	base, err := usecaseModule.ExecutePublishBaseline(ctx, repo, "mod_1", "usr_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base == nil || base.Version != 1 || base.CreatedBy != "usr_1" {
		t.Fatalf("unexpected baseline: %+v", base)
	}
}
