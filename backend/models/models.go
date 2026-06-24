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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UserLoginRequest is the payload to login
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginResponse is returned upon successful authentication
type UserLoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
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
	Name          string `json:"name"`
	Description   string `json:"description"`
	Nodes         any    `json:"nodes,omitempty"`         // Raw JSON array
	Edges         any    `json:"edges,omitempty"`         // Raw JSON array
	SchemaVersion int    `json:"schemaVersion,omitempty"`
}
