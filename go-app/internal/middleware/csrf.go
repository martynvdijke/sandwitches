package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const csrfKey = "csrf_token"

func generateToken() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return hex.EncodeToString(b)
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get(csrfKey)
		if token == nil {
			token = generateToken()
			session.Set(csrfKey, token)
			session.Save()
		}
		tokStr := token.(string)

		c.Set("csrf_token", tokStr)
		c.Set("csrf_token_hidden", template.HTML(`<input type="hidden" name="csrf_token" value="`+tokStr+`">`))

		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
			submitted := c.PostForm("csrf_token")
			if submitted == "" {
				submitted = c.GetHeader("X-CSRF-Token")
			}
			if submitted == "" {
				submitted = c.GetHeader("X-CSRF-TOKEN")
			}
			if submitted != tokStr {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
				return
			}
		}

		c.Next()
	}
}
