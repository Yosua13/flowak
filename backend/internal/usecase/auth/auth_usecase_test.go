package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainAuth "backend/internal/domain/auth"
	usecaseAuth "backend/internal/usecase/auth"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	findByEmailFunc            func(ctx context.Context, email string) (*domainAuth.User, string, error)
	createUserFunc             func(ctx context.Context, user *domainAuth.User, passwordHash, orgName string) (string, error)
	findByIDFunc               func(ctx context.Context, id string) (*domainAuth.User, error)
	getPrimaryOrganizationFunc func(ctx context.Context, userID string) (string, string, error)
	createSessionFunc          func(ctx context.Context, sessionID, userID, orgID, refreshTokenHash string, expiresAt time.Time) error
	findSessionFunc            func(ctx context.Context, tokenHash string) (*domainAuth.User, string, string, string, error)
	rotateRefreshTokenFunc     func(ctx context.Context, sessionID, oldTokenHash, newTokenHash string) error
	revokeSessionFunc          func(ctx context.Context, tokenHash string) error
}

func (m *mockAuthRepo) FindByEmail(ctx context.Context, email string) (*domainAuth.User, string, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, "", domainAuth.ErrUserNotFound
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *domainAuth.User, passwordHash, orgName string) (string, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user, passwordHash, orgName)
	}
	return "org_default", nil
}

func (m *mockAuthRepo) FindByID(ctx context.Context, id string) (*domainAuth.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, domainAuth.ErrUserNotFound
}

func (m *mockAuthRepo) GetPrimaryOrganization(ctx context.Context, userID string) (string, string, error) {
	if m.getPrimaryOrganizationFunc != nil {
		return m.getPrimaryOrganizationFunc(ctx, userID)
	}
	return "org_default", "owner", nil
}

func (m *mockAuthRepo) CreateSession(ctx context.Context, sessionID, userID, orgID, refreshTokenHash string, expiresAt time.Time) error {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, sessionID, userID, orgID, refreshTokenHash, expiresAt)
	}
	return nil
}

func (m *mockAuthRepo) FindSessionByRefreshToken(ctx context.Context, tokenHash string) (*domainAuth.User, string, string, string, error) {
	if m.findSessionFunc != nil {
		return m.findSessionFunc(ctx, tokenHash)
	}
	return nil, "", "", "", domainAuth.ErrInvalidSession
}

func (m *mockAuthRepo) RotateRefreshToken(ctx context.Context, sessionID, oldTokenHash, newTokenHash string) error {
	if m.rotateRefreshTokenFunc != nil {
		return m.rotateRefreshTokenFunc(ctx, sessionID, oldTokenHash, newTokenHash)
	}
	return nil
}

func (m *mockAuthRepo) RevokeSessionByRefreshToken(ctx context.Context, tokenHash string) error {
	if m.revokeSessionFunc != nil {
		return m.revokeSessionFunc(ctx, tokenHash)
	}
	return nil
}

type mockBlacklistRepo struct {
	blacklistedTokens map[string]bool
}

func (m *mockBlacklistRepo) BlacklistToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if m.blacklistedTokens == nil {
		m.blacklistedTokens = make(map[string]bool)
	}
	m.blacklistedTokens[tokenID] = true
	return nil
}

func (m *mockBlacklistRepo) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if m.blacklistedTokens == nil {
		return false, nil
	}
	return m.blacklistedTokens[tokenID], nil
}

