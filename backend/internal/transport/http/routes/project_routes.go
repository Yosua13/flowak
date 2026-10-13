package routes

import (
	"backend/handlers"
	"backend/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterProjectRoutes registers project lifecycle, membership, environments, and contributor management routes.
func RegisterProjectRoutes(r *gin.RouterGroup) {
	// Project CRUD & lifecycle
	r.GET("/projects", handlers.GetProjectsHandler)
	r.POST("/projects", handlers.CreateProjectHandler)
	r.GET("/projects/:id", handlers.GetProjectDetailHandler)
	r.DELETE("/projects/:id", handlers.DeleteProjectHandler)
	r.POST("/projects/:id/restore", handlers.RestoreProjectHandler)

	// Project Members
	r.GET("/projects/:id/members", handlers.GetProjectMembersHandler)
	r.POST("/projects/:id/members", handlers.AddProjectMemberHandler)
	r.DELETE("/projects/:id/members/:userId", handlers.RemoveProjectMemberHandler)

	// Project Events, Derived Views & Environments
	r.GET("/projects/:id/events", handlers.ProjectEventsHandler)
	r.GET("/projects/:id/derived-view-data", handlers.GetDerivedViewDataHandler)
	r.GET("/projects/:id/environments", handlers.ListEnvironmentsHandler)
	r.POST("/projects/:id/environments", handlers.CreateEnvironmentHandler)
	r.PATCH("/environments/:id", handlers.UpdateEnvironmentHandler)
	r.DELETE("/environments/:id", handlers.DeleteEnvironmentHandler)
	r.GET("/environments/:id/variables", handlers.ListEnvironmentVariablesHandler)
	r.POST("/environments/:id/variables", handlers.UpsertEnvironmentVariableHandler)
	r.DELETE("/environments/:id/variables/:variableId", handlers.DeleteEnvironmentVariableHandler)

	// User / Contributor Management
	r.GET("/users", handlers.GetUsersHandler)
	r.POST("/users", middleware.RequireRole("pm"), handlers.PostUsersHandler)
	r.DELETE("/users/:id", middleware.RequireRole("pm"), handlers.DeleteUserHandler)
	r.GET("/users/dashboard-stats", handlers.UserDashboardStatsHandler)
	r.POST("/invitations", handlers.CreateInvitationHandler)
}
