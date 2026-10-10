package response

import "github.com/gin-gonic/gin"

// WriteSuccess writes a JSON success response with the specified HTTP status code.
func WriteSuccess(c *gin.Context, statusCode int, data any) {
	if data == nil {
		c.Status(statusCode)
		c.Writer.WriteHeaderNow()
		return
	}
	c.JSON(statusCode, data)
}
