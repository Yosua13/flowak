package module

import (
	"context"
	"fmt"

	"backend/internal/domain/module"
)

// ExecutePublishBaseline creates an immutable versioned baseline snapshot of the module graph.
func ExecutePublishBaseline(
	ctx context.Context,
	repo module.ModuleRepository,
	moduleID string,
	actorID string,
) (*module.ModuleBaseline, error) {
	if moduleID == "" {
		return nil, fmt.Errorf("moduleID cannot be empty")
	}

	baseline, err := repo.PublishBaseline(ctx, moduleID, actorID)
	if err != nil {
		return nil, fmt.Errorf("failed to publish baseline: %w", err)
	}

	return baseline, nil
}
