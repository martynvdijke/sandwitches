package handlers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

type SignupForm struct {
	Username  string `form:"username" binding:"required,min=3,max=150"`
	Password1 string `form:"password1" binding:"required,min=8"`
	Password2 string `form:"password2" binding:"required,eqfield=Password1"`
	FirstName string `form:"first_name"`
	LastName  string `form:"last_name"`
	Email     string `form:"email"`
	Bio       string `form:"bio"`
}

func SignupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "signup.html", gin.H{})
}

func Signup(c *gin.Context) {
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "signup.html", gin.H{"error": err.Error(), "form": form})
		return
	}

	var count int64
	database.DB.Model(&database.User{}).Where("username = ?", form.Username).Count(&count)
	if count > 0 {
		c.HTML(http.StatusOK, "signup.html", gin.H{"error": "Username already taken", "form": form})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password1), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "signup.html", gin.H{"error": "Server error", "form": form})
		return
	}

	user := database.User{
		Username:  form.Username,
		Password:  string(hashed),
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     form.Email,
		Bio:       form.Bio,
		Language:  "en",
		Theme:     "light",
		IsActive:  true,
		DateJoined: time.Now(),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusOK, "signup.html", gin.H{"error": "Could not create account", "form": form})
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
		c.HTML(http.StatusOK, "signup.html", gin.H{"error": "Session error"})
		return
	}

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
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Invalid form", "form": form})
		return
	}

	var user database.User
	if err := database.DB.Where("username = ?", form.Username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Invalid username or password", "form": form})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Invalid username or password", "form": form})
		return
	}

	now := time.Now()
	user.LastLogin = &now
	database.DB.Save(&user)

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	if err := session.Save(); err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"error": "Session error"})
		return
	}

	next := c.PostForm("next")
	if next == "" {
		next = "/"
	}
	c.Redirect(http.StatusFound, next)
}

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	_ = session.Save()
	c.Redirect(http.StatusFound, "/")
}

func SetupPage(c *gin.Context) {
	var count int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "setup.html", gin.H{})
}

func Setup(c *gin.Context) {
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusOK, "setup.html", gin.H{"error": err.Error()})
		return
	}

	var count int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		c.Redirect(http.StatusFound, "/")
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

	c.Redirect(http.StatusFound, "/dashboard/")
}

func AuthUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user": middleware.GetUser(c)})
}
