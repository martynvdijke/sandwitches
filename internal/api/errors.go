package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

// Error codes used in the API error envelope.
const (
	CodeBadRequest   = "bad_request"
	CodeValidation   = "validation_error"
	CodeNotFound     = "not_found"
	CodeUnauthorized = "unauthorized"
	CodeForbidden    = "forbidden"
	CodeRateLimited  = "rate_limited"
	CodeInternal     = "internal"
)

// apiError writes the standard API error envelope:
//
//	{"message": ..., "code": ..., "request_id": ...}
//
// The code is derived from the HTTP status unless a specific one is supplied.
func apiError(c *gin.Context, status int, message string) {
	code := CodeInternal
	switch status {
	case http.StatusBadRequest:
		code = CodeBadRequest
	case http.StatusUnprocessableEntity:
		code = CodeValidation
	case http.StatusNotFound:
		code = CodeNotFound
	case http.StatusUnauthorized:
		code = CodeUnauthorized
	case http.StatusForbidden:
		code = CodeForbidden
	case http.StatusTooManyRequests:
		code = CodeRateLimited
	}
	// A 403 with a sign-in message is really an authentication failure.
	if status == http.StatusForbidden && (message == "Please sign in" || message == "Please sign in first") {
		code = CodeUnauthorized
	}

	c.JSON(status, gin.H{
		"message":    message,
		"code":       code,
		"request_id": middleware.GetRequestID(c),
	})
}
