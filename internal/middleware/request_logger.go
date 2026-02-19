package middleware

import (
	"Sparrow/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		traceID := GetTraceID(c)
		latency := time.Since(start)
		status := c.Writer.Status()
		routePath := c.FullPath()
		if routePath == "" {
			routePath = c.Request.URL.Path
		}

		fields := []zap.Field{
			zap.String("traceId", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", routePath),
			zap.String("uri", c.Request.RequestURI),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("clientIP", c.ClientIP()),
			zap.Int("size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case status >= 500:
			utils.Log.Error("request completed", fields...)
		case status >= 400:
			utils.Log.Warn("request completed", fields...)
		default:
			utils.Log.Info("request completed", fields...)
		}
	}
}
