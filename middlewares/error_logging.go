package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		status := c.Writer.Status()
		// File serving writes HTTP errors directly without adding them to c.Errors.
		// API errors are recorded by their handlers; leave unmatched API routes alone.
		if status >= http.StatusBadRequest && len(c.Errors) == 0 && !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Error(fmt.Errorf("file serving failed: %s", http.StatusText(status)))
		}

		if status != http.StatusOK && status != http.StatusNoContent {
			if len(c.Errors) > 0 {
				for _, err := range c.Errors {
					fmt.Printf("[ERR] %s | %3d | %s %s | %s\n",
						time.Now().Format("2006/01/02 - 15:04:05"),
						status,
						c.Request.Method,
						c.Request.URL.Path,
						err.Err,
					)
				}
			}
		}
	}
}
