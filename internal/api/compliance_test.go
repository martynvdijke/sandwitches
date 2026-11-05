package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func TestAuthConfigShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	SetDemoCredentials(true, "demo", "demo")
	t.Cleanup(func() { SetDemoCredentials(false, "", "") })

	w := apiGet(t, r, "/api/v1/auth/config", "")
	if w.Code != http.StatusOK {
		t.Fatalf("auth config = %d: %s", w.Code, w.Body.String())
	}
	var payload struct {
		Demo struct {
			Enabled  bool   `json:"enabled"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"demo"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Demo.Enabled || payload.Demo.Username != "demo" || payload.Demo.Password != "demo" {
		t.Errorf("unexpected demo config: %+v", payload.Demo)
	}
}

func TestAuthConfigDisabledHidesCredentials(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	SetDemoCredentials(false, "demo", "secret")
	t.Cleanup(func() { SetDemoCredentials(false, "", "") })

	w := apiGet(t, r, "/api/v1/auth/config", "")
	var payload map[string]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	demo := payload["demo"]
	if demo == nil {
		t.Fatal("demo object missing")
	}
	if demo["enabled"] != false {
		t.Error("demo.enabled should be false")
	}
	if _, ok := demo["username"]; ok {
		t.Error("disabled demo must not expose a username")
	}
	if _, ok := demo["password"]; ok {
		t.Error("disabled demo must not expose a password")
	}
}

func TestDemoAccountDeleteConflict(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	user := createAPIUser(t, "demo", false)
	if err := database.DB.Model(user).Update("is_demo", true).Error; err != nil {
		t.Fatal(err)
	}
	cookie := loginAPIUser(t, r, user.Username)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/me", strings.NewReader(`{"confirm":"demo"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "sandwitches_session", Value: cookie})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 for demo deletion, got %d: %s", w.Code, w.Body.String())
	}

	var count int64
	database.DB.Model(&database.User{}).Where("id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Error("demo account was deleted")
	}
}
