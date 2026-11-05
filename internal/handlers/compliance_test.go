package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
)

func TestSignupRequiresTermsAcceptance(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	client := newClient()
	body := url.Values{
		"username":  {"no_consent"},
		"email":     {"no_consent@example.com"},
		"password1": {"password123"},
		"password2": {"password123"},
	}
	resp, err := client.PostForm(srv.URL+"/signup", body)
	if err != nil {
		t.Fatalf("signup request failed: %v", err)
	}
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected re-rendered form 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(raw), "You must accept the Terms of use and Privacy policy.") {
		t.Error("missing consent error message")
	}
	var count int64
	database.DB.Model(&database.User{}).Where("username = ?", "no_consent").Count(&count)
	if count != 0 {
		t.Error("account was created without consent")
	}
}

func TestSignupWithConsentStoresAcceptedTermsAt(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	client := newClient()
	body := url.Values{
		"username":     {"with_consent"},
		"email":        {"with_consent@example.com"},
		"password1":    {"password123"},
		"password2":    {"password123"},
		"accept_terms": {"true"},
	}
	resp, err := client.PostForm(srv.URL+"/signup", body)
	if err != nil {
		t.Fatalf("signup request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect after signup, got %d", resp.StatusCode)
	}

	var user database.User
	if err := database.DB.Where("username = ?", "with_consent").First(&user).Error; err != nil {
		t.Fatalf("user not created: %v", err)
	}
	if user.AcceptedTermsAt == nil {
		t.Error("AcceptedTermsAt was not stored")
	}
}

func TestAnalyticsOptOutPersistsAndSuppressesTemplateFlag(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	t.Setenv("UMAMI_HOST", "https://umami.example.com")
	t.Setenv("UMAMI_WEBSITE_ID", "site-id")

	user := createTestUser(t, "analytics_user", "password123", false, false, "")
	client := newClient()
	loginUser(t, srv, client, user.Username, "password123")
	token := getCSRFToken(t, srv, client)

	resp, err := client.PostForm(srv.URL+"/settings", url.Values{
		"language":          {"en"},
		"theme":             {"light"},
		"analytics_opt_out": {"on"},
		"csrf_token":        {token},
	})
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	resp.Body.Close()

	var stored database.User
	if err := database.DB.First(&stored, user.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.AnalyticsOptOut {
		t.Fatal("AnalyticsOptOut was not persisted")
	}

	// With Umami configured, the template flag is suppressed for this user.
	if td := templateDataForUser(t, &stored); td.Analytics {
		t.Error("analytics flag should be suppressed for an opted-out user")
	}

	// A user who has not opted out still gets analytics when configured.
	other := database.User{Username: "other", IsActive: true}
	if td := templateDataForUser(t, &other); !td.Analytics {
		t.Error("analytics flag should be true for a user who did not opt out")
	}
}

// templateDataForUser runs NewTemplateData inside a context with a session
// store, because template data also reads flash messages from the session.
func templateDataForUser(t *testing.T, user *database.User) *utils.TemplateData {
	t.Helper()
	var td *utils.TemplateData
	r := gin.New()
	r.Use(sessions.Sessions("sandwitches_session", cookie.NewStore([]byte("test-key"))))
	r.GET("/", func(c *gin.Context) {
		c.Set("user", user)
		td = utils.NewTemplateData(c)
		c.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if td == nil {
		t.Fatal("template data was not built")
	}
	return td
}

func TestDemoAccountDeletionIsRejected(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	demoUser := createTestUser(t, "demo", "demo", false, false, "")
	if err := database.DB.Model(demoUser).Update("is_demo", true).Error; err != nil {
		t.Fatal(err)
	}

	client := newClient()
	loginUser(t, srv, client, demoUser.Username, "demo")
	token := getCSRFToken(t, srv, client)

	resp, err := client.PostForm(srv.URL+"/settings/account/delete", url.Values{
		"confirm":    {demoUser.Username},
		"csrf_token": {token},
	})
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/settings") {
		t.Fatalf("demo delete should redirect to settings, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	var count int64
	database.DB.Model(&database.User{}).Where("id = ?", demoUser.ID).Count(&count)
	if count != 1 {
		t.Error("demo account must not be deleted")
	}
}
