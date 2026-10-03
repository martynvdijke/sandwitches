package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

func setupSMTPTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	tmp := t.TempDir()
	database.Init(&config.Config{DatabaseFile: tmp + "/test.db", Debug: false, SecretKey: "testkey", LanguageCode: "en"})
	SetMediaRoot(tmp + "/media")
	r := gin.New()
	store := cookie.NewStore([]byte("testkey"))
	r.Use(sessions.Sessions("sandwitches_session", store))
	return r
}

func loginStaff(t *testing.T, r *gin.Engine) *http.Cookie {
	t.Helper()
	pw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	u := database.User{Username: "admin", Password: string(pw), Email: "admin@example.com", IsStaff: true, IsActive: true, Language: "en", Theme: "light"}
	database.DB.Create(&u)
	// login via session
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	r2 := gin.New()
	store := cookie.NewStore([]byte("testkey"))
	r2.Use(sessions.Sessions("sandwitches_session", store))
	// Instead do direct handler: create request to login would be complex; manually create session cookie
	// Use a test server approach: create a route that sets session
	r2.GET("/set", func(c2 *gin.Context) {
		s := sessions.Default(c2)
		s.Set("user_id", u.ID)
		s.Save()
		c2.String(200, "ok")
	})
	req, _ := http.NewRequest("GET", "/set", nil)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, req)
	cookies := w2.Result().Cookies()
	for _, ck := range cookies {
		if ck.Name == "sandwitches_session" {
			return ck
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func TestAdminSettingsSMTPSave(t *testing.T) {
	r := setupSMTPTestRouter(t)
	r.Use(middleware.OptionalAuth())
	r.POST("/dashboard/settings", middleware.StaffRequired(), AdminSettings)
	ck := loginStaff(t, r)

	form := url.Values{}
	form.Set("site_name", "Test")
	form.Set("email", "site@example.com")
	form.Set("log_level", "INFO")
	form.Set("smtp_enabled", "on")
	form.Set("smtp_host", "smtp.example.com")
	form.Set("smtp_port", "587")
	form.Set("smtp_user", "user1")
	form.Set("smtp_password", "secret123")
	form.Set("smtp_from_email", "from@example.com")
	form.Set("smtp_tls", "on")

	req, _ := http.NewRequest("POST", "/dashboard/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(ck)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect got %d body %s", w.Code, w.Body.String())
	}
	var s database.Setting
	database.DB.First(&s)
	if s.SMTPHost != "smtp.example.com" || s.SMTPPort != "587" || s.SMTPUser != "user1" || s.SMTPPassword != "secret123" || s.SMTPFromEmail != "from@example.com" {
		t.Fatalf("SMTP not saved: %+v", s)
	}
	if !s.SMTPEnabled || !s.SMTPTLS {
		t.Fatal("bools not saved")
	}

	// blank password keeps existing
	form2 := url.Values{}
	form2.Set("site_name", "Test2")
	form2.Set("email", "site@example.com")
	form2.Set("log_level", "INFO")
	form2.Set("smtp_enabled", "on")
	form2.Set("smtp_host", "smtp.example.com")
	form2.Set("smtp_port", "587")
	form2.Set("smtp_user", "user1")
	form2.Set("smtp_password", "")
	form2.Set("smtp_from_email", "from@example.com")
	req2, _ := http.NewRequest("POST", "/dashboard/settings", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(ck)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	var s2 database.Setting
	database.DB.First(&s2)
	if s2.SMTPPassword != "secret123" {
		t.Fatalf("blank password should keep existing, got %q", s2.SMTPPassword)
	}
}
