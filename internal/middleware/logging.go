package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	// Keep a copy but cap to avoid huge logs
	const max = 2048
	if w.buf.Len() < max {
		remain := max - w.buf.Len()
		chunk := b
		if len(chunk) > remain {
			chunk = chunk[:remain]
		}
		w.buf.Write(chunk)
	}
	return w.ResponseWriter.Write(b)
}

// Logging returns structured access log
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// read small body for debug (for LINE webhook it’s small JSON)
		var reqBodyPreview string
		if c.Request.Body != nil {
			const max = 2048
			b, _ := io.ReadAll(io.LimitReader(c.Request.Body, max))
			reqBodyPreview = string(b)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b)) // restore
		}

		blw := &bodyLogWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		reqID, _ := c.Get(ContextKeyRequestID)
		traceID, _ := c.Get(ContextKeyTraceParent)

		// Collect selected headers (masked)
		lineSig := c.Request.Header.Get("X-Line-Signature")
		if lineSig != "" {
			lineSig = mask(lineSig)
		}

		fields := []zap.Field{
			zap.String("req_id", toStr(reqID)),
			zap.String("trace_id", toStr(traceID)),
			zap.String("remote_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),      // matched route
			zap.String("url", c.Request.URL.Path), // raw path
			zap.Int("status", status),
			zap.Int("resp_size", size),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("line_signature", lineSig),
		}

		// attach request/response previews in dev environment only
		if gin.Mode() != gin.ReleaseMode {
			fields = append(fields,
				zap.String("req_body_preview", reqBodyPreview),
				zap.String("resp_body_preview", blw.buf.String()),
			)
		}

		if len(c.Errors) > 0 {
			logger.Error("request_error",
				append(fields, zap.String("error", c.Errors.String()))...,
			)
		} else {
			logger.Info("request_complete", fields...)
		}
	}
}

func toStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
