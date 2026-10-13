package module

import (
	"context"
	"fmt"
	"time"

	"backend/internal/domain/module"
	"backend/internal/repository/redis/cache"
	"backend/internal/repository/redis/lock"
	goredis "github.com/redis/go-redis/v9"
)

// SyncModuleGraphRequest holds the payload for synchronizing module graph entities.
type SyncModuleGraphRequest struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Nodes        []module.Node        `json:"nodes"`
	Edges        []module.Edge        `json:"edges"`
	DeletedNodes []module.GraphDelete `json:"deletedNodes"`
	DeletedEdges []module.GraphDelete `json:"deletedEdges"`
}

// ExecuteSyncModuleGraph synchronizes nodes, edges, doc, and facets.
// Order of operations:
// 1. Acquire distributed lock (lock:module:<id>:sync)
// 2. Validate acyclic topology via cycle detector
// 3. Persist to PostgreSQL (upsert/delete) in transaction
// 4. Invalidate Redis hot graph cache (cache:module:<id>:graph)
// 5. Release distributed lock
func ExecuteSyncModuleGraph(
	ctx context.Context,
	repo module.ModuleRepository,
	redisClient *goredis.Client,
	moduleID string,
	userID string,
	req SyncModuleGraphRequest,
) (*module.ModuleGraph, error) {
	if moduleID == "" {
		return nil, fmt.Errorf("moduleID cannot be empty")
	}

	// 1. Acquire distributed lock
	lockToken, err := lock.AcquireModuleLock(ctx, redisClient, moduleID, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("could not acquire lock: %w", err)
	}
	defer func() {
		_ = lock.ReleaseModuleLock(ctx, redisClient, moduleID, lockToken)
	}()

	// 2. Validate cycle detection
	if len(req.Nodes) > 0 || len(req.Edges) > 0 {
		if err := module.DetectCycles(req.Nodes, req.Edges); err != nil {
			return nil, err
		}
	}

	// 3. Begin DB transaction
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update module metadata if provided
	if req.Name != "" || req.Description != "" {
		if err := repo.UpdateModuleMeta(ctx, tx, moduleID, req.Name, req.Description, userID); err != nil {
			return nil, fmt.Errorf("failed to update module meta: %w", err)
		}
	}

	// Upsert nodes
	if len(req.Nodes) > 0 {
		if err := repo.UpsertNodes(ctx, tx, moduleID, req.Nodes); err != nil {
			return nil, err
		}
	}

	// Upsert edges
	if len(req.Edges) > 0 {
		if err := repo.UpsertEdges(ctx, tx, moduleID, req.Edges); err != nil {
			return nil, err
		}
	}

	// Delete edges
	if len(req.DeletedEdges) > 0 {
		if err := repo.DeleteEdges(ctx, tx, moduleID, req.DeletedEdges); err != nil {
			return nil, err
		}
	}

	// Delete nodes
	if len(req.DeletedNodes) > 0 {
		if err := repo.DeleteNodes(ctx, tx, moduleID, req.DeletedNodes); err != nil {
			return nil, err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 4. Invalidate Redis hot cache
	_ = cache.InvalidateGraphCache(ctx, redisClient, moduleID)

	// 5. Read fresh graph and warm cache
	freshGraph, err := repo.FindGraphByModuleID(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	_ = cache.SetGraphCache(ctx, redisClient, moduleID, freshGraph, 5*time.Minute)

	return freshGraph, nil
}
