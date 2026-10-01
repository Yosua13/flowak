package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultTimelinePageSize = 50
const maxTimelinePageSize = 100

// timelineCursor is intentionally opaque. Pairing the timestamp with the ID
// gives a stable order when several records are created in the same instant.
type timelineCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func decodeTimelineCursor(c *gin.Context) (timelineCursor, bool) {
	raw := c.Query("cursor")
	if raw == "" {
		return timelineCursor{}, true
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	var cursor timelineCursor
	if err != nil || json.Unmarshal(data, &cursor) != nil || cursor.ID == "" || cursor.CreatedAt.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor", "error_code": "validation_error"})
		return timelineCursor{}, false
	}
	return cursor, true
}

func encodeTimelineCursor(createdAt time.Time, id string) string {
	data, _ := json.Marshal(timelineCursor{CreatedAt: createdAt, ID: id})
	return base64.RawURLEncoding.EncodeToString(data)
}

func timelineLimit(c *gin.Context) (int, bool) {
	limit := defaultTimelinePageSize
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxTimelinePageSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100", "error_code": "validation_error"})
			return 0, false
		}
		limit = parsed
	}
	return limit, true
}
