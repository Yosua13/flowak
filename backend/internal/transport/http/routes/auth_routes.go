package routes

import (
	"backend/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers all public authentication and onboarding routes on the router group.
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/register", handlers.RegisterHandler)
	r.POST("/auth/login", handlers.LoginHandler)
	r.POST("/auth/refresh", handlers.RefreshHandler)
	r.POST("/auth/logout", handlers.LogoutHandler)
	r.POST("/auth/password-reset/request", handlers.PasswordResetRequestHandler)
	r.POST("/invitations/accept", handlers.AcceptInvitationHandler)
}
