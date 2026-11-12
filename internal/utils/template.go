package utils

import (
	"html/template"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

var (
	notifMu     sync.Mutex
	notifValue  bool
	notifLoaded time.Time
)

type TemplateData struct {
	User       *database.User
	Flashes    []FlashMessage
	CSRFToken  string
	CSRFHidden template.HTML
	Error      string
	EinkMode   bool
	// Analytics is true only when Umami is configured and this visitor has not
	// opted out. Templates use it to decide whether to emit the Umami script.
	Analytics bool
	// NotificationsEnabled mirrors the site-level master switch so the base
	// layout can gate the browser notification client.
	NotificationsEnabled bool
	Data                 gin.H
}

// AnalyticsEnabled reports whether the optional Umami analytics script is
// configured for this deployment. It is self-hosted and cookieless; consent is
// not required, but users can still object in Settings.
func AnalyticsEnabled() bool {
	return strings.TrimSpace(os.Getenv("UMAMI_HOST")) != "" &&
		strings.TrimSpace(os.Getenv("UMAMI_WEBSITE_ID")) != ""
}

// IsEinkMode detects e-ink mode: URL param ?eink=1 > cookie eink_mode=1 > user theme.
func IsEinkMode(c *gin.Context, user *database.User) bool {
	if c.Query("eink") == "1" {
		return true
	}
	if cookie, err := c.Cookie("eink_mode"); err == nil && cookie == "1" {
		return true
	}
	return user != nil && user.Theme == "eink"
}

// NotificationsEnabled reports whether the site-level notification master
// switch is on. The value is cached briefly because it is read on every page
// render; a nil or uninitialised database reads as disabled.
func NotificationsEnabled() bool {
	notifMu.Lock()
	defer notifMu.Unlock()
	if !notifLoaded.IsZero() && time.Since(notifLoaded) < 30*time.Second {
		return notifValue
	}
	if database.DB == nil {
		return notifValue
	}
	var s database.Setting
	if err := database.DB.Select("notifications_enabled").First(&s).Error; err != nil {
		return notifValue
	}
	notifValue = s.NotificationsEnabled
	notifLoaded = time.Now()
	return notifValue
}

func NewTemplateData(c *gin.Context) *TemplateData {
	td := &TemplateData{
		User:       middleware.GetUser(c),
		Flashes:    GetFlashes(c),
		CSRFToken:  "",
		CSRFHidden: "",
		Data:       make(gin.H),
	}
	td.EinkMode = IsEinkMode(c, td.User)
	td.Analytics = AnalyticsEnabled() && (td.User == nil || !td.User.AnalyticsOptOut)
	td.NotificationsEnabled = NotificationsEnabled()
	if tok, exists := c.Get("csrf_token"); exists {
		td.CSRFToken = tok.(string)
		td.CSRFHidden = template.HTML(`<input type="hidden" name="csrf_token" value="` + tok.(string) + `">`)
	}
	return td
}

func (td *TemplateData) With(key string, value interface{}) *TemplateData {
	td.Data[key] = value
	return td
}

func (td *TemplateData) ToGinH() gin.H {
	h := gin.H{
		"user":    td.User,
		"flashes": td.Flashes,
	}
	if td.CSRFHidden != "" {
		h["csrf_token"] = td.CSRFHidden
		h["csrf_token_value"] = td.CSRFToken
	}
	if td.Error != "" {
		h["error"] = td.Error
	}
	h["eink_mode"] = td.EinkMode
	h["analytics"] = td.Analytics
	h["notifications_enabled"] = td.NotificationsEnabled
	for k, v := range td.Data {
		h[k] = v
	}
	return h
}
