package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupOIDCTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	database.DB = db
	_ = db.AutoMigrate(&database.User{}, &database.Group{})
	db.Create(&database.Group{Name: "community"})
	db.Create(&database.Group{Name: "admin"})
}

func TestLoadOIDCConfig(t *testing.T) {
	// valid config
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_ISSUER_URL", "https://authelia.vandijke.xyz")
	t.Setenv("OIDC_CLIENT_ID", "sandwitches")
	t.Setenv("OIDC_CLIENT_SECRET", "s3cret")
	t.Setenv("OIDC_REDIRECT_URL", "https://sandwitches.vandijke.xyz/api/auth/oidc/callback")
	cfg := loadOIDCConfig()
	if !cfg.valid() {
		t.Fatal("expected valid")
	}
	if cfg.LogoutURL != "https://authelia.vandijke.xyz/logout" {
		t.Fatalf("logout url %q", cfg.LogoutURL)
	}
	if len(cfg.Scopes) != 4 {
		t.Fatalf("scopes %v", cfg.Scopes)
	}
	// disabled via false
	t.Setenv("OIDC_ENABLED", "false")
	if loadOIDCConfig().valid() {
		t.Fatal("expected invalid when disabled")
	}
	t.Setenv("OIDC_ENABLED", "0")
	if loadOIDCConfig().valid() {
		t.Fatal("expected invalid when 0")
	}
	// alias OIDC_ISSUER
	t.Setenv("OIDC_ENABLED", "true")
	os.Unsetenv("OIDC_ISSUER_URL")
	t.Setenv("OIDC_ISSUER", "https://authelia.vandijke.xyz")
	if loadOIDCConfig().Issuer != "https://authelia.vandijke.xyz" {
		t.Fatal("alias issuer not read")
	}
	// secret file fallback
	f := t.TempDir() + "/secret"
	os.WriteFile(f, []byte("  fromfile \n"), 0600)
	t.Setenv("OIDC_ISSUER_URL", "https://authelia.vandijke.xyz")
	t.Setenv("OIDC_CLIENT_SECRET", "")
	t.Setenv("OIDC_CLIENT_SECRET_FILE", f)
	if loadOIDCConfig().Secret != "fromfile" {
		t.Fatalf("secret file %q", loadOIDCConfig().Secret)
	}
	// invalid when missing secret
	os.Remove(f)
	t.Setenv("OIDC_CLIENT_SECRET_FILE", f)
	t.Setenv("OIDC_CLIENT_SECRET", "")
	if loadOIDCConfig().valid() {
		t.Fatal("expected invalid without secret")
	}
}

func TestApplyGroupsToUser(t *testing.T) {
	// contains admins -> promote
	u := &database.User{IsStaff: false, IsSuperuser: false}
	if !applyGroupsToUser(u, []string{"admins"}, true) || !u.IsStaff || !u.IsSuperuser {
		t.Fatal("should promote")
	}
	// present but lacks admins -> demote
	u2 := &database.User{IsStaff: true, IsSuperuser: true}
	if !applyGroupsToUser(u2, []string{"users"}, true) || u2.IsStaff || u2.IsSuperuser {
		t.Fatal("should demote")
	}
	// demote when empty groups present
	u3 := &database.User{IsStaff: true, IsSuperuser: true}
	if !applyGroupsToUser(u3, []string{}, true) || u3.IsStaff || u3.IsSuperuser {
		t.Fatal("should demote on empty present")
	}
	// absent claim -> no change
	u4 := &database.User{IsStaff: true, IsSuperuser: true}
	if applyGroupsToUser(u4, nil, false) || !u4.IsStaff {
		t.Fatal("absent should not change")
	}
	u5 := &database.User{IsStaff: false, IsSuperuser: false}
	if applyGroupsToUser(u5, nil, false) || u5.IsStaff {
		t.Fatal("absent should not change 2")
	}
}

func TestOIDCCallbackBadState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOIDCTestDB(t)
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_ISSUER_URL", "https://authelia.vandijke.xyz")
	t.Setenv("OIDC_CLIENT_ID", "sandwitches")
	t.Setenv("OIDC_CLIENT_SECRET", "secret")
	t.Setenv("OIDC_REDIRECT_URL", "https://sandwitches.vandijke.xyz/api/auth/oidc/callback")
	resetOIDCCacheForTest()

	store := cookie.NewStore([]byte("test-secret-12345678901234567890"))
	r := gin.New()
	r.Use(sessions.Sessions("sandwitches_session", store))
	r.GET("/api/auth/oidc/callback", OIDCCallback)

	// No cookies -> should redirect to login?error=oidc_expired
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/auth/oidc/callback?code=abc&state=xyz", nil)
	r.ServeHTTP(w, req)
	if w.Code != 302 {
		t.Fatalf("expected 302 got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc == "" || loc == "/" {
		t.Fatalf("unexpected location %q", loc)
	}

	// Wrong state
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/auth/oidc/callback?code=abc&state=wrong", nil)
	req2.AddCookie(&http.Cookie{Name: "oidc_state", Value: "correct"})
	req2.AddCookie(&http.Cookie{Name: "oidc_nonce", Value: "nonce123"})
	req2.AddCookie(&http.Cookie{Name: "oidc_verifier", Value: "verifier123"})
	r.ServeHTTP(w2, req2)
	if w2.Code != 302 {
		t.Fatalf("expected 302 for bad state got %d", w2.Code)
	}
	if w2.Header().Get("Location") != "/login?error=oidc_state" {
		t.Fatalf("location %q", w2.Header().Get("Location"))
	}
}

func TestLinkOrProvisionOIDCUser(t *testing.T) {
	setupOIDCTestDB(t)
	issuer := "https://authelia.vandijke.xyz"
	sub1 := issuer + "|sub1"
	claims := oidcClaims{Sub: "sub1", Email: "a@example.com", EmailVerified: true, Name: "Alice"}
	u, err := linkOrProvisionOIDCUser(sub1, claims, []string{"admins"}, true)
	if err != nil || !u.IsStaff || !u.IsSuperuser {
		t.Fatalf("provision admins failed %v %+v", err, u)
	}
	// second login without groups claim -> roles untouched
	u2, _ := linkOrProvisionOIDCUser(sub1, claims, nil, false)
	if !u2.IsStaff {
		t.Fatal("absent groups should not demote")
	}
	// present without admins -> demote
	u3, _ := linkOrProvisionOIDCUser(sub1, claims, []string{"users"}, true)
	if u3.IsStaff || u3.IsSuperuser {
		t.Fatal("should demote")
	}
	// link existing email without oidc_sub
	var user2 database.User
	database.DB.Where("email=?", "b@example.com").First(&user2)
	// create manually
	database.DB.Create(&database.User{Username: "bob", Email: "b@example.com", Password: "x", IsActive: true})
	claimsB := oidcClaims{Sub: "sub-b", Email: "b@example.com", EmailVerified: true}
	subB := issuer + "|sub-b"
	ub, err := linkOrProvisionOIDCUser(subB, claimsB, nil, false)
	if err != nil || ub.OIDCSub != subB {
		t.Fatalf("link by email failed %v %+v", err, ub)
	}
}
