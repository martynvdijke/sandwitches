package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

const (
	mobileAuthCodeTTL = 2 * time.Minute
	// The mobile app registers this scheme/host in its AndroidManifest.xml.
	mobileAuthCallbackURL = "sandwitches://auth"
)

var (
	errMobileCodeInvalid  = errors.New("invalid mobile auth code")
	errMobileUserInactive = errors.New("inactive user")
)

// isMobileOIDC reports whether the current browser flow was started by the
// mobile app. The flag is an HttpOnly cookie set by OIDCLogin and is scoped to
// the same five minute window as the other OIDC cookies.
func isMobileOIDC(c *gin.Context) bool {
	v, err := c.Cookie("oidc_mobile")
	return err == nil && v == "1"
}

func hashMobileAuthCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// createMobileAuthCode issues a one-time code for userID. The raw code is only
// returned to the caller and is never persisted.
func createMobileAuthCode(userID uint) (string, error) {
	code, err := oidcRandHex(32)
	if err != nil {
		return "", err
	}
	now := time.Now()
	// Opportunistic cleanup; failure here must not block a login.
	_ = database.DB.Where("expires_at < ?", now).Delete(&database.MobileAuthCode{}).Error
	row := database.MobileAuthCode{
		UserID:    userID,
		CodeHash:  hashMobileAuthCode(code),
		CreatedAt: now,
		ExpiresAt: now.Add(mobileAuthCodeTTL),
	}
	if err := database.DB.Create(&row).Error; err != nil {
		return "", err
	}
	return code, nil
}

// consumeMobileAuthCode atomically marks a code as used and returns its user.
// A code can be redeemed at most once, even if two requests race.
func consumeMobileAuthCode(code string) (*database.User, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errMobileCodeInvalid
	}
	now := time.Now()
	hash := hashMobileAuthCode(code)
	res := database.DB.Model(&database.MobileAuthCode{}).
		Where("code_hash = ? AND used_at IS NULL AND expires_at > ?", hash, now).
		Update("used_at", now)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, errMobileCodeInvalid
	}
	var row database.MobileAuthCode
	if err := database.DB.Where("code_hash = ?", hash).First(&row).Error; err != nil {
		return nil, errMobileCodeInvalid
	}
	var user database.User
	if err := database.DB.First(&user, row.UserID).Error; err != nil {
		return nil, errMobileCodeInvalid
	}
	if !user.IsActive {
		return nil, errMobileUserInactive
	}
	return &user, nil
}

// MobileAuthExchange trades a one-time OIDC handoff code for a normal session
// cookie. It is intentionally unauthenticated: the code is the proof of login.
func MobileAuthExchange(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var req struct {
		Code string `json:"code" form:"code"`
	}
	if err := c.ShouldBind(&req); err != nil || strings.TrimSpace(req.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
		return
	}
	user, err := consumeMobileAuthCode(req.Code)
	if err != nil {
		if errors.Is(err, errMobileUserInactive) {
			c.JSON(http.StatusForbidden, gin.H{"error": "inactive_user"})
			return
		}
		if errors.Is(err, errMobileCodeInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	session := sessions.Default(c)
	session.Clear()
	session.Set("user_id", user.ID)
	session.Set("auth_method", "oidc")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
