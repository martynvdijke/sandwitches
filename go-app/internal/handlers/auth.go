package handlers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
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
