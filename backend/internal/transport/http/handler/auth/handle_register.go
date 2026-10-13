package auth

import (
	"errors"
	"net/http"

	"backend/db"
	domainAuth "backend/internal/domain/auth"
	repoAuth "backend/internal/repository/postgres/auth"
	"backend/internal/transport/http/response"
	usecaseAuth "backend/internal/usecase/auth"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// HandleRegister processes user registration requests.
func HandleRegister(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteValidationError(c, "Invalid request body")
		return
	}

	repo := repoAuth.NewPostgresAuthRepository(db.DB)
	uc := usecaseAuth.NewRegisterUserUseCase(repo)

	out, err := uc.Execute(c.Request.Context(), usecaseAuth.RegisterInput{
		Name:             req.Name,
		Email:            req.Email,
		Password:         req.Password,
		Role:             req.Role,
		OrganizationName: req.OrganizationName,
	})
	if err != nil {
		if errors.Is(err, domainAuth.ErrEmailAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email address already registered"})
			return
		}
		if errors.Is(err, domainAuth.ErrAllFieldsRequired) ||
			errors.Is(err, domainAuth.ErrInvalidNameLength) ||
			errors.Is(err, domainAuth.ErrInvalidEmailFormat) ||
			errors.Is(err, domainAuth.ErrPasswordTooShort) ||
			errors.Is(err, domainAuth.ErrElevatedRoleRestricted) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": out.Message,
		"user_id": out.UserID,
	})
}
