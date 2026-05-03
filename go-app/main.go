package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/api"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/handlers"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/tasks"
)

func main() {
	cfg := config.Load()

	database.Init(cfg)
	tasks.Init(cfg)

	djangoDB := os.Getenv("Django_DB_PATH")
	if djangoDB == "" {
		for _, p := range []string{"../src/db.sqlite3", "db.sqlite3", "/config/db.sqlite3"} {
			if _, err := os.Stat(p); err == nil {
				djangoDB = p
				break
			}
		}
	}
	if djangoDB != "" {
		log.Printf("Checking for Django database at: %s", djangoDB)
		if err := database.MigrateFromDjango(djangoDB); err != nil {
			log.Printf("Django migration skipped: %v", err)
		}
	}

	router := setupRouter(cfg)

	port := "6270"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))
	router.Use(gin.Recovery())

	store := cookie.NewStore([]byte(cfg.SecretKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   !cfg.Debug,
	})
	router.Use(sessions.Sessions("sandwitches_session", store))

	router.SetFuncMap(template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.Title,
		"contains":   strings.Contains,
		"join":       func(sep string, s []string) string { return strings.Join(s, sep) },
		"split":      func(sep, s string) []string { return strings.Split(s, sep) },
		"now":        func() time.Time { return time.Now() },
		"div":        func(a, b int) int { return a / b },
		"mul":        func(a, b int) int { return a * b },
		"mod":        func(a, b int) int { return a % b },
		"seq":        func(n int) []int { s := make([]int, n); for i := range s { s[i] = i }; return s },
		"dict":       func(values ...interface{}) map[string]interface{} { m := make(map[string]interface{}); for i := 0; i < len(values); i += 2 { m[fmt.Sprint(values[i])] = values[i+1] }; return m },
		"default":    func(def, val interface{}) interface{} { if val == nil || val == "" { return def }; return val },
		"first":      func(s string) string { if len(s) > 0 { return string(s[0]) } else { return "" } },
		"urlencode":  func(s string) string { return strings.ReplaceAll(s, " ", "+") },
		"floatmul":   func(a float64, b int) float64 { return a * float64(b) },
		"striptags":  func(s string) string { return strings.Map(func(r rune) rune { if r == '<' || r == '>' { return -1 }; return r }, s) },
		"truncatechars": func(n int, s string) string { runes := []rune(s); if len(runes) > n { return string(runes[:n]) + "..." }; return s },
		"safe":       func(s string) template.HTML { return template.HTML(s) },
		"convert_markdown": func(s string) template.HTML {
			md := s
			md = strings.ReplaceAll(md, "\n\n", "</p><p>")
			md = "<p>" + md + "</p>"
			md = strings.ReplaceAll(md, "**", "<strong>")
			md = strings.ReplaceAll(md, "__", "<em>")
			return template.HTML(md)
		},
		"to_json": func(v interface{}) string { return fmt.Sprintf("%v", v) },
		"iso8601_duration": func(minutes interface{}) string {
			m := 0
			switch v := minutes.(type) {
			case int: m = v
			case *int: if v != nil { m = *v }
			}
			return fmt.Sprintf("PT%dM", m)
		},
		"floatformat": func(precision int, val interface{}) string {
			f := 0.0
			switch v := val.(type) {
			case float64: f = v
			case *float64: if v != nil { f = *v }
			}
			return fmt.Sprintf("%."+fmt.Sprint(precision)+"f", f)
		},
	})

	templatesDir := "templates"
	router.LoadHTMLGlob(filepath.Join(templatesDir, "**/*.html"))

	router.StaticFS("/static", http.Dir("static"))
	router.StaticFS("/media", http.Dir(cfg.MediaRoot))
	router.StaticFile("/favicon.ico", "static/icons/favicon.svg")

	router.Use(middleware.OptionalAuth())

	router.GET("/", handlers.Index)

	router.GET("/setup", handlers.SetupPage)
	router.POST("/setup", handlers.Setup)

	router.GET("/signup", handlers.SignupPage)
	router.POST("/signup", handlers.Signup)

	router.GET("/login", handlers.LoginPage)
	router.POST("/login", handlers.Login)

	router.GET("/logout", handlers.Logout)

	api.RegisterRoutes(router.Group("/api"))

	router.GET("/recipes/:slug", handlers.RecipeDetail)
	router.GET("/orders/track/:token", handlers.OrderTracker)

	router.GET("/favorites", middleware.AuthRequired(), handlers.Favorites)
	router.GET("/community", middleware.AuthRequired(), handlers.Community)
	router.POST("/community", middleware.AuthRequired(), handlers.Community)

	router.GET("/profile", middleware.AuthRequired(), handlers.UserProfile)
	router.POST("/profile", middleware.AuthRequired(), handlers.UserProfile)

	router.GET("/settings", middleware.AuthRequired(), handlers.UserSettings)
	router.POST("/settings", middleware.AuthRequired(), handlers.UserSettings)

	router.GET("/orders/:id", middleware.AuthRequired(), handlers.UserOrderDetail)

	router.GET("/cart", middleware.AuthRequired(), handlers.ViewCart)
	router.GET("/cart/add/:id", middleware.AuthRequired(), handlers.AddToCart)
	router.GET("/cart/remove/:id", middleware.AuthRequired(), handlers.RemoveFromCart)
	router.POST("/cart/update/:id", middleware.AuthRequired(), handlers.UpdateCartQuantity)
	router.POST("/cart/checkout", middleware.AuthRequired(), handlers.Checkout)

	router.GET("/recipes/rate/:id", middleware.AuthRequired(), handlers.RecipeRate)
	router.POST("/recipes/rate/:id", middleware.AuthRequired(), handlers.RecipeRate)
	router.GET("/recipes/favorite/:id", middleware.AuthRequired(), handlers.ToggleFavorite)

	admin := router.Group("/dashboard", middleware.StaffRequired())
	{
		admin.GET("", handlers.AdminDashboard)

		admin.GET("/recipes", handlers.AdminRecipeList)
		admin.GET("/recipes/add", handlers.AdminRecipeAdd)
		admin.POST("/recipes/add", handlers.AdminRecipeAdd)
		admin.GET("/recipes/:id/edit", handlers.AdminRecipeEdit)
		admin.POST("/recipes/:id/edit", handlers.AdminRecipeEdit)
		admin.GET("/recipes/:id/delete", handlers.AdminRecipeDelete)
		admin.POST("/recipes/:id/delete", handlers.AdminRecipeDelete)
		admin.GET("/recipes/:id/approve", handlers.AdminRecipeApprove)

		admin.GET("/approvals", handlers.AdminRecipeApprovalList)

		admin.GET("/users", handlers.AdminUserList)
		admin.GET("/users/:id/edit", handlers.AdminUserEdit)
		admin.POST("/users/:id/edit", handlers.AdminUserEdit)
		admin.GET("/users/:id/delete", handlers.AdminUserDelete)
		admin.POST("/users/:id/delete", handlers.AdminUserDelete)

		admin.GET("/tags", handlers.AdminTagList)
		admin.GET("/tags/add", handlers.AdminTagAdd)
		admin.POST("/tags/add", handlers.AdminTagAdd)
		admin.GET("/tags/:id/edit", handlers.AdminTagEdit)
		admin.POST("/tags/:id/edit", handlers.AdminTagEdit)
		admin.GET("/tags/:id/delete", handlers.AdminTagDelete)
		admin.POST("/tags/:id/delete", handlers.AdminTagDelete)

		admin.GET("/orders", handlers.AdminOrderList)
		admin.POST("/orders/:id/status", handlers.AdminOrderUpdateStatus)

		admin.GET("/ratings", handlers.AdminRatingList)
		admin.GET("/ratings/:id/delete", handlers.AdminRatingDelete)
		admin.POST("/ratings/:id/delete", handlers.AdminRatingDelete)

		admin.GET("/settings", handlers.AdminSettings)
		admin.POST("/settings", handlers.AdminSettings)
	}

	_ = io.Discard

	return router
}
