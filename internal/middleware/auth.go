package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

// SessionCookieName is the signed cookie that carries the whole session.
const SessionCookieName = "sandwitches_session"

// sessionVersionKey stores User.SessionVersion in the client-side session.
const sessionVersionKey = "session_version"

// isAPIRequest reports whether the request targets the JSON API (mounted
// under /api), as opposed to a web page.
func isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/")
}

// unauthorized aborts with a 401 JSON body for API requests (matching the
// v2.x django-ninja behavior) or a 302 login redirect for web requests.
func unauthorized(c *gin.Context) {
	if isAPIRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":    "Please sign in first",
			"code":       "unauthorized",
			"request_id": GetRequestID(c),
		})
	} else {
		c.Redirect(http.StatusFound, "/login?next="+c.Request.URL.Path)
	}
	c.Abort()
}

// CurrentSessionVersion returns the version that must be embedded in a session
// cookie for user. Users created before the column existed read as 1, which is
// also what the database default applies to new rows.
func CurrentSessionVersion(user *database.User) int {
	if user == nil || user.SessionVersion < 1 {
		return 1
	}
	return user.SessionVersion
}

// SetSessionUser binds user to the signed cookie. Every sign-in path must use
// this instead of setting user_id alone, otherwise the cookie would not carry
// the version that invalidates it later.
func SetSessionUser(session sessions.Session, user *database.User) {
	session.Set("user_id", user.ID)
	session.Set(sessionVersionKey, CurrentSessionVersion(user))
}

// sessionVersionMatches rejects cookies issued before the account's version was
// bumped. A cookie without a version is accepted only for accounts that have
// never been bumped (version 1) so that sessions issued before this feature
// existed survive the upgrade; the first password change or deletion makes
// every old cookie invalid.
func sessionVersionMatches(raw interface{}, user *database.User) bool {
	want := CurrentSessionVersion(user)
	if raw == nil {
		return want <= 1
	}
	got, ok := raw.(int)
	if !ok {
		return false
	}
	return got == want
}

// loadSessionUser resolves the cookie to a live, active account and verifies
// that the cookie predates no version bump.
func loadSessionUser(c *gin.Context) *database.User {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		return nil
	}

	var user database.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		session.Clear()
		_ = session.Save()
		return nil
	}
	if !user.IsActive || !sessionVersionMatches(session.Get(sessionVersionKey), &user) {
		session.Clear()
		_ = session.Save()
		return nil
	}
	return &user
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := loadSessionUser(c)
		if user == nil {
			unauthorized(c)
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func StaffRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := loadSessionUser(c)
		if user == nil || !user.IsStaff {
			unauthorized(c)
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user := loadSessionUser(c); user != nil {
			c.Set("user", user)
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
