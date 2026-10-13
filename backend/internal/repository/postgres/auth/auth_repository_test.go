package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	domainAuth "backend/internal/domain/auth"
	repoAuth "backend/internal/repository/postgres/auth"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := repoAuth.NewPostgresAuthRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at"}).
			AddRow("usr_1", "Test User", "test@flowak.com", "hashed_pw", "frontend", time.Now())

		mock.ExpectQuery("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = \\$1").
			WithArgs("test@flowak.com").
			WillReturnRows(rows)

		user, pwHash, err := repo.FindByEmail(ctx, "test@flowak.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != "usr_1" || pwHash != "hashed_pw" {
			t.Errorf("unexpected user data: %+v, pw: %s", user, pwHash)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = \\$1").
			WithArgs("missing@flowak.com").
			WillReturnError(sql.ErrNoRows)

		_, _, err := repo.FindByEmail(ctx, "missing@flowak.com")
		if !errors.Is(err, domainAuth.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})
}

func TestFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := repoAuth.NewPostgresAuthRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "email", "role", "status", "created_at"}).
			AddRow("usr_1", "Test User", "test@flowak.com", "frontend", "active", time.Now())

		mock.ExpectQuery("SELECT id, name, email, role, status, created_at FROM users WHERE id = \\$1").
			WithArgs("usr_1").
			WillReturnRows(rows)

		user, err := repo.FindByID(ctx, "usr_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != "usr_1" {
			t.Errorf("expected user ID usr_1, got: %s", user.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name, email, role, status, created_at FROM users WHERE id = \\$1").
			WithArgs("usr_unknown").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindByID(ctx, "usr_unknown")
		if !errors.Is(err, domainAuth.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := repoAuth.NewPostgresAuthRepository(db)
	ctx := context.Background()

	t.Run("email already exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE email = \\$1\\)").
			WithArgs("existing@flowak.com").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		user := &domainAuth.User{Email: "existing@flowak.com"}
		_, err := repo.CreateUser(ctx, user, "hash", "")
		if !errors.Is(err, domainAuth.ErrEmailAlreadyRegistered) {
			t.Errorf("expected ErrEmailAlreadyRegistered, got: %v", err)
		}
	})

	t.Run("success with default org", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE email = \\$1\\)").
			WithArgs("new@flowak.com").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO users").
			WithArgs(sqlmock.AnyArg(), "New User", "new@flowak.com", "hash", "frontend").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO organization_members").
			WithArgs("org_default", sqlmock.AnyArg(), "member").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		user := &domainAuth.User{Name: "New User", Email: "new@flowak.com", Role: "frontend"}
		orgID, err := repo.CreateUser(ctx, user, "hash", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if orgID != "org_default" {
			t.Errorf("expected org_default, got: %s", orgID)
		}
	})
}
