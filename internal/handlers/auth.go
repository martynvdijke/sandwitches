package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/tasks"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

// verifyPassword checks a stored password hash against the supplied password.
// Supports bcrypt (native, used by new signups) and Django PBKDF2 hashes
// (pbkdf2_sha256$..., carried over from the old Django backend) so that
// migrated users can keep logging in with their original passwords.
func verifyPassword(stored, password string) bool {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
	}
	return verifyDjangoPBKDF2(stored, password)
}

// verifyDjangoPBKDF2 verifies a Django PBKDF2-SHA256 hash of the form
// pbkdf2_sha256$iterations$salt$b64hash.
func verifyDjangoPBKDF2(stored, password string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}
	salt := parts[2]
	expected, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	dk := pbkdf2.Key([]byte(password), []byte(salt), iterations, 32, sha256.New)
	return subtle.ConstantTimeCompare(dk, expected) == 1
}

type SignupForm struct {
	Username  string `form:"username" binding:"required,min=3,max=150"`
	Password1 string `form:"password1" binding:"required,min=8"`
	Password2 string `form:"password2" binding:"required,eqfield=Password1"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Email     string `form:"email" binding:"required,email"`
	Bio       string `form:"bio"`
}

func SignupPage(c *gin.Context) {
	td := utils.NewTemplateData(c)
	c.HTML(http.StatusOK, "signup.html", td.ToGinH())
}

func Signup(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "signup.html", td.With("error", err.Error()).With("form", form).ToGinH())
		return
	}

	var count int64
	database.DB.Model(&database.User{}).Where("username = ?", form.Username).Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "signup.html", td.With("error", "Username already taken").With("form", form).ToGinH())
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password1), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "signup.html", td.With("error", "Server error").With("form", form).ToGinH())
		return
	}

	user := database.User{
		Username:   form.Username,
		Password:   string(hashed),
		FirstName:  form.FirstName,
		LastName:   form.LastName,
		Email:      form.Email,
		Bio:        form.Bio,
		Language:   "en",
		Theme:      "light",
		IsActive:   true,
		DateJoined: time.Now(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusOK, "signup.html", td.With("error", "Could not create account").With("form", form).ToGinH())
		return
	}

	var communityGroup database.Group
	if err := database.DB.Where("name = ?", "community").First(&communityGroup).Error; err == nil {
		database.DB.Model(&user).Association("Favorites").Append(nil)
		database.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, communityGroup.ID)
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	if err := session.Save(); err != nil {
		c.HTML(http.StatusOK, "signup.html", td.With("error", "Session error").ToGinH())
		return
	}

	utils.AddFlash(c, "success", "Welcome to Sandwitches!")
	c.Redirect(http.StatusFound, "/")
}

type LoginForm struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func LoginPage(c *gin.Context) {
	next := c.Query("next")
	c.HTML(http.StatusOK, "login.html", gin.H{"next": next})
}

func Login(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "login.html", td.With("error", "Invalid form").With("form", form).ToGinH())
		return
	}

	var user database.User
	if err := database.DB.Where("username = ?", form.Username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", td.With("error", "Invalid username or password").With("form", form).ToGinH())
		return
	}

	if !verifyPassword(user.Password, form.Password) {
		c.HTML(http.StatusOK, "login.html", td.With("error", "Invalid username or password").With("form", form).ToGinH())
		return
	}

	now := time.Now()
	user.LastLogin = &now
	database.DB.Save(&user)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	if err := session.Save(); err != nil {
		c.HTML(http.StatusOK, "login.html", td.With("error", "Session error").ToGinH())
		return
	}

	next := c.PostForm("next")
	if next == "" {
		next = "/"
	}
	utils.AddFlash(c, "success", "Welcome back, "+user.Username+"!")
	c.Redirect(http.StatusFound, next)
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	utils.AddFlash(c, "success", "You have been logged out")
	c.Redirect(http.StatusFound, "/")
}

const resetTokenTTL = 24 * time.Hour

type ForgotPasswordForm struct {
	Email string `form:"email" binding:"required,email"`
}

// generateResetToken returns a cryptographically random, URL-safe token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashResetToken hashes a raw reset token so the plaintext is never stored.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// validResetToken looks up an unused, unexpired reset token by its raw value.
func validResetToken(token string) (*database.PasswordResetToken, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}
	var rt database.PasswordResetToken
	if err := database.DB.Where("token_hash = ?", hashResetToken(token)).First(&rt).Error; err != nil {
		return nil, err
	}
	if rt.UsedAt != nil {
		return nil, fmt.Errorf("token already used")
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}
	return &rt, nil
}

func ForgotPasswordPage(c *gin.Context) {
	td := utils.NewTemplateData(c)
	c.HTML(http.StatusOK, "forgot_password.html", td.ToGinH())
}

func ForgotPassword(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var form ForgotPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "forgot_password.html", td.With("error", "Enter a valid email address").ToGinH())
		return
	}

	// Always report success regardless of whether the account exists so the
	// endpoint cannot be used to enumerate registered emails.
	td.With("sent", true)

	var user database.User
	if err := database.DB.Where("email = ?", form.Email).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "forgot_password.html", td.ToGinH())
		return
	}

	// Invalidate any previously issued tokens for this user (single active token).
	database.DB.Model(&database.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", user.ID).
		Update("used_at", time.Now())

	token, err := generateResetToken()
	if err != nil {
		log.Printf("password reset token generation failed: %v", err)
		c.HTML(http.StatusOK, "forgot_password.html", td.ToGinH())
		return
	}

	rt := database.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashResetToken(token),
		ExpiresAt: time.Now().Add(resetTokenTTL),
	}
	if err := database.DB.Create(&rt).Error; err != nil {
		log.Printf("password reset token save failed: %v", err)
		c.HTML(http.StatusOK, "forgot_password.html", td.ToGinH())
		return
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL(c), token)
	subject := "Reset your Sandwitches password"
	textBody := fmt.Sprintf("You requested a password reset. Open this link to choose a new password:\n\n%s\n\nIf you didn't request this, you can ignore this email.", resetURL)
	htmlBody := fmt.Sprintf("<p>You requested a password reset. Open this link to choose a new password:</p><p><a href=\"%s\">%s</a></p>", resetURL, resetURL)
	if tasks.EmailEnabled() {
		tasks.SendHTMLEmail(user.Email, subject, textBody, htmlBody)
	} else {
		log.Printf("[password-reset] reset link for %s: %s", user.Email, resetURL)
	}

	c.HTML(http.StatusOK, "forgot_password.html", td.ToGinH())
}

type ResetPasswordForm struct {
	Token     string `form:"token" binding:"required"`
	Password1 string `form:"password1" binding:"required,min=8"`
	Password2 string `form:"password2" binding:"required,eqfield=Password1"`
}

func ResetPasswordPage(c *gin.Context) {
	td := utils.NewTemplateData(c)
	token := c.Query("token")
	if _, err := validResetToken(token); err != nil {
		c.HTML(http.StatusOK, "reset_password.html", td.With("error", "This reset link is invalid or has expired.").ToGinH())
		return
	}
	c.HTML(http.StatusOK, "reset_password.html", td.With("token", token).ToGinH())
}

func ResetPassword(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var form ResetPasswordForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "reset_password.html", td.With("error", "Passwords must match and be at least 8 characters.").With("token", form.Token).ToGinH())
		return
	}

	rt, err := validResetToken(form.Token)
	if err != nil {
		c.HTML(http.StatusOK, "reset_password.html", td.With("error", "This reset link is invalid or has expired.").With("token", form.Token).ToGinH())
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password1), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "reset_password.html", td.With("error", "Server error").With("token", form.Token).ToGinH())
		return
	}

	now := time.Now()
	if err := database.DB.Model(&database.User{}).Where("id = ?", rt.UserID).Update("password", string(hashed)).Error; err != nil {
		c.HTML(http.StatusOK, "reset_password.html", td.With("error", "Server error").With("token", form.Token).ToGinH())
		return
	}
	database.DB.Model(&database.PasswordResetToken{}).Where("id = ?", rt.ID).Update("used_at", now)

	utils.AddFlash(c, "success", "Password reset. You can log in now.")
	c.Redirect(http.StatusFound, "/login")
}

func SetupPage(c *gin.Context) {
	var count int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}
	td := utils.NewTemplateData(c)
	c.HTML(http.StatusOK, "setup.html", td.ToGinH())
}

func Setup(c *gin.Context) {
	var count int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}
	td := utils.NewTemplateData(c)

	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "setup.html", td.With("error", err.Error()).ToGinH())
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(form.Password1), bcrypt.DefaultCost)
	user := database.User{
		Username:    form.Username,
		Password:    string(hashed),
		Email:       form.Email,
		IsSuperuser: true,
		IsStaff:     true,
		IsActive:    true,
		Language:    "en",
		Theme:       "light",
		DateJoined:  time.Now(),
	}
	database.DB.Create(&user)

	var adminGroup database.Group
	database.DB.Where("name = ?", "admin").First(&adminGroup)
	database.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, adminGroup.ID)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	_ = session.Save()

	utils.AddFlash(c, "success", "Admin account created")
	c.Redirect(http.StatusFound, "/dashboard/")
}

func AuthUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user": middleware.GetUser(c)})
}
