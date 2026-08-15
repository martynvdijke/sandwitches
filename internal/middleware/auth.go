package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

// isAPIRequest reports whether the request targets the JSON API (mounted
// under /api), as opposed to a web page.
func isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/")
}

// unauthorized aborts with a 401 JSON body for API requests (matching the
// v2.x django-ninja behavior) or a 302 login redirect for web requests.
func unauthorized(c *gin.Context) {
	if isAPIRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Please sign in first"})
	} else {
		c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
	}
	c.Abort()
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			unauthorized(c)
			return
		}

		var user database.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			session.Clear()
			_ = session.Save()
			unauthorized(c)
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func StaffRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			unauthorized(c)
			return
		}

		var user database.User
		if err := database.DB.First(&user, userID).Error; err != nil || !user.IsStaff {
			unauthorized(c)
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID != nil {
			var user database.User
			if err := database.DB.First(&user, userID).Error; err == nil {
				c.Set("user", &user)
			}
		}
		c.Next()
	}
}

func GetUser(c *gin.Context) *database.User {
	if u, exists := c.Get("user"); exists {
		if user, ok := u.(*database.User); ok {
			return user
		}
	}
	return nil
}
