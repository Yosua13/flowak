package routes

import (
	"backend/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterWorkItemRoutes registers work items, activity, comments, artifacts, notifications, and API contract runner routes.
func RegisterWorkItemRoutes(r *gin.RouterGroup) {
	// Work items CRUD & workflow
	r.GET("/projects/:id/work-items", handlers.ListWorkItemsHandler)
	r.POST("/projects/:id/work-items", handlers.CreateWorkItemHandler)
	r.GET("/work-items/:key", handlers.GetWorkItemHandler)
	r.PATCH("/work-items/:key", handlers.UpdateWorkItemHandler)
	r.POST("/work-items/:key/transitions", handlers.TransitionWorkItemHandler)
	r.GET("/work-items/:key/activity", handlers.GetWorkItemActivityHandler)

	// Work item artifacts
	r.GET("/work-items/:key/artifacts", handlers.GetWorkItemArtifactsHandler)
	r.POST("/work-items/:key/artifacts/:kind", handlers.MutateWorkItemArtifactHandler)
	r.PATCH("/work-items/:key/artifacts/:kind/:artifactId", handlers.MutateWorkItemArtifactHandler)
	r.DELETE("/work-items/:key/artifacts/:kind/:artifactId", handlers.MutateWorkItemArtifactHandler)

	// Work item comments
	r.POST("/work-items/:key/comments", handlers.CreateWorkItemCommentHandler)
	r.GET("/work-items/:key/comments", handlers.ListWorkItemCommentsHandler)

	// Canvas node comments & activity
	r.GET("/nodes/:id/comments", handlers.ListNodeCommentsHandler)
	r.POST("/nodes/:id/comments", handlers.CreateNodeCommentHandler)
	r.GET("/nodes/:id/activity", handlers.ListNodeActivityHandler)
	r.PATCH("/comments/:id", handlers.UpdateCommentHandler)

	// Notifications
	r.GET("/notifications", handlers.ListNotificationsHandler)
	r.POST("/notifications/read", handlers.MarkNotificationsReadHandler)

	// Node API requests & test runner
	r.GET("/nodes/:id/api-requests", handlers.ListAPIRequestsHandler)
	r.POST("/nodes/:id/api-requests", handlers.CreateAPIRequestHandler)
	r.PATCH("/api-requests/:id", handlers.UpdateAPIRequestHandler)
	r.DELETE("/api-requests/:id", handlers.DeleteAPIRequestHandler)
	r.GET("/api-requests/:id/runs", handlers.ListAPIRunsHandler)
	r.PATCH("/api-runs/:id/evidence", handlers.SaveAPIRunEvidenceHandler)
	r.POST("/api-requests/:id/runs", handlers.RunAPIRequestHandler)
}
