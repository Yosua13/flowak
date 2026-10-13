package handlers

import (
	"context"

	"backend/db"
	domainModule "backend/internal/domain/module"
	usecaseModule "backend/internal/usecase/module"
)

// GraphReconciliationReport aliases domainModule.GraphReconciliationReport for backward compatibility.
type GraphReconciliationReport = domainModule.GraphReconciliationReport

// GraphEntityReconciliation aliases domainModule.GraphEntityReconciliation.
type GraphEntityReconciliation = domainModule.GraphEntityReconciliation

// GraphFieldDifference aliases domainModule.GraphFieldDifference.
type GraphFieldDifference = domainModule.GraphFieldDifference

// ReconcileModuleGraph reads a module in a read-only transaction and compares
// legacy JSON snapshots with active normalized graph rows.
func ReconcileModuleGraph(ctx context.Context, moduleID string) (GraphReconciliationReport, error) {
	return usecaseModule.ExecuteReconcileGraph(ctx, db.DB, moduleID)
}

// ReconcileProjectGraphs reads all modules in a project in one read-only transaction.
func ReconcileProjectGraphs(ctx context.Context, projectID string) ([]GraphReconciliationReport, error) {
	return usecaseModule.ExecuteReconcileProject(ctx, db.DB, projectID)
}
