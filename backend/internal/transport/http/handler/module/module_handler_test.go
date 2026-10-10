package module_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	moduleHandler "backend/internal/transport/http/handler/module"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleSyncGraph_Unauthorized(t *testing.T) {
	r := gin.New()
	r.PUT("/modules/:id", moduleHandler.HandleSyncGraph)

	req := httptest.NewRequest(http.MethodPut, "/modules/mod_1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got: %d", w.Code)
	}
}

func TestHandleGetGraph_Unauthorized(t *testing.T) {
	r := gin.New()
	r.GET("/modules/:id/graph", moduleHandler.HandleGetGraph)

	req := httptest.NewRequest(http.MethodGet, "/modules/mod_1/graph", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got: %d", w.Code)
	}
}
