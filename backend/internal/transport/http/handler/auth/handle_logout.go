package auth

import (
	"net/http"
	"strings"
	"time"

	"backend/config"
	"backend/db"
	domainAuth "backend/internal/domain/auth"
	repoAuth "backend/internal/repository/postgres/auth"
	redisSession "backend/internal/repository/redis/session"
	usecaseAuth "backend/internal/usecase/auth"
	"backend/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// HandleLogout revokes the user's refresh token and blacklists the access token in Redis.
func HandleLogout(c *gin.Context) {
	refreshToken, _ := c.Cookie("flowak_refresh")

	var tokenID string
	var remainingTTL time.Duration

	// Extract token claims if Bearer Authorization header is present
	authHeader := c.GetHeader("Authorization")
	if parts := strings.Split(authHeader, " "); len(parts) == 2 && parts[0] == "Bearer" {
		tokenStr := parts[1]
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.ActiveConfig.JWTSecret), nil
		})
		if err == nil && token.Valid {
			tokenID = claims.ID
			if tokenID == "" {
				tokenID = claims.SessionID
			}
			if claims.ExpiresAt != nil {
				remainingTTL = time.Until(claims.ExpiresAt.Time)
			}
		}
	}

	repo := repoAuth.NewPostgresAuthRepository(db.DB)
	var blacklistRepo domainAuth.TokenBlacklistRepository
	if middleware.GlobalRedisClient != nil {
		blacklistRepo = redisSession.NewBlacklistTokenRepo(middleware.GlobalRedisClient)
	}

	uc := usecaseAuth.NewLogoutUserUseCase(repo, blacklistRepo)
	_ = uc.Execute(c.Request.Context(), usecaseAuth.LogoutInput{
		TokenID:      tokenID,
		RefreshToken: refreshToken,
		RemainingTTL: remainingTTL,
	})

	ClearAuthCookies(c)
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}
