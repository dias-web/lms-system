package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger logs every HTTP request via logrus so all output shares one
// formatter and sink. Successful requests log at DEBUG (5xx → ERROR, 4xx → WARN).
func RequestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		entry := log.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency_ms": time.Since(start).Milliseconds(),
			"client_ip":  c.ClientIP(),
		})

		switch status := c.Writer.Status(); {
		case status >= 500:
			entry.Error("http request")
		case status >= 400:
			entry.Warn("http request")
		default:
			entry.Debug("http request")
		}
	}
}
