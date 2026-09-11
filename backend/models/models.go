package models

import "time"

// User represents the system user
type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // pm, uiux, frontend, backend
	CreatedAt    time.Time `json:"created_at"`
}

// UserRegisterRequest is the payload to register a new user
type UserRegisterRequest struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Role             string `json:"role,omitempty"` // Functional role only; PM is invite/bootstrap-only.
	OrganizationName string `json:"organization_name,omitempty"`
}

// UserLoginRequest is the payload to login
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginResponse is returned upon successful authentication
type UserLoginResponse struct {
	Token          string `json:"token"`
	User           User   `json:"user"`
	OrganizationID string `json:"organization_id"`
}

type InvitationRequest struct {
	Email          string `json:"email"`
	ProjectID      string `json:"project_id,omitempty"`
	ProjectRole    string `json:"project_role"`
	FunctionalRole string `json:"functional_role,omitempty"`
}

type InvitationAcceptRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Project represents a project container holding multiple modules
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProjectCreateRequest is the payload to create a new project
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Module matches the frontend structure, but stored under a project
type Module struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Nodes         string    `json:"nodes"` // JSON text representing Node[]
	Edges         string    `json:"edges"` // JSON text representing Edge[]
	SchemaVersion int       `json:"schemaVersion"`
	CreatedAt     time.Time `json:"created_at"`
}

// ModuleRequest is the payload to update or create a module
type ModuleRequest struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Nodes         any           `json:"nodes,omitempty"` // Raw JSON array
	Edges         any           `json:"edges,omitempty"` // Raw JSON array
	DeletedNodes  []GraphDelete `json:"deletedNodes,omitempty"`
	DeletedEdges  []GraphDelete `json:"deletedEdges,omitempty"`
	SchemaVersion int           `json:"schemaVersion,omitempty"`
}

// GraphDelete identifies a normalized graph record to tombstone. RowVersion is
// required for existing records so a stale client cannot delete a newer revision.
type GraphDelete struct {
	ID         string `json:"id"`
	RowVersion int    `json:"rowVersion"`
}

// WorkItem is an execution unit and is deliberately separate from graph readiness facets.
type WorkItem struct {
	ID            string     `json:"id"`
	Key           string     `json:"key"`
	ProjectID     string     `json:"project_id"`
	ModuleID      *string    `json:"module_id,omitempty"`
	NodeID        *string    `json:"node_id,omitempty"`
	FacetKey      *string    `json:"facet_key,omitempty"`
	ParentID      *string    `json:"parent_id,omitempty"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	Priority      string     `json:"priority"`
	Points        *int       `json:"points,omitempty"`
	Status        string     `json:"status"`
	AssigneeID    *string    `json:"assignee_id,omitempty"`
	ReporterID    string     `json:"reporter_id"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	BlockedReason *string    `json:"blocked_reason,omitempty"`
	Resolution    *string    `json:"resolution,omitempty"`
	RowVersion    int        `json:"row_version"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type WorkItemRequest struct {
	ModuleID      *string `json:"module_id,omitempty"`
	NodeID        *string `json:"node_id,omitempty"`
	FacetKey      *string `json:"facet_key,omitempty"`
	ParentID      *string `json:"parent_id,omitempty"`
	Type          string  `json:"type"`
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Priority      string  `json:"priority,omitempty"`
	Points        *int    `json:"points,omitempty"`
	Status        string  `json:"status,omitempty"`
	AssigneeID    *string `json:"assignee_id,omitempty"`
	StartDate     *string `json:"start_date,omitempty"`
	DueDate       *string `json:"due_date,omitempty"`
	BlockedReason *string `json:"blocked_reason,omitempty"`
	Resolution    *string `json:"resolution,omitempty"`
	RowVersion    int     `json:"row_version"`
}

type WorkItemTransitionRequest struct {
	Status     string `json:"status"`
	Note       string `json:"note,omitempty"`
	Resolution string `json:"resolution,omitempty"`
	RowVersion int    `json:"row_version"`
}

type CommentRequest struct {
	Body     string   `json:"body"`
	ParentID *string  `json:"parent_id,omitempty"`
	Mentions []string `json:"mentions,omitempty"`
}

type CommentUpdateRequest struct {
	Body     *string `json:"body,omitempty"`
	Resolved *bool   `json:"resolved,omitempty"`
}