func TestRegisterUserUseCase(t *testing.T) {
	repo := &mockAuthRepo{
		createUserFunc: func(ctx context.Context, user *domainAuth.User, passwordHash, orgName string) (string, error) {
			user.ID = "usr_100"
			return "org_100", nil
		},
	}
	uc := usecaseAuth.NewRegisterUserUseCase(repo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		out, err := uc.Execute(ctx, usecaseAuth.RegisterInput{
			Name:     "Valid User",
			Email:    "valid@flowak.com",
			Password: "securePassword123",
			Role:     "frontend",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.UserID != "usr_100" || out.OrganizationID != "org_100" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		_, err := uc.Execute(ctx, usecaseAuth.RegisterInput{
			Name:     "Valid User",
			Email:    "invalid-email",
			Password: "securePassword123",
		})
		if !errors.Is(err, domainAuth.ErrInvalidEmailFormat) {
			t.Errorf("expected ErrInvalidEmailFormat, got: %v", err)
		}
	})

	t.Run("short password", func(t *testing.T) {
		_, err := uc.Execute(ctx, usecaseAuth.RegisterInput{
			Name:     "Valid User",
			Email:    "valid@flowak.com",
			Password: "short",
		})
		if !errors.Is(err, domainAuth.ErrPasswordTooShort) {
			t.Errorf("expected ErrPasswordTooShort, got: %v", err)
		}
	})
}

func TestLoginUserUseCase(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctPassword123"), bcrypt.DefaultCost)
	repo := &mockAuthRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*domainAuth.User, string, error) {
			if email == "user@flowak.com" {
				return &domainAuth.User{ID: "usr_1", Name: "Test", Email: email, Role: "frontend"}, string(hashed), nil
			}
			return nil, "", domainAuth.ErrUserNotFound
		},
	}
	uc := usecaseAuth.NewLoginUserUseCase(repo, "secret123")
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		out, err := uc.Execute(ctx, usecaseAuth.LoginInput{
			Email:    "user@flowak.com",
			Password: "correctPassword123",
		})
		if err != nil {
			t.Fatalf("unexpected login error: %v", err)
		}
		if out.Token == "" || out.RefreshToken == "" {
			t.Errorf("expected non-empty tokens in output: %+v", out)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		_, err := uc.Execute(ctx, usecaseAuth.LoginInput{
			Email:    "user@flowak.com",
			Password: "wrongPassword",
		})
		if !errors.Is(err, domainAuth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got: %v", err)
		}
	})
}

func TestLogoutUserUseCase(t *testing.T) {
	var revokedHash string
	repo := &mockAuthRepo{
		revokeSessionFunc: func(ctx context.Context, tokenHash string) error {
			revokedHash = tokenHash
			return nil
		},
	}
	blacklistRepo := &mockBlacklistRepo{}
	uc := usecaseAuth.NewLogoutUserUseCase(repo, blacklistRepo)
	ctx := context.Background()

	err := uc.Execute(ctx, usecaseAuth.LogoutInput{
		TokenID:      "tok_123",
		RefreshToken: "refresh_abc",
		RemainingTTL: 10 * time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected logout error: %v", err)
	}

	if revokedHash == "" {
		t.Errorf("expected refresh token to be revoked")
	}

	isBlacklisted, _ := blacklistRepo.IsTokenBlacklisted(ctx, "tok_123")
	if !isBlacklisted {
		t.Errorf("expected token tok_123 to be blacklisted in Redis")
	}
}

func TestRefreshTokenUseCase(t *testing.T) {
	repo := &mockAuthRepo{
		findSessionFunc: func(ctx context.Context, tokenHash string) (*domainAuth.User, string, string, string, error) {
			if tokenHash == usecaseAuth.TokenHash("valid_refresh") {
				return &domainAuth.User{ID: "usr_1", Email: "test@flowak.com", Role: "frontend"}, "ses_1", "org_1", "member", nil
			}
			return nil, "", "", "", domainAuth.ErrInvalidSession
		},
		rotateRefreshTokenFunc: func(ctx context.Context, sessionID, oldTokenHash, newTokenHash string) error {
			return nil
		},
	}
	uc := usecaseAuth.NewRefreshTokenUseCase(repo, "secret123")
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		out, err := uc.Execute(ctx, usecaseAuth.RefreshInput{RefreshToken: "valid_refresh"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Token == "" || out.RefreshToken == "" {
			t.Errorf("expected new tokens")
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, err := uc.Execute(ctx, usecaseAuth.RefreshInput{RefreshToken: "bad_refresh"})
		if !errors.Is(err, domainAuth.ErrInvalidSession) {
			t.Errorf("expected ErrInvalidSession, got: %v", err)
		}
	})
}
