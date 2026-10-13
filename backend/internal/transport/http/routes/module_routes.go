package routes

import (
	"backend/handlers"
	moduleHandler "backend/internal/transport/http/handler/module"
	"github.com/gin-gonic/gin"
)

// RegisterModuleRoutes registers module CRUD and baseline versioning routes.
func RegisterModuleRoutes(r *gin.RouterGroup) {
	// Module CRUD
	r.GET("/projects/:id/modules", handlers.GetProjectModulesHandler)
	r.POST("/projects/:id/modules", handlers.CreateProjectModuleHandler)
	r.PUT("/modules/:id", moduleHandler.HandleSyncGraph)
	r.GET("/modules/:id/graph", moduleHandler.HandleGetGraph)
	r.DELETE("/modules/:id", handlers.DeleteModuleHandler)

	// Module Baselines & Versions
	r.POST("/modules/:id/publish", moduleHandler.HandlePublishBaseline)
	r.GET("/modules/:id/versions", handlers.ListModuleBaselinesHandler)
	r.POST("/modules/:id/versions/:version/restore", handlers.RestoreModuleBaselineHandler)
}
