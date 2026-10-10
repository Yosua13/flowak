package routes

import (
	"backend/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterAIRoutes registers AI generation and auditing proxy endpoints.
func RegisterAIRoutes(r *gin.RouterGroup) {
	r.POST("/ai/generate-flow", handlers.AiGenerateFlowHandler)
	r.POST("/ai/audit-flow", handlers.AiAuditFlowHandler)
}
