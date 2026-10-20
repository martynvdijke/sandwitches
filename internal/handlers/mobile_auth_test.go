package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func setupMobileAuthTest(t *testing.T) (*gin.Engine, *database.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	setupOIDCTestDB(t)
	user := database.User{
		Username: "mobileuser",
		Email:    "mobile@example.com",
		Password: "x",
		IsActive: true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	store := cookie.NewStore([]byte("test-secret-12345678901234567890"))
	r := gin.New()
	r.Use(sessions.Sessions("sandwitches_session", store))
	r.POST("/api/auth/mobile/exchange", MobileAuthExchange)
	r.GET("/session", func(c *gin.Context) {
		s := sessions.Default(c)
		c.JSON(http.StatusOK, gin.H{"user_id": s.Get("user_id"), "auth_method": s.Get("auth_method")})
	})
	return r, &user
}

func postExchange(r *gin.Engine, code string) *httptest.ResponseRecorder {
	form := url.Values{"code": {code}}
	req := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/exchange", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestMobileAuthExchangeCreatesSession(t *testing.T) {
	r, user := setupMobileAuthTest(t)
	code, err := createMobileAuthCode(user.ID)
	if err != nil {
		t.Fatalf("create code: %v", err)
	}
	if len(code) != 64 {
		t.Fatalf("unexpected code length %d", len(code))
	}
	var stored database.MobileAuthCode
	if err := database.DB.Where("user_id = ?", user.ID).First(&stored).Error; err != nil {
		t.Fatalf("stored code: %v", err)
	}
	if stored.CodeHash == code || stored.CodeHash != hashMobileAuthCode(code) {
		t.Fatal("raw code must not be stored")
	}

	w := postExchange(r, code)
	if w.Code != http.StatusOK {
		t.Fatalf("exchange status %d: %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("exchange did not set a session cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/session", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	sessionRes := httptest.NewRecorder()
	r.ServeHTTP(sessionRes, req)
	if sessionRes.Code != http.StatusOK {
		t.Fatalf("session status %d", sessionRes.Code)
	}
	body := sessionRes.Body.String()
	if !strings.Contains(body, `"user_id":`) || !strings.Contains(body, `"auth_method":"oidc"`) {
		t.Fatalf("unexpected session body %s", body)
	}

	// A code is strictly single use.
	replay := postExchange(r, code)
	if replay.Code != http.StatusBadRequest {
		t.Fatalf("replay status %d, want 400", replay.Code)
	}
}

func TestMobileAuthExchangeRejectsExpiredCode(t *testing.T) {
	r, user := setupMobileAuthTest(t)
	code := "expired-code"
	if err := database.DB.Create(&database.MobileAuthCode{
		UserID:    user.ID,
		CodeHash:  hashMobileAuthCode(code),
		ExpiresAt: time.Now().Add(-time.Minute),
	}).Error; err != nil {
		t.Fatalf("create expired code: %v", err)
	}
	w := postExchange(r, code)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expired code status %d, want 400", w.Code)
	}
}

func TestMobileAuthExchangeRejectsInactiveUser(t *testing.T) {
	r, _ := setupMobileAuthTest(t)
	inactive := database.User{
		Username: "inactive",
		Email:    "inactive@example.com",
		Password: "x",
		IsActive: false,
	}
	if err := database.DB.Create(&inactive).Error; err != nil {
		t.Fatalf("create inactive user: %v", err)
	}
	// IsActive has a GORM default of true, so create then flip it.
	if err := database.DB.Model(&inactive).Update("is_active", false).Error; err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	code, err := createMobileAuthCode(inactive.ID)
	if err != nil {
		t.Fatalf("create code: %v", err)
	}
	w := postExchange(r, code)
	if w.Code != http.StatusForbidden {
		t.Fatalf("inactive user status %d, want 403", w.Code)
	}
}

func TestMobileAuthExchangeRejectsUnknownCode(t *testing.T) {
	r, _ := setupMobileAuthTest(t)
	w := postExchange(r, "not-a-real-code")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unknown code status %d, want 400", w.Code)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control %q, want no-store", got)
	}
}
