package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/db"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestProjectAuthorizationMatrix(t *testing.T) {
	cases := []struct {
		role       string
		capability ProjectCapability
		want       bool
	}{
		{"owner", CapabilityManage, true},
		{"owner", CapabilityEditGraph, true},
		{"editor", CapabilityEditGraph, true},
		{"editor", CapabilityManage, false},
		{"commenter", CapabilityWorkItem, true},
		{"commenter", CapabilityEditGraph, false},
		{"viewer", CapabilityView, true},
		{"viewer", CapabilityComment, false},
		{"viewer", CapabilityEditGraph, false},
		{"viewer", CapabilityManage, false},
	}
	for _, tc := range cases {
		if got := can(tc.role, tc.capability); got != tc.want {
			t.Errorf("can(%q, %q) = %v, want %v", tc.role, tc.capability, got, tc.want)
		}
	}
}

func TestUnknownProjectRoleHasNoCapabilities(t *testing.T) {
	for _, capability := range []ProjectCapability{CapabilityView, CapabilityComment, CapabilityWorkItem, CapabilityEditGraph, CapabilityManage} {
		if can("pm", capability) || can("", capability) {
			t.Fatalf("functional or empty roles must not authorize %q", capability)
		}
	}
}

func TestAuthorizeProjectRejectsCrossTenantIDOR(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer mockDB.Close()
	previous := db.DB
	db.DB = mockDB
	defer func() { db.DB = previous }()

	// A guessed project ID in another tenant produces no membership row.
	mock.ExpectQuery("SELECT COALESCE").WithArgs("user-a", "project-b", "org-a").WillReturnError(sql.ErrNoRows)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/modules/module-b", nil)
	c.Set(string(UserContextKey), "user-a")
	c.Set(string(OrganizationContextKey), "org-a")
	if AuthorizeProject(c, "project-b", CapabilityEditGraph) { t.Fatal("cross-tenant project access was allowed") }
	if w.Code != http.StatusForbidden { t.Fatalf("status = %d, want 403", w.Code) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestAuthorizeProjectDeniesViewerGraphUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockDB, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer mockDB.Close()
	previous := db.DB
	db.DB = mockDB
	defer func() { db.DB = previous }()
	mock.ExpectQuery("SELECT COALESCE").WithArgs("viewer", "project-a", "org-a").WillReturnRows(sqlmock.NewRows([]string{"project_role"}).AddRow("viewer"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/modules/module-a", nil)
	c.Set(string(UserContextKey), "viewer")
	c.Set(string(OrganizationContextKey), "org-a")
	if AuthorizeProject(c, "project-a", CapabilityEditGraph) { t.Fatal("viewer was allowed to edit graph") }
	if w.Code != http.StatusForbidden { t.Fatalf("status = %d, want 403", w.Code) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
