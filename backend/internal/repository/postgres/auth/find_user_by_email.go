package auth

import (
	"context"
	"database/sql"

	domainAuth "backend/internal/domain/auth"
)

// FindByEmail queries a user and their hashed password by email from PostgreSQL.
func (r *PostgresAuthRepository) FindByEmail(ctx context.Context, email string) (*domainAuth.User, string, error) {
	var user domainAuth.User
	var passwordHash string

	query := `SELECT id, name, email, password_hash, role, created_at 
	          FROM users 
	          WHERE email = $1 AND status = 'active'`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&passwordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", domainAuth.ErrUserNotFound
		}
		return nil, "", err
	}

	user.Status = "active"
	return &user, passwordHash, nil
}
