package auth

import (
	"context"
	"regexp"
	"strings"

	domainAuth "backend/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// RegisterInput contains input parameters for registering a new user.
type RegisterInput struct {
	Name             string
	Email            string
	Password         string
	Role             string
	OrganizationName string
}

// RegisterOutput contains the result of a successful user registration.
type RegisterOutput struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Message        string `json:"message"`
}

// RegisterUserUseCase handles user registration business logic.
type RegisterUserUseCase struct {
	repo domainAuth.AuthRepository
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase.
func NewRegisterUserUseCase(repo domainAuth.AuthRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo}
}

// Execute performs validation, password hashing, and user creation.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password
	role := strings.TrimSpace(input.Role)

	if name == "" || email == "" || password == "" {
		return nil, domainAuth.ErrAllFieldsRequired
	}
	if len(name) < 2 || len(name) > 100 {
		return nil, domainAuth.ErrInvalidNameLength
	}
	if !emailRegex.MatchString(email) {
		return nil, domainAuth.ErrInvalidEmailFormat
	}
	if len(password) < 8 {
		return nil, domainAuth.ErrPasswordTooShort
	}

	if role == "" {
		role = "frontend"
	}
	if role != "uiux" && role != "frontend" && role != "backend" {
		return nil, domainAuth.ErrElevatedRoleRestricted
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domainAuth.User{
		Name:   name,
		Email:  email,
		Role:   role,
		Status: "active",
	}

	orgID, err := uc.repo.CreateUser(ctx, user, string(hashedPassword), input.OrganizationName)
	if err != nil {
		return nil, err
	}

	return &RegisterOutput{
		UserID:         user.ID,
		OrganizationID: orgID,
		Message:        "User registered successfully",
	}, nil
}
