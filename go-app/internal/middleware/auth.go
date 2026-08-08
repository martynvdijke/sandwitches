package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
			c.Abort()
			return
		}

		var user database.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			session.Clear()
			_ = session.Save()
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
			c.Abort()
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
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
			c.Abort()
			return
		}

		var user database.User
		if err := database.DB.First(&user, userID).Error; err != nil || !user.IsStaff {
			c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
			c.Abort()
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
