package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

// OIDCConfig holds OIDC RP config from env.
type OIDCConfig struct {
	Enabled     bool
	Issuer      string
	ClientID    string
	Secret      string
	RedirectURL string
	Scopes      []string
	LogoutURL   string
}

func loadOIDCConfig() OIDCConfig {
	enabled := true
	if v, ok := os.LookupEnv("OIDC_ENABLED"); ok {
		enabled = v == "true" || v == "1"
	}
	cfg := OIDCConfig{Enabled: enabled}
	if f := os.Getenv("OIDC_CLIENT_SECRET_FILE"); f != "" {
		if b, err := os.ReadFile(f); err == nil {
			s := strings.TrimSpace(string(b))
			if s != "" {
				cfg.Secret = s
			}
		}
	}
	if cfg.Secret == "" {
		cfg.Secret = strings.TrimSpace(os.Getenv("OIDC_CLIENT_SECRET"))
	}
	issuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER_URL"))
	if issuer == "" {
		issuer = strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	}
	cfg.Issuer = strings.TrimSuffix(issuer, "/")
	cfg.ClientID = strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))
	cfg.RedirectURL = strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URL"))
	if s := os.Getenv("OIDC_SCOPES"); s != "" {
		cfg.Scopes = strings.Fields(s)
	} else {
		cfg.Scopes = []string{"openid", "email", "profile", "groups"}
	}
	if u := strings.TrimSpace(os.Getenv("OIDC_LOGOUT_URL")); u != "" {
		cfg.LogoutURL = u
	} else {
		cfg.LogoutURL = cfg.Issuer + "/logout"
	}
	return cfg
}

func (c OIDCConfig) valid() bool {
	return c.Enabled && c.Issuer != "" && c.ClientID != "" && c.Secret != "" && c.RedirectURL != ""
}

type oidcProvider struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
	issuer   string
	clientID string
}

var (
	oidcMu    sync.Mutex
	oidcCache *oidcProvider
)

func getOIDCProvider(ctx context.Context, cfg OIDCConfig) (*oidcProvider, error) {
	oidcMu.Lock()
	defer oidcMu.Unlock()
	if oidcCache != nil && oidcCache.issuer == cfg.Issuer && oidcCache.clientID == cfg.ClientID {
		return oidcCache, nil
	}
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oidcCache = &oidcProvider{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.Secret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
		},
		issuer:   cfg.Issuer,
		clientID: cfg.ClientID,
	}
	return oidcCache, nil
}

// resetOIDCCacheForTest clears cache (test helper).
func resetOIDCCacheForTest() {
	oidcMu.Lock()
	oidcCache = nil
	oidcMu.Unlock()
}

func oidcRandHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func setOIDCCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := os.Getenv("DEBUG") != "true"
	c.SetCookie(name, value, maxAge, "/", "", secure, true)
}

func OIDCStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": loadOIDCConfig().valid()})
}

func OIDCLogin(c *gin.Context) {
	cfg := loadOIDCConfig()
	if !cfg.valid() {
		c.JSON(http.StatusNotFound, gin.H{"error": "oidc disabled"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	p, err := getOIDCProvider(ctx, cfg)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "oidc discovery failed"})
		return
	}
	state, err := oidcRandHex(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}
	nonce, err := oidcRandHex(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}
	verifier := oauth2.GenerateVerifier()
	setOIDCCookie(c, "oidc_state", state, 300)
	setOIDCCookie(c, "oidc_nonce", nonce, 300)
	setOIDCCookie(c, "oidc_verifier", verifier, 300)
	url := p.oauth.AuthCodeURL(state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
	)
	c.Redirect(http.StatusFound, url)
}

type oidcClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	PreferredName string `json:"preferred_username"`
	Nonce         string `json:"nonce"`
}

func clearOIDCCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := os.Getenv("DEBUG") != "true"
	for _, n := range []string{"oidc_state", "oidc_nonce", "oidc_verifier"} {
		c.SetCookie(n, "", -1, "/", "", secure, true)
	}
}

func oidcFail(c *gin.Context, msg string) {
	clearOIDCCookies(c)
	c.Redirect(http.StatusFound, "/login?error="+msg)
	c.Abort()
}

