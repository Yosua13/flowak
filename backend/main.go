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
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle OPTIONS preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Initialize config & environment
	config.InitConfig()
	log.Printf("Starting Flowak Backend on port %s...", config.ActiveConfig.Port)

	// 2. Initialize database
	db.InitDB("flowak.db")
	defer db.DB.Close()

	// 3. Setup router (Go 1.22+ custom pattern routing)
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("POST /api/auth/login", handlers.LoginHandler)

	// Protected routes sub-router simulation via inline handler matching
	protectedMux := http.NewServeMux()

	// Project CRUD
	protectedMux.HandleFunc("GET /api/projects", handlers.ProjectsHandler)
	protectedMux.HandleFunc("POST /api/projects", handlers.ProjectsHandler)
	protectedMux.HandleFunc("GET /api/projects/{id}", handlers.ProjectDetailHandler)
	protectedMux.HandleFunc("DELETE /api/projects/{id}", handlers.ProjectDetailHandler)

	// Module CRUD
	protectedMux.HandleFunc("GET /api/projects/{id}/modules", handlers.ProjectModulesHandler)
	protectedMux.HandleFunc("POST /api/projects/{id}/modules", handlers.ProjectModulesHandler)
	protectedMux.HandleFunc("PUT /api/modules/{id}", handlers.ModuleDetailHandler)
	protectedMux.HandleFunc("DELETE /api/modules/{id}", handlers.ModuleDetailHandler)

	// AI Proxies
	protectedMux.HandleFunc("POST /api/ai/generate-flow", handlers.AiGenerateFlowHandler)
	protectedMux.HandleFunc("POST /api/ai/mock-payload", handlers.AiMockPayloadHandler)
	protectedMux.HandleFunc("POST /api/ai/audit-flow", handlers.AiAuditFlowHandler)
	protectedMux.HandleFunc("POST /api/ai/generate-code", handlers.AiGenerateCodeHandler)

	// Register protected endpoints with JWT middleware
	mux.Handle("/api/projects", middleware.AuthMiddleware(protectedMux))
	mux.Handle("/api/projects/", middleware.AuthMiddleware(protectedMux))
	mux.Handle("/api/modules/", middleware.AuthMiddleware(protectedMux))
	mux.Handle("/api/ai/", middleware.AuthMiddleware(protectedMux))

	// Catch-all route to serve compiled static assets from the frontend/dist folder (Production)
	distDir := "../frontend/dist"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent directory traversal vulnerabilities
		cleanedPath := filepath.Clean(r.URL.Path)
		filePath := filepath.Join(distDir, cleanedPath)

		// Security: Prevent directory traversal escaping the distDir root
		absDist, _ := filepath.Abs(distDir)
		absFile, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFile, absDist) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if file exists and is not a directory
		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			// If file does not exist or is a directory, fallback to index.html (SPA routing)
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}

		// Otherwise serve the static asset
		http.ServeFile(w, r, filePath)
	})

	// Global Middleware
	handler := CorsMiddleware(mux)

	// 4. Start HTTP Server
	serverAddr := ":" + config.ActiveConfig.Port
	log.Printf("Server listening on http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
