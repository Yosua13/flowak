package auth

import (
	"context"
	"time"

	domainAuth "backend/internal/domain/auth"
	"backend/middleware"
	"github.com/golang-jwt/jwt/v5"
)

// RefreshInput contains the incoming refresh token cookie.
type RefreshInput struct {
	RefreshToken string
}

// RefreshOutput contains the newly rotated tokens and user info.
type RefreshOutput struct {
	Token            string           `json:"token"`
	RefreshToken     string           `json:"refresh_token,omitempty"`
	User             *domainAuth.User `json:"user"`
	OrganizationID   string           `json:"organization_id"`
	OrganizationRole string           `json:"organization_role"`
}

// RefreshTokenUseCase handles refresh token rotation and new access token generation.
type RefreshTokenUseCase struct {
	repo      domainAuth.AuthRepository
	jwtSecret string
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase.
func NewRefreshTokenUseCase(repo domainAuth.AuthRepository, jwtSecret string) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{repo: repo, jwtSecret: jwtSecret}
}

// Execute validates the old refresh token, rotates it, and returns a new access token.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshInput) (*RefreshOutput, error) {
	if input.RefreshToken == "" {
		return nil, domainAuth.ErrInvalidSession
	}

	oldHash := TokenHash(input.RefreshToken)
	user, sessionID, orgID, orgRole, err := uc.repo.FindSessionByRefreshToken(ctx, oldHash)
	if err != nil {
		return nil, domainAuth.ErrInvalidSession
	}

	newRefresh, err := RandomToken()
	if err != nil {
		return nil, err
	}

	newHash := TokenHash(newRefresh)
	if err := uc.repo.RotateRefreshToken(ctx, sessionID, oldHash, newHash); err != nil {
		return nil, domainAuth.ErrInvalidSession
	}

	tokenID := "tok_" + generateUUID()
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

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &RefreshOutput{
		Token:            accessToken,
		RefreshToken:     newRefresh,
		User:             user,
		OrganizationID:   orgID,
		OrganizationRole: orgRole,
	}, nil
}
