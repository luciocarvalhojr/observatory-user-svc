// Package middleware provides HTTP middleware for user-svc.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger logs every request with method, path, status, latency, and client IP.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}

// Recovery recovers from panics and returns 500.
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(nil)
}
