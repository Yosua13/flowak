package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	domainAuth "backend/internal/domain/auth"
	"backend/middleware"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

// LoginInput contains credentials for user authentication.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput contains authenticated session tokens and user details.
type LoginOutput struct {
	Token            string           `json:"token"`
	RefreshToken     string           `json:"refresh_token,omitempty"`
	User             *domainAuth.User `json:"user"`
	OrganizationID   string           `json:"organization_id"`
	OrganizationRole string           `json:"organization_role"`
}

// LoginUserUseCase handles credential verification and session creation.
type LoginUserUseCase struct {
	repo      domainAuth.AuthRepository
	jwtSecret string
}

// NewLoginUserUseCase creates a new LoginUserUseCase.
func NewLoginUserUseCase(repo domainAuth.AuthRepository, jwtSecret string) *LoginUserUseCase {
	return &LoginUserUseCase{repo: repo, jwtSecret: jwtSecret}
}

// Execute authenticates the user, generates session tokens, and returns the output.
func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password

	if email == "" || password == "" {
		return nil, domainAuth.ErrInvalidCredentials
	}

	user, passwordHash, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainAuth.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, domainAuth.ErrInvalidCredentials
	}

	orgID, orgRole, err := uc.repo.GetPrimaryOrganization(ctx, user.ID)
	if err != nil {
		return nil, domainAuth.ErrNoActiveOrganization
	}

	refreshToken, err := RandomToken()
	if err != nil {
		return nil, err
	}

	sessionID := "ses_" + generateUUID()
	tokenID := "tok_" + generateUUID()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	if err := uc.repo.CreateSession(ctx, sessionID, user.ID, orgID, TokenHash(refreshToken), expiresAt); err != nil {
		return nil, err
	}

	claims := &middleware.Claims{
		UserID:         user.ID,
		Email:          user.Email,
		Role:           user.Role,
		OrganizationID: orgID,
		SessionID:      sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:            tokenString,
		RefreshToken:     refreshToken,
		User:             user,
		OrganizationID:   orgID,
		OrganizationRole: orgRole,
	}, nil
}
