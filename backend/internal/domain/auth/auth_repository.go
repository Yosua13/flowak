package auth

import (
	"context"
	"time"
)

// AuthRepository defines the contract for user persistence and session management.
type AuthRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, string, error)
	CreateUser(ctx context.Context, user *User, passwordHash, organizationName string) (string, error)
	FindByID(ctx context.Context, id string) (*User, error)
	GetPrimaryOrganization(ctx context.Context, userID string) (string, string, error)
	CreateSession(ctx context.Context, sessionID, userID, orgID, refreshTokenHash string, expiresAt time.Time) error
	FindSessionByRefreshToken(ctx context.Context, tokenHash string) (*User, string, string, string, error)
	RotateRefreshToken(ctx context.Context, sessionID, oldTokenHash, newTokenHash string) error
	RevokeSessionByRefreshToken(ctx context.Context, tokenHash string) error
}
