package routes

import (
	"backend/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterModuleRoutes registers module CRUD and baseline versioning routes.
func RegisterModuleRoutes(r *gin.RouterGroup) {
	// Module CRUD
	r.GET("/projects/:id/modules", handlers.GetProjectModulesHandler)
	r.POST("/projects/:id/modules", handlers.CreateProjectModuleHandler)
	r.PUT("/modules/:id", handlers.UpdateModuleHandler)
	r.DELETE("/modules/:id", handlers.DeleteModuleHandler)

	// Module Baselines & Versions
	r.POST("/modules/:id/publish", handlers.PublishModuleBaselineHandler)
	r.GET("/modules/:id/versions", handlers.ListModuleBaselinesHandler)
	r.POST("/modules/:id/versions/:version/restore", handlers.RestoreModuleBaselineHandler)
}
