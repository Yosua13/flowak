package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backend/config"
	"backend/db"
	"backend/handlers"
	"backend/internal/infra/rabbitmq"
	"backend/internal/infra/rabbitmq/consumer"
	"backend/internal/infra/rabbitmq/producer"
	"backend/internal/repository/redis"
	workitemHandler "backend/internal/transport/http/handler/workitem"
	"backend/internal/transport/http/routes"
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
	if err := handlers.PruneExpiredCollaborationRecords(); err != nil {
		log.Printf("collaboration retention cleanup skipped: %v", err)
	}

	// 3. Initialize Redis infrastructure (with graceful degradation fallback)
	if redisClient, err := redis.NewRedisClient(&config.ActiveConfig); err == nil && redisClient != nil {
		middleware.GlobalRedisClient = redisClient
	} else if redisClient != nil {
		middleware.GlobalRedisClient = redisClient
	}

	// 4. Initialize RabbitMQ infrastructure & consumers (with graceful degradation fallback)
	if rabbitConn, err := rabbitmq.NewRabbitMQConnection(config.ActiveConfig.RabbitMQURL); err == nil && rabbitConn != nil {
		if ch, err := rabbitConn.Channel(); err == nil && ch != nil {
			_ = rabbitmq.SetupTopology(ch)
			pub := producer.NewWorkItemEventPublisher(ch)
			workitemHandler.SetEventPublisher(pub)

			go func() {
				_ = consumer.ConsumeSSEBroadcast(context.Background(), ch, func(projectID string, event []byte) {
					handlers.BroadcastProjectEvent(projectID, event)
				})
			}()
		}
	}

	// 5. Setup router (Gin Engine)
	r := gin.Default()

	// Global Middleware
	r.Use(CorsMiddleware())

	// Public routes group
	public := r.Group("/api")
	routes.RegisterAuthRoutes(public)

	// Protected routes group
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	routes.RegisterProjectRoutes(api)
	routes.RegisterModuleRoutes(api)
	routes.RegisterWorkItemRoutes(api)
	routes.RegisterAIRoutes(api)

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
