package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	traceIDContextKey = "trace_id"
	traceIDHeaderKey  = "X-Trace-Id"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := strings.TrimSpace(c.GetHeader(traceIDHeaderKey))
		if traceID == "" {
			traceID = generateTraceID()
		}

		c.Set(traceIDContextKey, traceID)
		c.Writer.Header().Set(traceIDHeaderKey, traceID)
		c.Next()
	}
}

func GetTraceID(c *gin.Context) string {
	value, exists := c.Get(traceIDContextKey)
	if !exists {
		return ""
	}

	traceID, ok := value.(string)
	if !ok {
		return ""
	}
	return traceID
}

func generateTraceID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
