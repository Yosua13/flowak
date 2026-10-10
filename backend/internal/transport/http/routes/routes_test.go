package routes_test

import (
	"testing"

	"backend/internal/transport/http/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegisterAllRoutes(t *testing.T) {
	r := gin.New()

	public := r.Group("/api")
	routes.RegisterAuthRoutes(public)

	api := r.Group("/api")
	routes.RegisterProjectRoutes(api)
	routes.RegisterModuleRoutes(api)
	routes.RegisterWorkItemRoutes(api)
	routes.RegisterAIRoutes(api)

	registeredRoutes := make(map[string]bool)
	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		registeredRoutes[key] = true
	}

	expectedRoutes := []string{
		// Auth routes
		"POST /api/auth/register",
		"POST /api/auth/login",
		"POST /api/auth/refresh",
		"POST /api/auth/logout",
		"POST /api/auth/password-reset/request",
		"POST /api/invitations/accept",

		// Project routes
		"GET /api/projects",
		"POST /api/projects",
		"GET /api/projects/:id",
		"DELETE /api/projects/:id",
		"POST /api/projects/:id/restore",
		"GET /api/projects/:id/members",
		"POST /api/projects/:id/members",
		"DELETE /api/projects/:id/members/:userId",
		"GET /api/projects/:id/events",
		"GET /api/projects/:id/derived-view-data",
		"GET /api/projects/:id/environments",
		"POST /api/projects/:id/environments",
		"PATCH /api/environments/:id",
		"DELETE /api/environments/:id",
		"GET /api/environments/:id/variables",
		"POST /api/environments/:id/variables",
		"DELETE /api/environments/:id/variables/:variableId",
		"GET /api/users",
		"POST /api/users",
		"DELETE /api/users/:id",
		"GET /api/users/dashboard-stats",
		"POST /api/invitations",

		// Module routes
		"GET /api/projects/:id/modules",
		"POST /api/projects/:id/modules",
		"PUT /api/modules/:id",
		"DELETE /api/modules/:id",
		"POST /api/modules/:id/publish",
		"GET /api/modules/:id/versions",
		"POST /api/modules/:id/versions/:version/restore",

		// Work item routes
		"GET /api/projects/:id/work-items",
		"POST /api/projects/:id/work-items",
		"GET /api/work-items/:key",
		"PATCH /api/work-items/:key",
		"POST /api/work-items/:key/transitions",
		"GET /api/work-items/:key/activity",
		"GET /api/work-items/:key/artifacts",
		"POST /api/work-items/:key/artifacts/:kind",
		"PATCH /api/work-items/:key/artifacts/:kind/:artifactId",
		"DELETE /api/work-items/:key/artifacts/:kind/:artifactId",
		"POST /api/work-items/:key/comments",
		"GET /api/work-items/:key/comments",
		"GET /api/nodes/:id/comments",
		"POST /api/nodes/:id/comments",
		"GET /api/nodes/:id/activity",
		"PATCH /api/comments/:id",
		"GET /api/notifications",
		"POST /api/notifications/read",
		"GET /api/nodes/:id/api-requests",
		"POST /api/nodes/:id/api-requests",
		"PATCH /api/api-requests/:id",
		"DELETE /api/api-requests/:id",
		"GET /api/api-requests/:id/runs",
		"PATCH /api/api-runs/:id/evidence",
		"POST /api/api-requests/:id/runs",

		// AI routes
		"POST /api/ai/generate-flow",
		"GET /api/ai/jobs/:id",
		"POST /api/ai/audit-flow",
	}

	for _, expected := range expectedRoutes {
		if !registeredRoutes[expected] {
			t.Errorf("missing expected route: %s", expected)
		}
	}
}
