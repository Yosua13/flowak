package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"backend/db"
	"backend/middleware"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

const (
	tenantTestUser = "user_a"
	tenantTestOrg  = "org_a"
)

func tenantTestRouter(method, path string, handler gin.HandlerFunc, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserContextKey), tenantTestUser)
		c.Set(string(middleware.OrganizationContextKey), tenantTestOrg)
		c.Set(string(middleware.ProjectRoleContextKey), role)
		c.Next()
	})
	router.Handle(method, path, handler)
	return router
}

func isolatedTenantDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	database, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	previous := db.DB
	db.DB = database
	t.Cleanup(func() {
		db.DB = previous
		database.Close()
	})
	return mock
}

func expectProjectDenied(mock sqlmock.Sqlmock, projectID string) {
	mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, projectID, tenantTestOrg).WillReturnError(sql.ErrNoRows)
}

func assertAPIError(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d: %s", response.Code, wantStatus, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["error"] != wantError {
		t.Fatalf("error = %#v, want %q", body["error"], wantError)
	}
}

func TestTenantEndpointsRejectCrossOrganizationResources(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		route   string
		path    string
		body    string
		handler gin.HandlerFunc
		prepare func(sqlmock.Sqlmock)
	}{
		{
			name: "project detail", method: http.MethodGet, route: "/projects/:id", path: "/projects/project_b", handler: GetProjectDetailHandler,
			prepare: func(mock sqlmock.Sqlmock) { expectProjectDenied(mock, "project_b") },
		},
		{
			name: "module list", method: http.MethodGet, route: "/projects/:id/modules", path: "/projects/project_b/modules", handler: GetProjectModulesHandler,
			prepare: func(mock sqlmock.Sqlmock) { expectProjectDenied(mock, "project_b") },
		},
		{
			name: "graph update", method: http.MethodPut, route: "/modules/:id", path: "/modules/module_b", body: `{}`, handler: UpdateModuleHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1")).
					WithArgs("module_b").WillReturnRows(sqlmock.NewRows([]string{"owner_id", "project_id"}).AddRow("owner_b", "project_b"))
				expectProjectDenied(mock, "project_b")
			},
		},
		{
			name: "work item", method: http.MethodGet, route: "/work-items/:key", path: "/work-items/WORK-B-1", handler: GetWorkItemHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id FROM work_items WHERE work_key=$1 AND deleted_at IS NULL")).
					WithArgs("WORK-B-1").WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow("project_b"))
				expectProjectDenied(mock, "project_b")
			},
		},
		{
			name: "node comment", method: http.MethodPost, route: "/nodes/:id/comments", path: "/nodes/node_b/comments", body: `{"body":"should not be created"}`, handler: CreateNodeCommentHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT m.project_id FROM workflow_nodes").WithArgs("node_b").
					WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow("project_b"))
				expectProjectDenied(mock, "project_b")
			},
		},
		{
			name: "api runner", method: http.MethodPost, route: "/api-requests/:id/runs", path: "/api-requests/request_b/runs", body: `{"environment_id":"env_b"}`, handler: RunAPIRequestHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT m.project_id,r.method,r.relative_path,e.approved_base_url").WithArgs("request_b", "env_b").
					WillReturnRows(sqlmock.NewRows([]string{"project_id", "method", "relative_path", "approved_base_url"}).AddRow("project_b", "GET", "/v1/items", "https://api.example.test"))
				expectProjectDenied(mock, "project_b")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := isolatedTenantDB(t)
			tc.prepare(mock)
			router := tenantTestRouter(tc.method, tc.route, tc.handler, "viewer")
			request := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertAPIError(t, response, http.StatusForbidden, "Project access denied")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestViewerCannotMutateTenantResources(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		route   string
		path    string
		body    string
		handler gin.HandlerFunc
		prepare func(sqlmock.Sqlmock)
	}{
		{
			name: "work item", method: http.MethodPost, route: "/projects/:id/work-items", path: "/projects/project_a/work-items", body: `{"type":"Task","title":"Blocked","description":"Blocked"}`, handler: CreateWorkItemHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project_a", tenantTestOrg).
					WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
			},
		},
		{
			name: "graph", method: http.MethodPut, route: "/modules/:id", path: "/modules/module_a", body: `{}`, handler: UpdateModuleHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT p.owner_id, m.project_id FROM modules m JOIN projects p ON m.project_id = p.id WHERE m.id = $1")).
					WithArgs("module_a").WillReturnRows(sqlmock.NewRows([]string{"owner_id", "project_id"}).AddRow("owner_a", "project_a"))
				mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project_a", tenantTestOrg).
					WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
			},
		},
		{
			name: "comment", method: http.MethodPost, route: "/nodes/:id/comments", path: "/nodes/node_a/comments", body: `{"body":"Blocked"}`, handler: CreateNodeCommentHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT m.project_id FROM workflow_nodes").WithArgs("node_a").
					WillReturnRows(sqlmock.NewRows([]string{"project_id"}).AddRow("project_a"))
				mock.ExpectQuery("SELECT COALESCE").WithArgs(tenantTestUser, "project_a", tenantTestOrg).
					WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := isolatedTenantDB(t)
			tc.prepare(mock)
			router := tenantTestRouter(tc.method, tc.route, tc.handler, "viewer")
			request := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertAPIError(t, response, http.StatusForbidden, "Project access denied")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTenantScopedNotificationAndUserPickerQueries(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
		prepare func(sqlmock.Sqlmock)
	}{
		{
			name: "notifications", method: http.MethodGet, path: "/notifications", handler: ListNotificationsHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id,title,body,type,read_at,created_at,event_name FROM notifications WHERE user_id=\\$1").WithArgs(tenantTestUser).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "body", "type", "read_at", "created_at", "event_name"}))
			},
		},
		{
			name: "user picker", method: http.MethodGet, path: "/users", handler: GetUsersHandler,
			prepare: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT u.id, u.name, u.email, u.role FROM users u").WithArgs(tenantTestOrg).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role"}).AddRow(tenantTestUser, "Tenant User", "user@example.test", "frontend"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := isolatedTenantDB(t)
			tc.prepare(mock)
			router := tenantTestRouter(tc.method, tc.path, tc.handler, "viewer")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSessionLogoutAndRotationPreventRefreshReuse(t *testing.T) {
	mock := isolatedTenantDB(t)
	refreshToken := "existing-refresh-token"

	mock.ExpectExec("UPDATE user_sessions SET revoked_at=CURRENT_TIMESTAMP").WithArgs(tokenHash(refreshToken)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT s.id, s.organization_id, om.role").WithArgs(tokenHash(refreshToken)).WillReturnError(sql.ErrNoRows)

	router := gin.New()
	router.POST("/logout", LogoutHandler)
	router.POST("/refresh", RefreshHandler)
	logoutRequest := httptest.NewRequest(http.MethodPost, "/logout", nil)
	logoutRequest.AddCookie(&http.Cookie{Name: "flowak_refresh", Value: refreshToken})
	logoutResponse := httptest.NewRecorder()
	router.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logoutResponse.Code, http.StatusNoContent)
	}
	refreshRequest := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	refreshRequest.AddCookie(&http.Cookie{Name: "flowak_refresh", Value: refreshToken})
	refreshResponse := httptest.NewRecorder()
	router.ServeHTTP(refreshResponse, refreshRequest)
	assertAPIError(t, refreshResponse, http.StatusUnauthorized, "Invalid session")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshRotationRejectsReusedToken(t *testing.T) {
	mock := isolatedTenantDB(t)
	refreshToken := "rotated-refresh-token"
	createdAt := time.Now().UTC()
	mock.ExpectQuery("SELECT s.id, s.organization_id, om.role").WithArgs(tokenHash(refreshToken)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "role", "id", "name", "email", "role", "created_at"}).
			AddRow("session_a", tenantTestOrg, "member", tenantTestUser, "Tenant User", "user@example.test", "frontend", createdAt))
	mock.ExpectExec("UPDATE user_sessions SET refresh_token_hash=\\$1, last_used_at=CURRENT_TIMESTAMP").
		WithArgs(sqlmock.AnyArg(), "session_a", tokenHash(refreshToken)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT s.id, s.organization_id, om.role").WithArgs(tokenHash(refreshToken)).WillReturnError(sql.ErrNoRows)

	router := gin.New()
	router.POST("/refresh", RefreshHandler)
	for attempt, wantStatus := range []int{http.StatusOK, http.StatusUnauthorized} {
		request := httptest.NewRequest(http.MethodPost, "/refresh", nil)
		request.AddCookie(&http.Cookie{Name: "flowak_refresh", Value: refreshToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != wantStatus {
			t.Fatalf("attempt %d status = %d, want %d: %s", attempt+1, response.Code, wantStatus, response.Body.String())
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInvitationExpiryAndReuseReturnStableError(t *testing.T) {
	for _, token := range []string{"expired-token", "used-token"} {
		t.Run(token, func(t *testing.T) {
			mock := isolatedTenantDB(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT organization_id,email,COALESCE\\(project_id,''\\),project_role").WithArgs(tokenHash(token)).WillReturnError(sql.ErrNoRows)
			mock.ExpectRollback()
			router := gin.New()
			router.POST("/invitations/accept", AcceptInvitationHandler)
			request := httptest.NewRequest(http.MethodPost, "/invitations/accept", bytes.NewBufferString(`{"token":"`+token+`"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assertAPIError(t, response, http.StatusBadRequest, "Invitation is invalid, expired, or already used")
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
