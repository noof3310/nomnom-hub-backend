package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID       = "X-Request-Id"
	HeaderCloudTrace      = "X-Cloud-Trace-Context"
	ContextKeyRequestID   = "req_id"
	ContextKeyTraceParent = "trace_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// prefer Google trace header: X-Cloud-Trace-Context: TRACE_ID/SPAN_ID;o=TRACE_TRUE
		trace := c.GetHeader(HeaderCloudTrace)
		if trace != "" {
			parts := strings.SplitN(trace, "/", 2)
			if len(parts) > 0 && parts[0] != "" {
				c.Set(ContextKeyTraceParent, parts[0])
			}
		}
		reqID := c.GetHeader(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Writer.Header().Set(HeaderRequestID, reqID)
		c.Set(ContextKeyRequestID, reqID)
		c.Next()
	}
}
