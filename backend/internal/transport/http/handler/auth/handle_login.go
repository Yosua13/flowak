package auth

import (
	"errors"
	"net/http"

	"backend/config"
	"backend/db"
	domainAuth "backend/internal/domain/auth"
	repoAuth "backend/internal/repository/postgres/auth"
	"backend/internal/transport/http/response"
	usecaseAuth "backend/internal/usecase/auth"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// HandleLogin processes user login requests and returns JWT credentials with session cookies.
func HandleLogin(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteValidationError(c, "Invalid request body")
		return
	}

	repo := repoAuth.NewPostgresAuthRepository(db.DB)
	uc := usecaseAuth.NewLoginUserUseCase(repo, config.ActiveConfig.JWTSecret)

	out, err := uc.Execute(c.Request.Context(), usecaseAuth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainAuth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		if errors.Is(err, domainAuth.ErrNoActiveOrganization) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No active organization membership"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
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
