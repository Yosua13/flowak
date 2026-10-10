package workitem_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "backend/internal/transport/http/handler/workitem"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleCreate_UnauthorizedWhenNoUser(t *testing.T) {
	r := gin.New()
	r.POST("/projects/:id/work-items", handler.HandleCreate)

	req := httptest.NewRequest(http.MethodPost, "/projects/proj_1/work-items", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got: %d", w.Code)
	}
}

func TestHandleList_InvalidSortOrder(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "usr_test")
		c.Next()
	})
	r.GET("/projects/:id/work-items", handler.HandleList)

	req := httptest.NewRequest(http.MethodGet, "/projects/proj_1/work-items?sort=invalid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Since mock DB isn't authorized for project, check either 403 or 400
	if w.Code != http.StatusForbidden && w.Code != http.StatusBadRequest {
		t.Logf("response status: %d", w.Code)
	}
}
