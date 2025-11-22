package middleware

import (
	"strconv"
	"time"

	"minifeed/internal/metrics"

	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		method := c.Request.Method

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		statusCode := c.Writer.Status()
		statusStr := strconv.Itoa(statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
