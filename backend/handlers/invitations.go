package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"backend/db"
	"backend/middleware"
	"backend/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func validProjectRole(role string) bool {
	return role == "owner" || role == "editor" || role == "commenter" || role == "viewer"
}
func validFunctionalRole(role string) bool {
	return role == "" || role == "pm" || role == "uiux" || role == "frontend" || role == "backend"
}

func CreateInvitationHandler(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	organizationID, ok := c.Get(string(middleware.OrganizationContextKey))
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization access denied"})
		return
	}
	var req models.InvitationRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	email, projectRole, functionalRole := strings.ToLower(strings.TrimSpace(req.Email)), strings.TrimSpace(req.ProjectRole), strings.TrimSpace(req.FunctionalRole)
	if !emailRegex.MatchString(email) || !validProjectRole(projectRole) || !validFunctionalRole(functionalRole) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation"})
		return
	}
	if req.ProjectID != "" {
		if !middleware.AuthorizeProject(c, req.ProjectID, middleware.CapabilityManage) {
			return
		}
	} else {
		var role string
		if db.DB.QueryRow(`SELECT role FROM organization_members WHERE organization_id=$1 AND user_id=$2 AND status='active'`, organizationID, userID).Scan(&role) != nil || role != "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Organization owner permission required"})
			return
		}
	}
	token, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}
	_, err = db.DB.Exec(`INSERT INTO organization_invitations (id, organization_id, email, project_id, project_role, functional_role, token_hash, expires_at, invited_by) VALUES ($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,''),$7,$8,$9)`, "inv_"+GenerateUUID(), organizationID, email, req.ProjectID, projectRole, functionalRole, tokenHash(token), time.Now().Add(7*24*time.Hour), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save invitation"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "expires_in": "168h"})
}

func AcceptInvitationHandler(c *gin.Context) {
	var req models.InvitationAcceptRequest
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Token) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation"})
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}
	defer tx.Rollback()
	var organizationID, email, projectID, projectRole, functionalRole string
	err = tx.QueryRow(`SELECT organization_id,email,COALESCE(project_id,''),project_role,COALESCE(functional_role,'') FROM organization_invitations WHERE token_hash=$1 AND accepted_at IS NULL AND expires_at>CURRENT_TIMESTAMP FOR UPDATE`, tokenHash(req.Token)).Scan(&organizationID, &email, &projectID, &projectRole, &functionalRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation is invalid, expired, or already used"})
		return
	}
	var userID string
	err = tx.QueryRow(`SELECT id FROM users WHERE email=$1`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		if len(strings.TrimSpace(req.Name)) < 2 || len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and password are required for a new account"})
			return
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		userID = "usr_" + GenerateUUID()
		if _, err = tx.Exec(`INSERT INTO users (id,name,email,password_hash,role) VALUES ($1,$2,$3,$4,$5)`, userID, strings.TrimSpace(req.Name), email, string(hash), defaultFunctionalRole(functionalRole)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}
	if _, err = tx.Exec(`INSERT INTO organization_members (organization_id,user_id,role,status) VALUES ($1,$2,'member','active') ON CONFLICT (organization_id,user_id) DO UPDATE SET status='active'`, organizationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join organization"})
		return
	}
	if projectID != "" {
		if _, err = tx.Exec(`INSERT INTO project_members (project_id,user_id,project_role,functional_role,added_by) VALUES ($1,$2,$3,NULLIF($4,''),$5) ON CONFLICT (project_id,user_id) DO UPDATE SET project_role=EXCLUDED.project_role,functional_role=EXCLUDED.functional_role`, projectID, userID, projectRole, functionalRole, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join project"})
			return
		}
	}
	if _, err = tx.Exec(`UPDATE organization_invitations SET accepted_at=CURRENT_TIMESTAMP WHERE token_hash=$1 AND accepted_at IS NULL`, tokenHash(req.Token)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}
	c.Status(http.StatusNoContent)
}

func defaultFunctionalRole(role string) string {
	if role == "" {
		return "frontend"
	}
	return role
}
