package auth

import (
	"context"
	"database/sql"

	domainAuth "backend/internal/domain/auth"
)

// FindByID retrieves a user domain entity by user ID from PostgreSQL.
func (r *PostgresAuthRepository) FindByID(ctx context.Context, id string) (*domainAuth.User, error) {
	var user domainAuth.User

	query := `SELECT id, name, email, role, status, created_at 
	          FROM users 
	          WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domainAuth.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
