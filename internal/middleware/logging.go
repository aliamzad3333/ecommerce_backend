package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware returns a logging middleware
func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			"status", statusCode,
			"latency", latency,
			"client_ip", clientIP,
			"method", method,
			"path", path,
			"body_size", bodySize,
		)
	}
}
