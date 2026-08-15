package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID ensures every request carries an X-Request-Id: it uses the client
// supplied value when present, otherwise generates a UUID. The id is stored in
// the request context (for error envelopes and access logging) and echoed back
// on the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-Id", id)
		c.Next()
	}
}

// GetRequestID returns the request id stored by the RequestID middleware, or
// an empty string when the middleware is not in the chain.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
