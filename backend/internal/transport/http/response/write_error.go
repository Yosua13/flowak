package response

import (
	"errors"
	"net/http"

	"backend/internal/domain/common"
	"github.com/gin-gonic/gin"
)

// WriteError maps domain errors to the appropriate HTTP status code and writes a JSON error response.
func WriteError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, common.ErrNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, common.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, common.ErrForbidden):
		statusCode = http.StatusForbidden
	case errors.Is(err, common.ErrConflict):
		statusCode = http.StatusConflict
	case errors.Is(err, common.ErrValidation):
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"error": err.Error(),
	})
}
