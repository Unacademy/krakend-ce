package krakend

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CaptureBodyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		// Restore the io.ReadCloser to the original state
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// Store the body in context for later use
		c.Set("rawBody", bodyBytes)
		c.Next()
	}
}
