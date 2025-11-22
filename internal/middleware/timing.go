package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestTiming() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		cost := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()

		log.Printf("[timing] %s %s -> %d in %v\n", method, path, status, cost)
	}
}
