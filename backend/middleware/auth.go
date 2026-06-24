package middleware

import (
	"errors"
	"net/http"
	"strings"

	"backend/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	UserContextKey ContextKey = "user"
	RoleContextKey ContextKey = "role"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT tokens and sets user claims in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.ActiveConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(string(UserContextKey), claims.UserID)
		c.Set(string(RoleContextKey), claims.Role)
		c.Next()
	}
}

// RequireRole restricts access to specific roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleVal, exists := c.Get(string(RoleContextKey))
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in context"})
			c.Abort()
			return
		}
		userRole, ok := userRoleVal.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in context"})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to perform this action"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID helper retrieves user ID from request context
func GetUserID(c *gin.Context) (string, error) {
	userIDVal, exists := c.Get(string(UserContextKey))
	if !exists {
		return "", errors.New("unauthorized")
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		return "", errors.New("unauthorized")
	}
	return userID, nil
}

// GetUserRole helper retrieves user role from request context
func GetUserRole(c *gin.Context) (string, error) {
	roleVal, exists := c.Get(string(RoleContextKey))
	if !exists {
		return "", errors.New("role not found")
	}
	role, ok := roleVal.(string)
	if !ok || role == "" {
		return "", errors.New("role not found")
	}
	return role, nil
}
