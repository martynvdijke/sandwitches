package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin access for the API. originsCSV is a
// comma-separated allow-list of origins (e.g. "https://app.example,https://dev.example").
// An empty string disables CORS entirely (no headers are emitted).
func CORS(originsCSV string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range strings.Split(originsCSV, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		permitted := origin != "" && (allowed["*"] || allowed[origin])
		if permitted {
			if allowed["*"] {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")

			// Preflight request.
			if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
				c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
				c.Header("Access-Control-Max-Age", "86400")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}
		c.Next()
	}
}
