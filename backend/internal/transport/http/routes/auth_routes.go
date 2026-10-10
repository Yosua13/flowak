package routes

import (
	"backend/handlers"
	authHandler "backend/internal/transport/http/handler/auth"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers all public authentication and onboarding routes on the router group.
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/auth/register", authHandler.HandleRegister)
	r.POST("/auth/login", authHandler.HandleLogin)
	r.POST("/auth/refresh", authHandler.HandleRefresh)
	r.POST("/auth/logout", authHandler.HandleLogout)
	r.POST("/auth/password-reset/request", handlers.PasswordResetRequestHandler)
	r.POST("/invitations/accept", handlers.AcceptInvitationHandler)
}
