package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/config"
	"backend/db"
	"backend/handlers"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware sets up simple CORS headers for Gin context
func CorsMiddleware() gin.HandlerFunc {
	allowedOrigins := make(map[string]bool)
	for _, origin := range config.ActiveConfig.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			if origin != "" && !allowedOrigins[origin] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func main() {
	// 1. Initialize config & environment
	config.InitConfig()
	log.Printf("Starting Flowak Backend on port %s...", config.ActiveConfig.Port)

	// 2. Initialize database
	db.InitDB()
	defer db.DB.Close()

	// 3. Setup router (Gin Engine)
	r := gin.Default()

	// Global Middleware
	r.Use(CorsMiddleware())

	// Public routes
	r.POST("/api/auth/register", handlers.RegisterHandler)
	r.POST("/api/auth/login", handlers.LoginHandler)
	r.POST("/api/auth/refresh", handlers.RefreshHandler)
	r.POST("/api/auth/logout", handlers.LogoutHandler)
	r.POST("/api/auth/password-reset/request", handlers.PasswordResetRequestHandler)
	r.POST("/api/invitations/accept", handlers.AcceptInvitationHandler)

	// Protected routes group
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Project CRUD & sub-routes
		api.GET("/projects", handlers.GetProjectsHandler)
		api.POST("/projects", handlers.CreateProjectHandler)
		api.GET("/projects/:id", handlers.GetProjectDetailHandler)
		api.DELETE("/projects/:id", handlers.DeleteProjectHandler)
		api.POST("/projects/:id/restore", handlers.RestoreProjectHandler)

		// Project Members
		api.GET("/projects/:id/members", handlers.GetProjectMembersHandler)
		api.POST("/projects/:id/members", handlers.AddProjectMemberHandler)
		api.DELETE("/projects/:id/members/:userId", handlers.RemoveProjectMemberHandler)

		// Module CRUD
		api.GET("/projects/:id/modules", handlers.GetProjectModulesHandler)
		api.POST("/projects/:id/modules", handlers.CreateProjectModuleHandler)
		api.PUT("/modules/:id", handlers.UpdateModuleHandler)
		api.DELETE("/modules/:id", handlers.DeleteModuleHandler)

		// Work items, comments, and execution history
		api.GET("/projects/:id/work-items", handlers.ListWorkItemsHandler)
		api.GET("/projects/:id/events", handlers.ProjectEventsHandler)
		api.POST("/modules/:id/publish", handlers.PublishModuleBaselineHandler)
		api.POST("/modules/:id/versions/:version/restore", handlers.RestoreModuleBaselineHandler)
		api.GET("/notifications", handlers.ListNotificationsHandler)
		api.POST("/notifications/read", handlers.MarkNotificationsReadHandler)
		api.GET("/projects/:id/derived-view-data", handlers.GetDerivedViewDataHandler)
		api.POST("/projects/:id/work-items", handlers.CreateWorkItemHandler)
		api.GET("/work-items/:key", handlers.GetWorkItemHandler)
		api.PATCH("/work-items/:key", handlers.UpdateWorkItemHandler)
		api.POST("/work-items/:key/transitions", handlers.TransitionWorkItemHandler)
		api.POST("/work-items/:key/comments", handlers.CreateWorkItemCommentHandler)
		api.GET("/work-items/:key/comments", handlers.ListWorkItemCommentsHandler)
		api.GET("/nodes/:id/comments", handlers.ListNodeCommentsHandler)
		api.POST("/nodes/:id/comments", handlers.CreateNodeCommentHandler)
		api.PATCH("/comments/:id", handlers.UpdateCommentHandler)

		// Server-side only API contract execution; browser never calls targets directly.
		api.POST("/api-requests/:id/runs", handlers.RunAPIRequestHandler)

		// AI Proxies
		api.POST("/ai/generate-flow", handlers.AiGenerateFlowHandler)
		api.POST("/ai/audit-flow", handlers.AiAuditFlowHandler)

		// User / Contributor CRUD
		api.GET("/users", handlers.GetUsersHandler)
		api.POST("/users", middleware.RequireRole("pm"), handlers.PostUsersHandler)
		api.DELETE("/users/:id", middleware.RequireRole("pm"), handlers.DeleteUserHandler)
		api.GET("/users/dashboard-stats", handlers.UserDashboardStatsHandler)
		api.POST("/invitations", handlers.CreateInvitationHandler)
	}

	// Catch-all route to serve compiled static assets from the frontend/dist folder (Production)
	distDir := "../frontend/dist"
	r.NoRoute(func(c *gin.Context) {
		// Clean the path to prevent directory traversal vulnerabilities
		cleanedPath := filepath.Clean(c.Request.URL.Path)
		filePath := filepath.Join(distDir, cleanedPath)

		// Security: Prevent directory traversal escaping the distDir root
		absDist, _ := filepath.Abs(distDir)
		absFile, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFile, absDist) {
			c.String(http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		// Check if file exists and is not a directory
		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			// Fallback to index.html (SPA routing)
			c.File(filepath.Join(distDir, "index.html"))
			return
		}

		// Otherwise serve the static asset
		c.File(filePath)
	})

	// 4. Start HTTP Server
	serverAddr := ":" + config.ActiveConfig.Port
	log.Printf("Server listening on http://localhost%s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
