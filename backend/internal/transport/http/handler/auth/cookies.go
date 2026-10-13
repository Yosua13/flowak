package auth

import (
	"net/http"

	"backend/config"
	"github.com/gin-gonic/gin"
)

// SetAuthCookies writes refresh and short-lived SSE access cookies to the response.
func SetAuthCookies(c *gin.Context, refreshToken, accessToken string) {
	c.SetSameSite(http.SameSiteLaxMode)
	isProd := config.ActiveConfig.Environment == "production"

	if refreshToken != "" {
		c.SetCookie("flowak_refresh", refreshToken, 30*24*60*60, "/api/auth", "", isProd, true)
	}
	if accessToken != "" {
		c.SetCookie("flowak_sse_access", accessToken, 15*60, "/api/projects", "", isProd, true)
	}
}

// ClearAuthCookies clears the auth session cookies upon logout.
func ClearAuthCookies(c *gin.Context) {
	isProd := config.ActiveConfig.Environment == "production"
	c.SetCookie("flowak_refresh", "", -1, "/api/auth", "", isProd, true)
	c.SetCookie("flowak_sse_access", "", -1, "/api/projects", "", isProd, true)
}
