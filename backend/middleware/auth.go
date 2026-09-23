package middleware

import (
	"errors"
	"net/http"
	"strings"

	"backend/config"
	"backend/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const (
	UserContextKey         ContextKey = "user"
	RoleContextKey         ContextKey = "role"
	OrganizationContextKey ContextKey = "organization"
	ProjectRoleContextKey  ContextKey = "project_role"
)

type Claims struct {
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	OrganizationID string `json:"organization_id"`
	SessionID      string `json:"session_id"`
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
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(config.ActiveConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(string(UserContextKey), claims.UserID)
		c.Set(string(RoleContextKey), claims.Role)
		organizationID := strings.TrimSpace(c.GetHeader("X-Organization-ID"))
		if organizationID == "" {
			organizationID = claims.OrganizationID
		}
		if organizationID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization context is required"})
			c.Abort()
			return
		}
		var member bool
		if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM organization_members WHERE organization_id = $1 AND user_id = $2 AND status = 'active')`, organizationID, claims.UserID).Scan(&member); err != nil || !member {
			c.JSON(http.StatusForbidden, gin.H{"error": "Organization access denied"})
			c.Abort()
			return
		}
		var activeSession bool
		if err := db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id = $1 AND user_id = $2 AND organization_id = $3 AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP)`, claims.SessionID, claims.UserID, organizationID).Scan(&activeSession); err != nil || !activeSession {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session is invalid or expired"})
			c.Abort()
			return
		}
		c.Set(string(OrganizationContextKey), organizationID)
		c.Next()
	}
}

// ProjectCapability is the single authorization vocabulary for project resources.
type ProjectCapability string

const (
	CapabilityView      ProjectCapability = "view"
	CapabilityComment   ProjectCapability = "comment"
	CapabilityWorkItem  ProjectCapability = "work_item"
	CapabilityEditGraph ProjectCapability = "edit_graph"
	CapabilityManage    ProjectCapability = "manage"
)

// AuthorizeProject is used by handlers with a project id from a route or a
// resource lookup. A missing membership is deliberately indistinguishable from
// a foreign resource to prevent cross-tenant ID probing.
func AuthorizeProject(c *gin.Context, projectID string, capability ProjectCapability) bool {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}
	organizationID, ok := c.Get(string(OrganizationContextKey))
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization access denied"})
		return false
	}
	var projectRole string
	err = db.DB.QueryRow(`
		SELECT COALESCE(pm.project_role, CASE WHEN p.owner_id = $1 THEN 'owner' END)
		FROM projects p
		JOIN organization_members om ON om.organization_id = p.organization_id
		LEFT JOIN project_members pm ON pm.project_id = p.id AND pm.user_id = $1
		WHERE p.id = $2 AND p.organization_id = $3 AND p.status = 'active'
		  AND om.user_id = $1 AND om.status = 'active'
		  AND (p.owner_id = $1 OR pm.user_id IS NOT NULL)
		LIMIT 1`, userID, projectID, organizationID.(string)).Scan(&projectRole)
	if err != nil || !can(projectRole, capability) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project access denied"})
		return false
	}
	c.Set(string(ProjectRoleContextKey), projectRole)
	return true
}

func can(role string, capability ProjectCapability) bool {
	switch capability {
	case CapabilityView:
		return role == "owner" || role == "editor" || role == "commenter" || role == "viewer"
	case CapabilityComment:
		return role == "owner" || role == "editor" || role == "commenter"
	case CapabilityWorkItem:
		return role == "owner" || role == "editor" || role == "commenter"
	case CapabilityEditGraph:
		return role == "owner" || role == "editor"
	case CapabilityManage:
		return role == "owner"
	default:
		return false
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
