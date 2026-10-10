package auth

import (
	"context"
	"database/sql"
	"time"

	domainAuth "backend/internal/domain/auth"
)

// GetPrimaryOrganization retrieves the user's active organization ID and role.
func (r *PostgresAuthRepository) GetPrimaryOrganization(ctx context.Context, userID string) (string, string, error) {
	var orgID, role string
	query := `SELECT organization_id, role 
	          FROM organization_members 
	          WHERE user_id = $1 AND status = 'active' 
	          ORDER BY joined_at ASC 
	          LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&orgID, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", domainAuth.ErrNoActiveOrganization
		}
		return "", "", err
	}
	return orgID, role, nil
}

// CreateSession persists a new user session with refresh token hash into PostgreSQL.
func (r *PostgresAuthRepository) CreateSession(ctx context.Context, sessionID, userID, orgID, refreshTokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO user_sessions (id, user_id, organization_id, refresh_token_hash, expires_at) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, sessionID, userID, orgID, refreshTokenHash, expiresAt)
	return err
}

// FindSessionByRefreshToken finds an active session by refresh token hash.
func (r *PostgresAuthRepository) FindSessionByRefreshToken(ctx context.Context, tokenHash string) (*domainAuth.User, string, string, string, error) {
	var user domainAuth.User
	var sessionID, orgID, orgRole string

	query := `SELECT s.id, s.organization_id, om.role, u.id, u.name, u.email, u.role, u.created_at 
	          FROM user_sessions s 
	          JOIN users u ON u.id = s.user_id 
	          JOIN organization_members om ON om.organization_id = s.organization_id AND om.user_id = s.user_id AND om.status = 'active' 
	          WHERE s.refresh_token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > CURRENT_TIMESTAMP AND u.status = 'active'`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&sessionID, &orgID, &orgRole,
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", "", "", domainAuth.ErrInvalidSession
		}
		return nil, "", "", "", err
	}
	user.Status = "active"
	return &user, sessionID, orgID, orgRole, nil
}

// RotateRefreshToken updates the refresh token hash of an active session.
func (r *PostgresAuthRepository) RotateRefreshToken(ctx context.Context, sessionID, oldTokenHash, newTokenHash string) error {
	query := `UPDATE user_sessions 
	          SET refresh_token_hash = $1, last_used_at = CURRENT_TIMESTAMP 
	          WHERE id = $2 AND refresh_token_hash = $3 AND revoked_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, newTokenHash, sessionID, oldTokenHash)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return domainAuth.ErrInvalidSession
	}
	return nil
}

// RevokeSessionByRefreshToken revokes a session by setting revoked_at timestamp.
func (r *PostgresAuthRepository) RevokeSessionByRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE user_sessions 
	          SET revoked_at = CURRENT_TIMESTAMP 
	          WHERE refresh_token_hash = $1 AND revoked_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}