func OIDCCallback(c *gin.Context) {
	cfg := loadOIDCConfig()
	if !cfg.valid() {
		c.JSON(http.StatusNotFound, gin.H{"error": "oidc disabled"})
		return
	}
	state, err1 := c.Cookie("oidc_state")
	nonce, err2 := c.Cookie("oidc_nonce")
	verifier, err3 := c.Cookie("oidc_verifier")
	if err1 != nil || err2 != nil || err3 != nil || state == "" || nonce == "" || verifier == "" {
		oidcFail(c, "oidc_expired")
		return
	}
	if q := c.Query("state"); q == "" || subtle.ConstantTimeCompare([]byte(q), []byte(state)) != 1 {
		oidcFail(c, "oidc_state")
		return
	}
	if c.Query("code") == "" {
		oidcFail(c, "oidc_code")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	p, err := getOIDCProvider(ctx, cfg)
	if err != nil {
		oidcFail(c, "oidc_provider")
		return
	}
	token, err := p.oauth.Exchange(ctx, c.Query("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		oidcFail(c, "oidc_exchange")
		return
	}
	raw, ok := token.Extra("id_token").(string)
	if !ok || raw == "" {
		oidcFail(c, "oidc_token")
		return
	}
	idToken, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		oidcFail(c, "oidc_verify")
		return
	}
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		oidcFail(c, "oidc_claims")
		return
	}
	if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(nonce)) != 1 {
		oidcFail(c, "oidc_nonce")
		return
	}
	if claims.Email == "" || !claims.EmailVerified {
		oidcFail(c, "oidc_email")
		return
	}
	var rawClaims map[string]any
	groups := []string{}
	groupsPresent := false
	if err := idToken.Claims(&rawClaims); err == nil {
		if g, ok := rawClaims["groups"]; ok && g != nil {
			groupsPresent = true
			if arr, ok := g.([]any); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						groups = append(groups, s)
					}
				}
			}
		}
	}
	sub := cfg.Issuer + "|" + claims.Sub
	u, err := linkOrProvisionOIDCUser(sub, claims, groups, groupsPresent)
	if err != nil {
		oidcFail(c, "oidc_user")
		return
	}
	clearOIDCCookies(c)
	// Regenerate session to prevent fixation (clear pre-existing session)
	session := sessions.Default(c)
	session.Clear()
	session.Set("user_id", u.ID)
	session.Set("auth_method", "oidc")
	_ = session.Save()
	c.Redirect(http.StatusFound, "/")
}

func applyGroupsToUser(user *database.User, groups []string, groupsPresent bool) bool {
	if !groupsPresent {
		return false
	}
	hasAdmins := false
	for _, g := range groups {
		if g == "admins" {
			hasAdmins = true
			break
		}
	}
	changed := false
	if hasAdmins {
		if !user.IsStaff || !user.IsSuperuser {
			user.IsStaff = true
			user.IsSuperuser = true
			changed = true
		}
	} else {
		if user.IsStaff || user.IsSuperuser {
			user.IsStaff = false
			user.IsSuperuser = false
			changed = true
		}
	}
	return changed
}

func linkOrProvisionOIDCUser(sub string, claims oidcClaims, groups []string, groupsPresent bool) (*database.User, error) {
	var user database.User
	if sub != "" && sub != "|" {
		if err := database.DB.Where("oidc_sub = ?", sub).First(&user).Error; err == nil {
			if applyGroupsToUser(&user, groups, groupsPresent) {
				_ = database.DB.Save(&user).Error
			}
			return &user, nil
		}
	}
	emailLower := strings.ToLower(strings.TrimSpace(claims.Email))
	if emailLower != "" {
		if err := database.DB.Where("LOWER(email) = ?", emailLower).First(&user).Error; err == nil {
			if user.OIDCSub == "" && sub != "" && sub != "|" {
				user.OIDCSub = sub
			}
			applyGroupsToUser(&user, groups, groupsPresent)
			_ = database.DB.Save(&user).Error
			return &user, nil
		}
	}
	username := claims.PreferredName
	if username == "" {
		username = claims.Name
	}
	if username == "" {
		if idx := strings.Index(claims.Email, "@"); idx > 0 {
			username = claims.Email[:idx]
		} else {
			username = claims.Email
		}
	}
	username = uniquifyUsername(username)
	pw, _ := oidcRandHex(16)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("oidc$"+pw), bcrypt.DefaultCost)
	user = database.User{
		Username: username,
		Password: string(hashed),
		Email:    emailLower,
		OIDCSub:  sub,
		IsActive: true,
	}
	if claims.Name != "" {
		user.FirstName = claims.Name
	}
	applyGroupsToUser(&user, groups, groupsPresent)
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	var cg database.Group
	if err := database.DB.Where("name = ?", "community").First(&cg).Error; err == nil {
		_ = database.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, cg.ID).Error
	}
	return &user, nil
}

func uniquifyUsername(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "user"
	}
	base = strings.ReplaceAll(base, " ", "_")
	candidate := base
	for i := 2; ; i++ {
		var count int64
		database.DB.Model(&database.User{}).Where("username = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = base + "-" + strconv.Itoa(i)
		if i > 100 {
			return candidate
		}
	}
}
