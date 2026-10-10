package module

import "errors"

var (
	// ErrCycleDetected indicates that the graph contains a circular loop.
	ErrCycleDetected = errors.New("cycle detected in module graph")

	// ErrGraphConflict indicates an optimistic concurrency version conflict on node or edge.
	ErrGraphConflict = errors.New("graph version conflict")

	// ErrInvalidGraphReference indicates an invalid source, target, or foreign node reference.
	ErrInvalidGraphReference = errors.New("invalid graph reference")

	// ErrModuleNotFound indicates the requested module was not found.
	ErrModuleNotFound = errors.New("module not found")

	// ErrLockAcquisitionFailed indicates inability to acquire distributed lock for graph sync.
	ErrLockAcquisitionFailed = errors.New("failed to acquire distributed lock for module sync")

	// ErrInvalidGraphPayload indicates malformed graph nodes or edges payload.
	ErrInvalidGraphPayload = errors.New("invalid graph payload")
)
