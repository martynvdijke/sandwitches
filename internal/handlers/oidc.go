package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

var (
	oidcProvider     *oidc.Provider
	oidcVerifier     *oidc.IDTokenVerifier
	oidcOAuthConfig  *oauth2.Config
	oidcInitOnce     sync.Once
	oidcInitErr      error
)

func getOIDCConfig() (*oidc.Provider, *oidc.IDTokenVerifier, *oauth2.Config, error) {
	oidcInitOnce.Do(func() {
		cfg := config.Load()
		if !cfg.OIDCEnabled {
			oidcInitErr = fmt.Errorf("oidc disabled")
			return
		}
		ctx := context.Background()
		provider, err := oidc.NewProvider(ctx, cfg.OIDCIssuer)
		if err != nil {
			oidcInitErr = err
			return
		}
		verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})
		scopes := strings.Fields(cfg.OIDCScopes)
		if len(scopes) == 0 {
			scopes = []string{oidc.ScopeOpenID, "email", "profile", "groups"}
		}
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.OIDCClientID,
			ClientSecret: cfg.OIDCClientSecret,
			RedirectURL:  cfg.OIDCRedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       scopes,
		}
		oidcProvider = provider
		oidcVerifier = verifier
		oidcOAuthConfig = oauthCfg
	})
	return oidcProvider, oidcVerifier, oidcOAuthConfig, oidcInitErr
}

