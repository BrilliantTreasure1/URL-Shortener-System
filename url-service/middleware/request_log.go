package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		if status >= http.StatusBadRequest {
			log.Printf("gin: %d %s %s %s", status, c.Request.Method, c.Request.URL.Path, time.Since(start))
		}
	}
}