package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WriteValidationError writes a standardized 400 Bad Request error response for validation failures.
func WriteValidationError(c *gin.Context, message string) {
	if message == "" {
		message = "Invalid request body"
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}