// ResetOIDCForTest resets singleton (used in tests).
func ResetOIDCForTest() {
	oidcInitOnce = sync.Once{}
	oidcProvider = nil
	oidcVerifier = nil
	oidcOAuthConfig = nil
	oidcInitErr = nil
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func LoginOIDC(c *gin.Context) {
	cfg := config.Load()
	if !cfg.OIDCEnabled {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	_, _, oauthCfg, err := getOIDCConfig()
	if err != nil {
		log.Printf("oidc init error: %v", err)
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}

	state := randomString(32)
	nonce := randomString(32)
	verifier := randomString(32)
	challenge := pkceChallenge(verifier)

	// Store in session with expiry via cookie
	session := sessions.Default(c)
	session.Set("oidc_state", state)
	session.Set("oidc_nonce", nonce)
	session.Set("oidc_verifier", verifier)
	session.Set("oidc_expiry", time.Now().Add(10*time.Minute).Unix())
	// also set short-lived cookie for fallback
	c.SetCookie("oidc_state", state, 600, "/", "", !cfg.Debug, true)
	_ = session.Save()

	// preserve next param
	next := c.Query("next")
	if next != "" {
		session.Set("oidc_next", next)
		_ = session.Save()
	}

	authURL := oauthCfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	c.Redirect(http.StatusFound, authURL)
}

type oidcClaims struct {
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Groups            []string `json:"groups"`
	Nonce             string   `json:"nonce"`
}

func CallbackOIDC(c *gin.Context) {
	cfg := config.Load()
	if !cfg.OIDCEnabled {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	_, verifier, oauthCfg, err := getOIDCConfig()
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}

	session := sessions.Default(c)
	storedState, _ := session.Get("oidc_state").(string)
	storedNonce, _ := session.Get("oidc_nonce").(string)
	storedVerifier, _ := session.Get("oidc_verifier").(string)
	expiryVal := session.Get("oidc_expiry")
	var expiry int64
	switch v := expiryVal.(type) {
	case int64:
		expiry = v
	case int:
		expiry = int64(v)
	case float64:
		expiry = int64(v)
	}

	// fallback to cookie
	if storedState == "" {
		if ck, err := c.Cookie("oidc_state"); err == nil {
			storedState = ck
		}
	}

	state := c.Query("state")
	if state == "" || storedState == "" || state != storedState {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	if expiry != 0 && time.Now().Unix() > expiry {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	if c.Query("error") != "" {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}

	ctx := c.Request.Context()
	token, err := oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", storedVerifier))
	if err != nil {
		log.Printf("oidc code exchange failed: %v", err)
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		log.Printf("oidc no id_token")
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("oidc verify failed: %v", err)
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	// verify nonce
	if storedNonce != "" && idToken.Nonce != storedNonce {
		log.Printf("oidc nonce mismatch")
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Printf("oidc claims failed: %v", err)
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}
	if !claims.EmailVerified {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Email not verified by identity provider"})
		return
	}
	if claims.Email == "" {
		c.Redirect(http.StatusFound, "/login?error=oidc")
		return
	}

	// find or create user
	var user database.User
	err = database.DB.Where("email = ?", claims.Email).First(&user).Error
	if err != nil {
		// create
		username := strings.Split(claims.Email, "@")[0]
		// unique-ify
		baseUsername := username
		for i := 1; ; i++ {
			var cnt int64
			database.DB.Model(&database.User{}).Where("username = ?", username).Count(&cnt)
			if cnt == 0 {
				break
			}
			username = fmt.Sprintf("%s%d", baseUsername, i)
			if i > 100 {
				username = fmt.Sprintf("%s-%s", baseUsername, randomString(4))
				break
			}
		}
		// random password placeholder
		rndPass := randomString(32)
		hashed, _ := bcrypt.GenerateFromPassword([]byte(rndPass), bcrypt.DefaultCost)
		displayName := claims.PreferredUsername
		if displayName == "" {
			displayName = claims.Name
		}
		// split name into first/last if possible
		firstName, lastName := "", ""
		if claims.Name != "" {
			parts := strings.Fields(claims.Name)
			if len(parts) > 0 {
				firstName = parts[0]
				if len(parts) > 1 {
					lastName = strings.Join(parts[1:], " ")
				}
			}
		}
		_ = displayName
		now := time.Now()
		user = database.User{
			Username:   username,
			Email:      claims.Email,
			Password:   string(hashed),
			FirstName:  firstName,
			LastName:   lastName,
			IsActive:   true,
			DateJoined: now,
		}
		// sync groups for new user
		isAdmin := contains(claims.Groups, "admins")
		user.IsStaff = isAdmin
		user.IsSuperuser = isAdmin
		if err := database.DB.Create(&user).Error; err != nil {
			log.Printf("oidc user create failed: %v", err)
			c.Redirect(http.StatusFound, "/login?error=oidc")
			return
		}
		var communityGroup database.Group
		if err := database.DB.Where("name = ?", "community").First(&communityGroup).Error; err == nil {
			database.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, communityGroup.ID)
		}
	} else {
		// sync groups for existing user
		isAdmin := contains(claims.Groups, "admins")
		needsUpdate := false
		if isAdmin && (!user.IsStaff || !user.IsSuperuser) {
			user.IsStaff = true
			user.IsSuperuser = true
			needsUpdate = true
		} else if !isAdmin && (user.IsStaff || user.IsSuperuser) {
			// never demote last superuser
			var superCount int64
			database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&superCount)
			if superCount <= 1 && user.IsSuperuser {
				// keep superuser
			} else {
				user.IsStaff = false
				user.IsSuperuser = false
				needsUpdate = true
			}
		}
		if needsUpdate {
			database.DB.Save(&user)
		}
	}

	now := time.Now()
	user.LastLogin = &now
	database.DB.Save(&user)

	// set session
	session.Set("user_id", user.ID)
	session.Set("auth_method", "oidc")
	// clear temp oidc keys
	session.Delete("oidc_state")
	session.Delete("oidc_nonce")
	session.Delete("oidc_verifier")
	session.Delete("oidc_expiry")
	nextVal, _ := session.Get("oidc_next").(string)
	session.Delete("oidc_next")
	_ = session.Save()
	// clear cookie
	c.SetCookie("oidc_state", "", -1, "/", "", !cfg.Debug, true)

	safeNext := "/"
	if nextVal != "" && strings.HasPrefix(nextVal, "/") && !strings.HasPrefix(nextVal, "//") {
		safeNext = nextVal
	} else if qNext := c.Query("next"); qNext != "" && strings.HasPrefix(qNext, "/") {
		safeNext = qNext
	}
	// sanitize next via url parsing
	if u, err := url.Parse(safeNext); err == nil {
		if u.IsAbs() {
			safeNext = "/"
		}
	}
	c.Redirect(http.StatusFound, safeNext)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
