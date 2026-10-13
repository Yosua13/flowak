package routes

import (
	"backend/handlers"
	workitemHandler "backend/internal/transport/http/handler/workitem"
	"github.com/gin-gonic/gin"
)

// RegisterWorkItemRoutes registers work items, activity, comments, artifacts, notifications, and API contract runner routes.
func RegisterWorkItemRoutes(r *gin.RouterGroup) {
	// Work items CRUD & workflow
	r.GET("/projects/:id/work-items", workitemHandler.HandleList)
	r.POST("/projects/:id/work-items", workitemHandler.HandleCreate)
	r.GET("/work-items/:key", workitemHandler.HandleGetDetail)
	r.PATCH("/work-items/:key", workitemHandler.HandleUpdate)
	r.POST("/work-items/:key/transitions", workitemHandler.HandleTransition)
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
