package auth

import "errors"

// Auth domain sentinel errors.
var (
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrEmailAlreadyRegistered  = errors.New("email address already registered")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidSession          = errors.New("invalid session")
	ErrNoActiveOrganization    = errors.New("no active organization membership")
	ErrTokenBlacklisted        = errors.New("token has been revoked")
	ErrAllFieldsRequired       = errors.New("all fields are required")
	ErrInvalidNameLength       = errors.New("name must be between 2 and 100 characters")
	ErrInvalidEmailFormat      = errors.New("invalid email address format")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters")
	ErrElevatedRoleRestricted  = errors.New("public registration cannot assign PM or elevated privileges")
)
