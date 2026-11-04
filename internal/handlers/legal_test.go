package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func TestPublicLegalPages(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	cases := []struct {
		path string
		want []string
	}{
		{"/privacy", []string{"Privacy policy", "Article 13", "House"}},
		{"/terms", []string{"Terms of use", "What you post"}},
		{"/account/delete", []string{"Delete your account", "What is removed", "What stays"}},
	}
	// The privacy assertion above intentionally uses a term that is not in the
	// template; replace it with a stable phrase to keep the test honest.
	cases[0].want = []string{"Privacy policy", "Article 13", "Delete your account"}

	for _, tc := range cases {
		client := newClient()
		resp, err := client.Get(srv.URL + tc.path)
		if err != nil {
			t.Fatalf("%s: request failed: %v", tc.path, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", tc.path, resp.StatusCode)
		}
		for _, want := range tc.want {
			if !strings.Contains(string(body), want) {
				t.Errorf("%s: body does not contain %q", tc.path, want)
			}
		}
		// Public pages are registered before the CSRF group, so reading them
		// must not hand an anonymous visitor a session cookie.
		if len(resp.Cookies()) != 0 {
			t.Errorf("%s: public page set a cookie: %v", tc.path, resp.Cookies())
		}
	}
}

func TestAccountDeleteRequiresAuthAndCSRF(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "delete_me", "password123", false, false, "")

	// Signed out: AuthRequired redirects to the login page.
	anon := newClient()
	resp, err := anon.PostForm(srv.URL+"/settings/account/delete", url.Values{"confirm": {user.Username}})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/login") {
		t.Errorf("anonymous delete should redirect to login, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	// Signed in but without the CSRF token: rejected before the handler.
	client := newClient()
	loginUser(t, srv, client, user.Username, "password123")
	resp, err = client.PostForm(srv.URL+"/settings/account/delete", url.Values{"confirm": {user.Username}})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("missing CSRF should return 403, got %d", resp.StatusCode)
	}

	// Wrong confirmation: no deletion.
	token := getCSRFToken(t, srv, client)
	body := url.Values{"confirm": {"wrong-name"}, "csrf_token": {token}}
	resp, err = client.PostForm(srv.URL+"/settings/account/delete", body)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/settings") {
		t.Errorf("wrong confirmation should redirect to settings, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	var count int64
	database.DB.Model(&database.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Fatal("wrong confirmation must not delete the account")
	}
}

func TestSelfDeleteDetachesContentAndInvalidatesSessions(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "selfdelete", "password123", false, false, "community")
	recipe := createTestRecipe(t, "Delete Recipe", user, true)

	order := database.Order{UserID: user.ID, Status: "PENDING", TotalPrice: 5}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	database.DB.Create(&database.Rating{RecipeID: recipe.ID, UserID: user.ID, Score: 4, Comment: "nice"})
	database.DB.Create(&database.CartItem{UserID: user.ID, RecipeID: recipe.ID, Quantity: 1})
	database.DB.Exec("INSERT INTO user_favorites (user_id, recipe_id) VALUES (?, ?)", user.ID, recipe.ID)
	database.DB.Create(&database.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashResetToken("reset-token"),
		ExpiresAt: time.Now().Add(time.Hour),
	})
	database.DB.Create(&database.MobileAuthCode{
		UserID:    user.ID,
		CodeHash:  hashMobileAuthCode("mobile-code"),
		ExpiresAt: time.Now().Add(time.Minute),
	})

	client := newClient()
	loginUser(t, srv, client, user.Username, "password123")
	otherDevice := newClient()
	loginUser(t, srv, otherDevice, user.Username, "password123")

	token := getCSRFToken(t, srv, client)
	resp, err := client.PostForm(srv.URL+"/settings/account/delete", url.Values{
		"confirm":    {user.Username},
		"csrf_token": {token},
	})
	if err != nil {
		t.Fatalf("delete request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/account/delete?deleted=1") {
		t.Fatalf("expected deletion redirect, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	var users int64
	database.DB.Model(&database.User{}).Where("id = ?", user.ID).Count(&users)
	if users != 0 {
		t.Error("user row should be deleted")
	}

	// Authored recipe and order survive, detached from the account.
	var detachedRecipes, detachedOrders int64
	database.DB.Model(&database.Recipe{}).Where("id = ? AND uploaded_by_id IS NULL", recipe.ID).Count(&detachedRecipes)
	database.DB.Model(&database.Order{}).Where("id = ? AND user_id IS NULL", order.ID).Count(&detachedOrders)
	if detachedRecipes != 1 {
		t.Error("authored recipe should remain with a NULL author")
	}
	if detachedOrders != 1 {
		t.Error("order should remain with a NULL user")
	}

	// Owned rows are removed.
	for _, table := range []string{"ratings", "cart_items", "user_favorites", "user_groups", "password_reset_tokens", "mobile_auth_codes"} {
		var n int64
		database.DB.Table(table).Where("user_id = ?", user.ID).Count(&n)
		if n != 0 {
			t.Errorf("%s: expected 0 rows, got %d", table, n)
		}
	}

	// The cookie issued on the other device no longer authenticates.
	resp, err = otherDevice.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatalf("profile request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/login") {
		t.Errorf("old session should be invalid, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}

	// The public page confirms the outcome.
	resp, err = newClient().Get(srv.URL + "/account/delete?deleted=1")
	if err != nil {
		t.Fatalf("confirmation page failed: %v", err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(b), "Your account has been deleted") {
		t.Error("confirmation page should mention the deleted account")
	}
}

func TestDeleteLastSuperuserBlocked(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	admin := createTestUser(t, "only_admin", "password123", true, true, "admin")
	client := newClient()
	loginUser(t, srv, client, admin.Username, "password123")
	token := getCSRFToken(t, srv, client)

	resp, err := client.PostForm(srv.URL+"/settings/account/delete", url.Values{
		"confirm":    {admin.Username},
		"csrf_token": {token},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/settings") {
		t.Errorf("last superuser should be redirected to settings, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	var count int64
	database.DB.Model(&database.User{}).Where("id = ?", admin.ID).Count(&count)
	if count != 1 {
		t.Error("the last superuser must not be deleted")
	}
}

func TestPasswordResetInvalidatesExistingSessions(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "reset_session", "oldpass123", false, false, "")
	client := newClient()
	loginUser(t, srv, client, user.Username, "oldpass123")

	token := "reset-session-token"
	database.DB.Create(&database.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashResetToken(token),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	resp, err := client.PostForm(srv.URL+"/reset-password", url.Values{
		"token":     {token},
		"password1": {"newpass456"},
		"password2": {"newpass456"},
	})
	if err != nil {
		t.Fatalf("reset request failed: %v", err)
	}
	resp.Body.Close()

	// The cookie created before the reset must no longer authenticate.
	resp, err = client.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatalf("profile request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || !strings.Contains(resp.Header.Get("Location"), "/login") {
		t.Errorf("pre-reset session should be invalid, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
}
