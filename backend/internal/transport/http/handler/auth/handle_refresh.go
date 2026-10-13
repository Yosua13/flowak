package auth

import (
	"errors"
	"net/http"

	"backend/config"
	"backend/db"
	domainAuth "backend/internal/domain/auth"
	repoAuth "backend/internal/repository/postgres/auth"
	usecaseAuth "backend/internal/usecase/auth"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// HandleRefresh rotates the refresh token and issues a new access token.
func HandleRefresh(c *gin.Context) {
	refreshToken, err := c.Cookie("flowak_refresh")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	repo := repoAuth.NewPostgresAuthRepository(db.DB)
	uc := usecaseAuth.NewRefreshTokenUseCase(repo, config.ActiveConfig.JWTSecret)

	out, err := uc.Execute(c.Request.Context(), usecaseAuth.RefreshInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		if errors.Is(err, domainAuth.ErrInvalidSession) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
		return
	}

	SetAuthCookies(c, out.RefreshToken, out.Token)

	userModel := models.User{
		ID:        out.User.ID,
		Name:      out.User.Name,
		Email:     out.User.Email,
		Role:      out.User.Role,
		CreatedAt: out.User.CreatedAt,
	}

	c.JSON(http.StatusOK, models.UserLoginResponse{
		Token:            out.Token,
		User:             userModel,
		OrganizationID:   out.OrganizationID,
		OrganizationRole: out.OrganizationRole,
	})
}
