package routes

import (
	"backend/handlers"
	aiHandler "backend/internal/transport/http/handler/ai"
	"github.com/gin-gonic/gin"
)

// RegisterAIRoutes registers AI generation and auditing proxy endpoints.
func RegisterAIRoutes(r *gin.RouterGroup) {
	r.POST("/ai/generate-flow", aiHandler.HandleGenerateFlow)
	r.GET("/ai/jobs/:id", aiHandler.HandleGetJobStatus)
	r.POST("/ai/audit-flow", handlers.AiAuditFlowHandler)
}
