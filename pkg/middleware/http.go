package middleware

import (
	"io"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func Access() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		level := slog.LevelInfo
		if c.Writer.Status() >= 500 {
			level = slog.LevelError
		} else if c.Writer.Status() >= 400 {
			level = slog.LevelWarn
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		// Do not log query strings, request bodies, cookies or authorization headers.
		slog.Default().Log(c.Request.Context(), level, "http request",
			"method", c.Request.Method, "path", path, "status", c.Writer.Status(),
			"duration_ms", float64(time.Since(started).Microseconds())/1000, "client_ip", c.ClientIP())
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		slog.Error("http panic", "panic", recovered, "stack", string(debug.Stack()))
		c.AbortWithStatus(500)
	})
}
