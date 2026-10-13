package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	domainAuth "backend/internal/domain/auth"
)

func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// CreateUser persists a new user and establishes their initial organization membership in PostgreSQL.
func (r *PostgresAuthRepository) CreateUser(ctx context.Context, user *domainAuth.User, passwordHash, organizationName string) (string, error) {
	// 1. Check if email already exists
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
	if err != nil {
		return "", err
	}
	if exists {
		return "", domainAuth.ErrEmailAlreadyRegistered
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if user.ID == "" {
		user.ID = "usr_" + generateUUID()
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)",
		user.ID, user.Name, user.Email, passwordHash, user.Role,
	)
	if err != nil {
		return "", err
	}

	organizationID := "org_default"
	organizationRole := "member"
	if orgName := strings.TrimSpace(organizationName); orgName != "" {
		organizationID = "org_" + generateUUID()
		organizationRole = "owner"
		slug := strings.ToLower(strings.ReplaceAll(organizationID, "_", "-"))
		if _, err = tx.ExecContext(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`, organizationID, orgName, slug); err != nil {
			return "", err
		}
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO organization_members (organization_id, user_id, role, status) VALUES ($1, $2, $3, 'active')`, organizationID, user.ID, organizationRole); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return organizationID, nil
}
