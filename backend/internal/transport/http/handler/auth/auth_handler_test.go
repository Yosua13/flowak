package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	handlerAuth "backend/internal/transport/http/handler/auth"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleRegister_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString("invalid json"))

	handlerAuth.HandleRegister(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString("invalid json"))

	handlerAuth.HandleLogin(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got: %d", w.Code)
	}
}

func TestHandleRefresh_MissingCookie(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)

	handlerAuth.HandleRefresh(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got: %d", w.Code)
	}
}

func TestHandleLogout_ClearsCookies(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

	handlerAuth.HandleLogout(c)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got: %d", w.Code)
	}
}
