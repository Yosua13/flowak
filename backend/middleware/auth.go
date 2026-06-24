package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backend/config"
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
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header is required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Authorization header must be Bearer {token}"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.ActiveConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleContextKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole restricts access to specific roles
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleContextKey).(string)
			if !ok {
				http.Error(w, `{"error": "Role not found in context"}`, http.StatusForbidden)
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
				http.Error(w, `{"error": "You do not have permission to perform this action"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID helper retrieves user ID from request context
func GetUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(UserContextKey).(string)
	if !ok || userID == "" {
		return "", errors.New("unauthorized")
	}
	return userID, nil
}

// GetUserRole helper retrieves user role from request context
func GetUserRole(r *http.Request) (string, error) {
	role, ok := r.Context().Value(RoleContextKey).(string)
	if !ok || role == "" {
		return "", errors.New("role not found")
	}
	return role, nil
}
