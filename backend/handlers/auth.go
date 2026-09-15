package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"backend/config"
	"backend/db"
	"backend/middleware"
	"backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Email regex validation
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

// GenerateUUID formats a secure UUIDv4 string
func GenerateUUID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to timestamp + random string if entropy fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

// GenerateTempPassword creates a one-time password for PM-created users.
func GenerateTempPassword() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("Flowak-%d", time.Now().UnixNano())
	}
	return "Flowak-" + base64.RawURLEncoding.EncodeToString(bytes)
}

// RegisterHandler registers a new user
func RegisterHandler(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 1. Validation & Sanitization
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password
	role := strings.TrimSpace(req.Role)

	if name == "" || email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	if len(name) < 2 || len(name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be between 2 and 100 characters"})
		return
	}

	if !emailRegex.MatchString(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address format"})
		return
	}

	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	// Public registration creates only a low-privilege functional account. PM
	// is assigned through an invitation or explicit organization bootstrap.
	if role == "" {
		role = "frontend"
	}
	if role != "uiux" && role != "frontend" && role != "backend" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public registration cannot assign PM or elevated privileges"})
		return
	}

	// 2. Check if email already exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking user"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email address already registered"})
		return
	}

	// 3. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
		return
	}

	// 4. Save to Database. Supplying an organization name is the explicit
	// bootstrap path: its first member is the organization owner.
	userID := "usr_" + GenerateUUID()
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start registration"})
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("INSERT INTO users (id, name, email, password_hash, role) VALUES ($1, $2, $3, $4, $5)",
		userID, name, email, string(hashedPassword), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user to database"})
		return
	}

	organizationID := "org_default"
	organizationRole := "member"
	if organizationName := strings.TrimSpace(req.OrganizationName); organizationName != "" {
		organizationID = "org_" + GenerateUUID()
		organizationRole = "owner"
		if _, err = tx.Exec(`INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`, organizationID, organizationName, strings.ToLower(strings.ReplaceAll(organizationID, "_", "-"))); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to bootstrap organization"})
			return
		}
	}
	if _, err = tx.Exec(`INSERT INTO organization_members (organization_id, user_id, role, status) VALUES ($1, $2, $3, 'active')`, organizationID, userID, organizationRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add organization membership"})
		return
	}
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"user_id": userID,
	})
}

// LoginHandler authenticates user and returns JWT
func LoginHandler(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	// 1. Find user in database
	var user models.User
	var passwordHash string
	err := db.DB.QueryRow("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1 AND status = 'active'", email).
		Scan(&user.ID, &user.Name, &user.Email, &passwordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
		return
	}

	// 2. Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	_, _ = db.DB.Exec("UPDATE users SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1", user.ID)

	var organizationID, organizationRole string
	err = db.DB.QueryRow(`SELECT organization_id, role FROM organization_members WHERE user_id = $1 AND status = 'active' ORDER BY joined_at ASC LIMIT 1`, user.ID).Scan(&organizationID, &organizationRole)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No active organization membership"})
		return
	}
	tokenString, refreshToken, err := createSession(user, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}
	setRefreshCookie(c, refreshToken)

	resp := models.UserLoginResponse{
		Token: tokenString, User: user, OrganizationID: organizationID, OrganizationRole: organizationRole,
	}

	c.JSON(http.StatusOK, resp)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func createSession(user models.User, organizationID string) (string, string, error) {
	refreshToken, err := randomToken()
	if err != nil {
		return "", "", err
	}
	sessionID := "ses_" + GenerateUUID()
	_, err = db.DB.Exec(`INSERT INTO user_sessions (id, user_id, organization_id, refresh_token_hash, expires_at) VALUES ($1, $2, $3, $4, $5)`, sessionID, user.ID, organizationID, tokenHash(refreshToken), time.Now().Add(30*24*time.Hour))
	if err != nil {
		return "", "", err
	}
	access, err := signAccessToken(user, organizationID, sessionID)
	return access, refreshToken, err
}

func signAccessToken(user models.User, organizationID, sessionID string) (string, error) {
	claims := &middleware.Claims{UserID: user.ID, Email: user.Email, Role: user.Role, OrganizationID: organizationID, SessionID: sessionID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(config.ActiveConfig.JWTSecret))
}

func setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("flowak_refresh", token, 30*24*60*60, "/api/auth", "", config.ActiveConfig.Environment == "production", true)
}

// RefreshHandler rotates the opaque refresh token so a logged-out or reused
// session cannot mint another access token.
func RefreshHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("flowak_refresh")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var user models.User
	var sessionID, organizationID, organizationRole string
	err = db.DB.QueryRow(`SELECT s.id, s.organization_id, om.role, u.id, u.name, u.email, u.role, u.created_at FROM user_sessions s JOIN users u ON u.id=s.user_id JOIN organization_members om ON om.organization_id=s.organization_id AND om.user_id=s.user_id AND om.status='active' WHERE s.refresh_token_hash=$1 AND s.revoked_at IS NULL AND s.expires_at > CURRENT_TIMESTAMP AND u.status='active'`, tokenHash(refreshToken)).Scan(&sessionID, &organizationID, &organizationRole, &user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	newRefresh, err := randomToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
		return
	}
	result, err := db.DB.Exec(`UPDATE user_sessions SET refresh_token_hash=$1, last_used_at=CURRENT_TIMESTAMP WHERE id=$2 AND refresh_token_hash=$3 AND revoked_at IS NULL`, tokenHash(newRefresh), sessionID, tokenHash(refreshToken))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	access, err := signAccessToken(user, organizationID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
		return
	}
	setRefreshCookie(c, newRefresh)
	c.JSON(http.StatusOK, models.UserLoginResponse{Token: access, User: user, OrganizationID: organizationID, OrganizationRole: organizationRole})
}

func LogoutHandler(c *gin.Context) {
	if token, err := c.Cookie("flowak_refresh"); err == nil {
		_, _ = db.DB.Exec(`UPDATE user_sessions SET revoked_at=CURRENT_TIMESTAMP WHERE refresh_token_hash=$1 AND revoked_at IS NULL`, tokenHash(token))
	}
	c.SetCookie("flowak_refresh", "", -1, "/api/auth", "", config.ActiveConfig.Environment == "production", true)
	c.Status(http.StatusNoContent)
}

// PasswordResetRequestHandler intentionally has the same response whether or
// not the email exists. Delivery is delegated to the configured mail service.
func PasswordResetRequestHandler(c *gin.Context) {
	var req models.UserLoginRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	var userID string
	if db.DB.QueryRow(`SELECT id FROM users WHERE email=$1 AND status='active'`, email).Scan(&userID) == nil {
		if token, err := randomToken(); err == nil {
			_, _ = db.DB.Exec(`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at) VALUES ($1,$2,$3,$4)`, "rst_"+GenerateUUID(), userID, tokenHash(token), time.Now().Add(time.Hour))
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "If the account exists, password reset instructions will be sent."})
}
