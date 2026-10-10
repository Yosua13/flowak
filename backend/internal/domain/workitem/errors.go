package workitem

import "errors"

// Domain sentinel errors for the work item domain.
var (
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrOptimisticLockConflict   = errors.New("work item has changed")
	ErrWorkItemNotFound         = errors.New("work item not found")
	ErrInvalidWorkItemInput     = errors.New("invalid work item input")
	ErrHierarchyCycle           = errors.New("parent would create a hierarchy cycle")
	ErrSelfParent               = errors.New("work item cannot be its own parent")
	ErrConcurrentUpdate         = errors.New("concurrent update conflict")
	ErrProjectNotFound          = errors.New("project not found")
	ErrModuleNotBelongToProject  = errors.New("module does not belong to project")
	ErrNodeNotBelongToProject    = errors.New("node does not belong to project")
	ErrNodeNotBelongToModule     = errors.New("node does not belong to module")
	ErrParentNotSameProject      = errors.New("parent must belong to the same project")
	ErrAssigneeNotMember         = errors.New("assignee must be a project member")
)
