package ai

import (
	"database/sql"

	domainAI "backend/internal/domain/ai"
)

// PostgresAIJobRepository implements domainAI.AIJobRepository using PostgreSQL.
type PostgresAIJobRepository struct {
	db *sql.DB
}

// NewAIJobRepository creates an instance of domainAI.AIJobRepository.
func NewAIJobRepository(db *sql.DB) domainAI.AIJobRepository {
	return &PostgresAIJobRepository{db: db}
}
