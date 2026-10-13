package auth

import "database/sql"

// PostgresAuthRepository implements auth.AuthRepository for PostgreSQL.
type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresAuthRepository creates a new PostgreSQL auth repository.
func NewPostgresAuthRepository(db *sql.DB) *PostgresAuthRepository {
	return &PostgresAuthRepository{db: db}
}
