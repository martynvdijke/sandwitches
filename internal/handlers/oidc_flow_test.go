package handlers

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// fakeOIDCProvider is a minimal OIDC provider used to exercise the complete
// browser handoff without an external identity provider.
type fakeOIDCProvider struct {
	server   *httptest.Server
	key      *rsa.PrivateKey
	clientID string
	mu       sync.Mutex
	nonces   map[string]string
}

func newFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	f := &fakeOIDCProvider{key: key, clientID: "sandwitches", nonces: map[string]string{}}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)
	return f
}

func b64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func (f *fakeOIDCProvider) handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                f.server.URL,
			"authorization_endpoint":                f.server.URL + "/authorize",
			"token_endpoint":                        f.server.URL + "/token",
			"jwks_uri":                              f.server.URL + "/keys",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"scopes_supported":                      []string{"openid", "email", "profile", "groups"},
		})
	case "/keys":
		e := big.NewInt(int64(f.key.PublicKey.E)).Bytes()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kty": "RSA",
				"kid": "test",
				"use": "sig",
				"alg": "RS256",
				"n":   b64url(f.key.PublicKey.N.Bytes()),
				"e":   b64url(e),
			}},
		})
	case "/authorize":
		q := r.URL.Query()
		code, _ := oidcRandHex(8)
		f.mu.Lock()
		f.nonces[code] = q.Get("nonce")
		f.mu.Unlock()
		redirect, err := url.Parse(q.Get("redirect_uri"))
		if err != nil {
			http.Error(w, "bad redirect", http.StatusBadRequest)
			return
		}
		rq := redirect.Query()
		rq.Set("code", code)
		rq.Set("state", q.Get("state"))
		redirect.RawQuery = rq.Encode()
		http.Redirect(w, r, redirect.String(), http.StatusFound)
	case "/token":
		_ = r.ParseForm()
		code := r.PostForm.Get("code")
		f.mu.Lock()
		nonce := f.nonces[code]
		f.mu.Unlock()
		now := time.Now()
		claims := map[string]any{
			"iss":                f.server.URL,
			"sub":                "mobile-sub",
			"aud":                f.clientID,
			"exp":                now.Add(5 * time.Minute).Unix(),
			"iat":                now.Unix(),
			"nonce":              nonce,
			"email":              "oidc-mobile@example.com",
			"email_verified":     true,
			"name":               "OIDC Mobile",
			"preferred_username": "oidc_mobile",
			"groups":             []string{"community"},
		}
		header := b64url([]byte(`{"alg":"RS256","typ":"JWT","kid":"test"}`))
		payload, _ := json.Marshal(claims)
		signingInput := header + "." + b64url(payload)
		sum := sha256.Sum256([]byte(signingInput))
		sig, err := rsa.SignPKCS1v15(rand.Reader, f.key, crypto.SHA256, sum[:])
		if err != nil {
			http.Error(w, "sign", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-token",
			"token_type":   "Bearer",
			"expires_in":   300,
			"id_token":     signingInput + "." + b64url(sig),
		})
	case "/logout":
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

func TestMobileOIDCFlowEndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOIDCTestDB(t)
	fake := newFakeOIDCProvider(t)
	t.Setenv("DEBUG", "true")
	t.Setenv("OIDC_ENABLED", "true")
	t.Setenv("OIDC_ISSUER_URL", fake.server.URL)
	t.Setenv("OIDC_CLIENT_ID", fake.clientID)
	t.Setenv("OIDC_CLIENT_SECRET", "secret")
	t.Setenv("OIDC_REDIRECT_URL", "http://sandwitches.local/api/auth/oidc/callback")
	resetOIDCCacheForTest()
	t.Cleanup(resetOIDCCacheForTest)

	store := cookie.NewStore([]byte("test-secret-12345678901234567890"))
	r := gin.New()
	r.Use(sessions.Sessions("sandwitches_session", store))
	r.GET("/login/oidc", OIDCLogin)
	r.GET("/api/auth/oidc/callback", OIDCCallback)
	r.POST("/api/auth/mobile/exchange", MobileAuthExchange)
	r.GET("/me", func(c *gin.Context) {
		s := sessions.Default(c)
		c.JSON(http.StatusOK, gin.H{"user_id": s.Get("user_id"), "auth_method": s.Get("auth_method")})
	})

	// Start the browser flow and capture the OIDC cookies.
	login := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodGet, "/login/oidc?mobile=1", nil)
	r.ServeHTTP(login, loginReq)
	if login.Code != http.StatusFound {
		t.Fatalf("login status %d: %s", login.Code, login.Body.String())
	}
	authorizeURL := login.Header().Get("Location")
	if !strings.HasPrefix(authorizeURL, fake.server.URL+"/authorize") {
		t.Fatalf("authorize location %q", authorizeURL)
	}
	loginCookies := login.Result().Cookies()
	if len(loginCookies) == 0 {
		t.Fatal("login did not set OIDC cookies")
	}

	// IdP redirects back to the web callback with a code.
	noRedirect := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	idpResp, err := noRedirect.Get(authorizeURL)
	if err != nil {
		t.Fatalf("authorize request: %v", err)
	}
	defer idpResp.Body.Close()
	callbackURL := idpResp.Header.Get("Location")
	callback, err := url.Parse(callbackURL)
	if err != nil || callback.Query().Get("code") == "" {
		t.Fatalf("callback location %q: %v", callbackURL, err)
	}

	// Exchange the IdP code via the web callback. It must hand a one-time code
	// to the app instead of creating a browser session.
	cb := httptest.NewRecorder()
	cbReq := httptest.NewRequest(http.MethodGet, callback.RequestURI(), nil)
	for _, c := range loginCookies {
		cbReq.AddCookie(c)
	}
	r.ServeHTTP(cb, cbReq)
	if cb.Code != http.StatusFound {
		t.Fatalf("callback status %d: %s", cb.Code, cb.Body.String())
	}
	appRedirect := cb.Header().Get("Location")
	appURL, err := url.Parse(appRedirect)
	if err != nil || appURL.Scheme != "sandwitches" || appURL.Host != "auth" {
		t.Fatalf("app redirect %q: %v", appRedirect, err)
	}
	handoff := appURL.Query().Get("code")
	if handoff == "" {
		t.Fatalf("missing handoff code in %q", appRedirect)
	}

	// Redeem the code and verify a normal session cookie is issued.
	form := url.Values{"code": {handoff}}
	ex := httptest.NewRecorder()
	exReq := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/exchange", strings.NewReader(form.Encode()))
	exReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(ex, exReq)
	if ex.Code != http.StatusOK {
		t.Fatalf("exchange status %d: %s", ex.Code, ex.Body.String())
	}
	sessionCookies := ex.Result().Cookies()
	if len(sessionCookies) == 0 {
		t.Fatal("exchange did not set session cookie")
	}
	me := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	for _, c := range sessionCookies {
		meReq.AddCookie(c)
	}
	r.ServeHTTP(me, meReq)
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), `"auth_method":"oidc"`) {
		t.Fatalf("me status %d body %s", me.Code, me.Body.String())
	}

	// The handoff code is single use.
	replay := httptest.NewRecorder()
	replayReq := httptest.NewRequest(http.MethodPost, "/api/auth/mobile/exchange", strings.NewReader(form.Encode()))
	replayReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(replay, replayReq)
	if replay.Code != http.StatusBadRequest {
		t.Fatalf("replay status %d, want 400", replay.Code)
	}
}
