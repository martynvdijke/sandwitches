package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	code := m.Run()
	os.Remove("test_db.sqlite3")
	os.Exit(code)
}

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	tmp := t.TempDir()
	dbPath := tmp + "/test.db"

	database.Init(&config.Config{
		DatabaseFile: dbPath,
		Debug:        false,
		SecretKey:    "test-secret-key",
		LanguageCode: "en",
	})

	SetMediaRoot(tmp + "/media")
	os.MkdirAll(tmp+"/media", 0755)

	r := gin.New()
	store := cookie.NewStore([]byte("test-secret-key"))
	r.Use(sessions.Sessions("sandwitches_session", store))
	r.Use(middleware.OptionalAuth())

	r.SetFuncMap(template.FuncMap{
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
		"safe":         func(s string) template.HTML { return template.HTML(s) },
		"convert_markdown": func(s string) template.HTML { return template.HTML(s) },
		"to_json":           func(v interface{}) string { return fmt.Sprintf("%v", v) },
		"iso8601_duration": func(minutes interface{}) string {
			m := 0
			switch v := minutes.(type) {
			case int:
				m = v
			case *int:
				if v != nil {
					m = *v
				}
			}
			return fmt.Sprintf("PT%dM", m)
		},
		"floatformat": func(precision int, val interface{}) string {
			f := 0.0
			switch v := val.(type) {
			case float64:
				f = v
			case *float64:
				if v != nil {
					f = *v
				}
			}
			return fmt.Sprintf("%."+fmt.Sprint(precision)+"f", f)
		},
		"version": func() string { return "test" },
	})

	loadTestTemplates(r)

	r.GET("/", Index)
	r.GET("/feeds/latest", LatestRecipesFeed)
	r.GET("/recipes/:slug", RecipeDetail)
	r.GET("/orders/track/:token", OrderTracker)
	r.GET("/signup", SignupPage)
	r.POST("/signup", Signup)
	r.GET("/login", LoginPage)
	r.POST("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/setup", SetupPage)
	r.POST("/setup", Setup)

	csrf := r.Group("/")
	csrf.Use(middleware.CSRFMiddleware())
	{
		csrf.GET("/favorites", middleware.AuthRequired(), Favorites)
		csrf.GET("/profile", middleware.AuthRequired(), UserProfile)
		csrf.POST("/profile", middleware.AuthRequired(), UserProfile)
		csrf.GET("/cart", middleware.AuthRequired(), ViewCart)
		csrf.GET("/cart/add/:id", middleware.AuthRequired(), AddToCart)
		csrf.GET("/cart/remove/:id", middleware.AuthRequired(), RemoveFromCart)
		csrf.POST("/cart/update/:id", middleware.AuthRequired(), UpdateCartQuantity)
		csrf.POST("/cart/checkout", middleware.AuthRequired(), Checkout)
		csrf.GET("/recipes/favorite/:id", middleware.AuthRequired(), ToggleFavorite)
		csrf.GET("/recipes/rate/:id", middleware.AuthRequired(), RecipeRate)
		csrf.POST("/recipes/rate/:id", middleware.AuthRequired(), RecipeRate)
		csrf.GET("/community", middleware.AuthRequired(), Community)
		csrf.POST("/community", middleware.AuthRequired(), Community)

		admin := csrf.Group("/dashboard", middleware.StaffRequired())
		{
			admin.GET("", AdminDashboard)
			admin.GET("/recipes", AdminRecipeList)
			admin.GET("/recipes/add", AdminRecipeAdd)
			admin.POST("/recipes/add", AdminRecipeAdd)
			admin.GET("/recipes/:id/edit", AdminRecipeEdit)
			admin.POST("/recipes/:id/edit", AdminRecipeEdit)
			admin.GET("/recipes/:id/delete", AdminRecipeDelete)
			admin.POST("/recipes/:id/delete", AdminRecipeDelete)
			admin.GET("/recipes/:id/approve", AdminRecipeApprove)
			admin.GET("/recipes/:id/rotate", AdminRecipeRotate)
			admin.GET("/approvals", AdminRecipeApprovalList)
			admin.GET("/users", AdminUserList)
			admin.GET("/users/:id/edit", AdminUserEdit)
			admin.POST("/users/:id/edit", AdminUserEdit)
			admin.GET("/users/:id/delete", AdminUserDelete)
			admin.POST("/users/:id/delete", AdminUserDelete)
			admin.GET("/tags", AdminTagList)
			admin.GET("/tags/add", AdminTagAdd)
			admin.POST("/tags/add", AdminTagAdd)
			admin.GET("/tags/:id/edit", AdminTagEdit)
			admin.POST("/tags/:id/edit", AdminTagEdit)
			admin.GET("/tags/:id/delete", AdminTagDelete)
			admin.POST("/tags/:id/delete", AdminTagDelete)
			admin.GET("/orders", AdminOrderList)
			admin.POST("/orders/:id/status", AdminOrderUpdateStatus)
			admin.GET("/ratings", AdminRatingList)
			admin.GET("/ratings/:id/delete", AdminRatingDelete)
			admin.POST("/ratings/:id/delete", AdminRatingDelete)
			admin.GET("/settings", AdminSettings)
			admin.POST("/settings", AdminSettings)
			admin.GET("/logs", AdminLogs)
			admin.POST("/logs", AdminLogs)
			admin.GET("/tasks", AdminTasks)
		}
	}

	srv := httptest.NewServer(r)

	cleanup := func() {
		srv.Close()
		sqlDB, _ := database.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return srv, cleanup
}

func loadTestTemplates(r *gin.Engine) {
	tmpl := template.New("").Funcs(r.FuncMap)
	templatesDir := "../../templates"
	filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		relPath, _ := filepath.Rel(templatesDir, path)
		relPath = filepath.ToSlash(relPath)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(b)
		if strings.Contains(content, "{{ define ") || strings.Contains(content, "{{ block ") {
			tmpl = template.Must(tmpl.Parse(content))
		} else {
			tmpl = template.Must(tmpl.Parse(`{{ define "` + relPath + `" }}` + content + `{{ end }}`))
		}
		return nil
	})
	r.SetHTMLTemplate(tmpl)
}

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func createTestUser(t *testing.T, username, password string, isStaff, isSuperuser bool, groupName string) *database.User {
	t.Helper()
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := database.User{
		Username:    username,
		Password:    string(hashed),
		Email:       username + "@test.com",
		Language:    "en",
		Theme:       "light",
		IsActive:    true,
		IsStaff:     isStaff,
		IsSuperuser: isSuperuser,
		DateJoined:  time.Now(),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if groupName != "" {
		var group database.Group
		database.DB.Where("name = ?", groupName).FirstOrCreate(&group, database.Group{Name: groupName})
		database.DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, group.ID)
	}

	return &user
}

func loginUser(t *testing.T, srv *httptest.Server, client *http.Client, username, password string) {
	t.Helper()
	body := url.Values{}
	body.Set("username", username)
	body.Set("password", password)
	resp, err := client.PostForm(srv.URL+"/login", body)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("login returned %d, expected 302", resp.StatusCode)
	}
}

func getCSRFToken(t *testing.T, srv *httptest.Server, client *http.Client) string {
	t.Helper()
	resp, err := client.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatalf("failed to get CSRF token: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)

	nameIdx := strings.Index(s, `name="csrf_token"`)
	if nameIdx == -1 {
		nameIdx = strings.Index(s, "name='csrf_token'")
	}
	if nameIdx == -1 {
		t.Fatalf("CSRF token marker not found. Status: %d, first 500: %s",
			resp.StatusCode, s[:min(500, len(s))])
		return ""
	}

	valueIdx := strings.Index(s[nameIdx:], `value="`)
	if valueIdx == -1 {
		valueIdx = strings.Index(s[nameIdx:], "value='")
	}
	if valueIdx == -1 {
		t.Fatalf("CSRF value not found after name attr")
		return ""
	}

	start := nameIdx + valueIdx + len(`value="`)
	s = s[start:]
	end := strings.IndexByte(s, '"')
	if end == -1 {
		end = strings.IndexByte(s, '\'')
	}
	if end == -1 {
		end = strings.IndexByte(s, ' ')
	}
	if end == -1 {
		t.Fatalf("CSRF token value terminator not found")
		return ""
	}
	return s[:end]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func createTestRecipe(t *testing.T, title string, uploadedBy *database.User, approved bool) *database.Recipe {
	t.Helper()
	recipe := database.Recipe{
		Title:        title,
		Description:  "A test recipe",
		Ingredients:  "2 cups flour\n1 tsp salt",
		Instructions: "Mix and bake",
		Servings:     2,
		IsApproved:   approved,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if uploadedBy != nil {
		recipe.UploadedByID = &uploadedBy.ID
		recipe.UploadedBy = uploadedBy
	}
	recipe.Slug = database.UniqueSlug(title, nil)
	if err := database.DB.Create(&recipe).Error; err != nil {
		t.Fatalf("failed to create recipe: %v", err)
	}
	return &recipe
}

// ========== TESTS ==========

func TestIndexPage(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	adminUser := createTestUser(t, "admin_chef", "password123", true, true, "admin")
	createTestRecipe(t, "AdminRecipe", adminUser, true)

	communityUser := createTestUser(t, "community_i", "password123", false, false, "community")
	createTestRecipe(t, "CommunityRecipe", communityUser, true)

	resp, _ := newClient().Get(srv.URL + "/")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestIndexPageRedirectsToSetup(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	var count int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&count)
	if count > 0 {
		t.Skip("superuser already exists")
	}

	resp, _ := newClient().Get(srv.URL + "/")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect to setup, got %d", resp.StatusCode)
	}
}

func TestIndexPageFiltering(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	adminUser := createTestUser(t, "admin1", "password123", true, true, "admin")
	createTestRecipe(t, "PastaCarbonara", adminUser, true)
	createTestRecipe(t, "SushiRolls", adminUser, true)

	resp, _ := newClient().Get(srv.URL + "/?q=Pasta")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestIndexSorting(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	adminUser := createTestUser(t, "admin_sort", "password123", true, true, "admin")
	r1 := createTestRecipe(t, "RecipeA", adminUser, true)
	r2 := createTestRecipe(t, "RecipeB", adminUser, true)
	r1.CreatedAt = time.Now().Add(-1 * time.Hour)
	database.DB.Save(r1)
	r2.CreatedAt = time.Now()
	database.DB.Save(r2)

	resp, _ := newClient().Get(srv.URL + "/?sort=date_asc")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRecipeDetail(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	adminUser := createTestUser(t, "detail_user", "password123", true, true, "admin")
	recipe := createTestRecipe(t, "SpecialBrownie", adminUser, true)

	resp, _ := newClient().Get(srv.URL + "/recipes/" + recipe.Slug)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCommunityRecipeDetailHidden(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	communityUser := createTestUser(t, "community_det", "password123", false, false, "community")
	recipe := createTestRecipe(t, "HiddenRecipe", communityUser, false)

	resp, _ := newClient().Get(srv.URL + "/recipes/" + recipe.Slug)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404 for unapproved community recipe, got %d", resp.StatusCode)
	}
}

func TestSetupPage(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	resp, _ := newClient().Get(srv.URL + "/setup")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestSetupCreateSuperuser(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	client := newClient()
	body := url.Values{}
	body.Set("username", "new_admin")
	body.Set("password1", "adminpass123")
	body.Set("password2", "adminpass123")
	body.Set("email", "admin@test.com")

	resp, err := client.PostForm(srv.URL+"/setup", body)
	if err != nil {
		t.Fatalf("setup request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}

	var user database.User
	if err := database.DB.Where("username = ?", "new_admin").First(&user).Error; err != nil {
		t.Error("superuser was not created")
	}
	if !user.IsSuperuser || !user.IsStaff {
		t.Error("created user should be superuser and staff")
	}
}

func TestSetupRedirectWhenSuperuserExists(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "existing_admin", "password123", true, true, "admin")

	resp, _ := newClient().Get(srv.URL + "/setup")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}
}

func TestSignup(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	client := newClient()
	body := url.Values{}
	body.Set("username", "new_user")
	body.Set("password1", "userpass123")
	body.Set("password2", "userpass123")
	body.Set("email", "new@test.com")
	body.Set("first_name", "New")
	body.Set("last_name", "User")

	resp, err := client.PostForm(srv.URL+"/signup", body)
	if err != nil {
		t.Fatalf("signup request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}

	var user database.User
	if err := database.DB.Where("username = ?", "new_user").First(&user).Error; err != nil {
		t.Error("user was not created")
	}
	if user.FirstName != "New" || user.LastName != "User" {
		t.Error("user profile fields not saved")
	}

	var group database.Group
	database.DB.Where("name = ?", "community").First(&group)
	var count int64
	database.DB.Table("user_groups").Where("user_id = ? AND group_id = ?", user.ID, group.ID).Count(&count)
	if count == 0 {
		t.Error("user should be in community group")
	}
}

func TestSignupDuplicateUsername(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "existing", "password123", false, false, "")

	client := newClient()
	body := url.Values{}
	body.Set("username", "existing")
	body.Set("password1", "password123")
	body.Set("password2", "password123")

	resp, err := client.PostForm(srv.URL+"/signup", body)
	if err != nil {
		t.Fatalf("signup request failed: %v", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(string(respBody), "Username already taken") {
		t.Error("should show duplicate username error")
	}
}

func TestLogin(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "login_user", "password123", false, false, "")

	client := newClient()
	body := url.Values{}
	body.Set("username", "login_user")
	body.Set("password", "password123")

	resp, err := client.PostForm(srv.URL+"/login", body)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}

	hasCookie := false
	for _, c := range client.Jar.Cookies(resp.Request.URL) {
		if c.Name == "sandwitches_session" {
			hasCookie = true
			break
		}
	}
	if !hasCookie {
		t.Error("login should set session cookie")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "bad_login", "password123", false, false, "")

	client := newClient()
	body := url.Values{}
	body.Set("username", "bad_login")
	body.Set("password", "wrong_password")

	resp, err := client.PostForm(srv.URL+"/login", body)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if !strings.Contains(string(respBody), "Invalid username or password") {
		t.Error("should show invalid credentials error")
	}
}

func TestLogout(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "logout_user", "password123", false, false, "")

	client := newClient()
	loginUser(t, srv, client, "logout_user", "password123")

	resp, err := client.Get(srv.URL + "/logout")
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}
}

func TestFavoritesRequiresAuth(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	resp, _ := newClient().Get(srv.URL + "/favorites")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}
}

func TestFavoritesToggle(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "fav_toggle", "password123", false, false, "")
	adminUser := createTestUser(t, "fav_toggle_a", "password123", true, true, "admin")
	recipe := createTestRecipe(t, "FavoriteTog", adminUser, true)

	client := newClient()
	loginUser(t, srv, client, "fav_toggle", "password123")

	resp, err := client.Get(srv.URL + fmt.Sprintf("/recipes/favorite/%d", recipe.ID))
	if err != nil {
		t.Fatalf("favorite toggle failed: %v", err)
	}
	resp.Body.Close()

	var count int64
	database.DB.Table("user_favorites").Where("user_id = ? AND recipe_id = ?", user.ID, recipe.ID).Count(&count)
	if count == 0 {
		t.Error("recipe should be added to favorites")
	}

	resp2, _ := client.Get(srv.URL + fmt.Sprintf("/recipes/favorite/%d", recipe.ID))
	resp2.Body.Close()

	database.DB.Table("user_favorites").Where("user_id = ? AND recipe_id = ?", user.ID, recipe.ID).Count(&count)
	if count != 0 {
		t.Error("recipe should be removed from favorites on second toggle")
	}
}

func TestFavoritesPage(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "fav_list", "password123", false, false, "")
	adminUser := createTestUser(t, "fav_list_a", "password123", true, true, "admin")
	recipe := createTestRecipe(t, "FavListed", adminUser, true)

	database.DB.Exec("INSERT INTO user_favorites (user_id, recipe_id) VALUES (?, ?)", user.ID, recipe.ID)

	client := newClient()
	loginUser(t, srv, client, "fav_list", "password123")

	resp, err := client.Get(srv.URL + "/favorites")
	if err != nil {
		t.Fatalf("favorites request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusFound {
		t.Errorf("favorites page should return 200, got redirect")
	}
}

func TestCart(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "cart_user", "password123", false, false, "")
	adminUser := createTestUser(t, "cart_admin", "password123", true, true, "admin")
	price := 5.0
	recipe := createTestRecipe(t, "CartRecipe", adminUser, true)
	recipe.Price = &price
	database.DB.Save(recipe)

	client := newClient()
	loginUser(t, srv, client, "cart_user", "password123")

	resp, err := client.Get(srv.URL + fmt.Sprintf("/cart/add/%d", recipe.ID))
	if err != nil {
		t.Fatalf("cart add failed: %v", err)
	}
	resp.Body.Close()

	var items []database.CartItem
	database.DB.Where("user_id = ?", user.ID).Find(&items)
	if len(items) != 1 {
		t.Fatalf("expected 1 cart item, got %d", len(items))
	}
	if items[0].RecipeID != recipe.ID {
		t.Error("cart item should reference correct recipe")
	}
	if items[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", items[0].Quantity)
	}

	resp2, _ := client.Get(srv.URL + fmt.Sprintf("/cart/add/%d", recipe.ID))
	resp2.Body.Close()

	database.DB.Where("user_id = ?", user.ID).Find(&items)
	if len(items) != 1 {
		t.Fatal("should be 1 cart item still")
	}
	if items[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", items[0].Quantity)
	}

	resp3, _ := client.Get(srv.URL + fmt.Sprintf("/cart/remove/%d", items[0].ID))
	resp3.Body.Close()

	database.DB.Where("user_id = ?", user.ID).Find(&items)
	if len(items) != 0 {
		t.Error("cart should be empty after removal")
	}
}

func TestCartCheckout(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "checkout_u", "password123", false, false, "")
	adminUser := createTestUser(t, "checkout_a", "password123", true, true, "admin")
	price := 10.0
	recipe := createTestRecipe(t, "CheckoutRec", adminUser, true)
	recipe.Price = &price
	database.DB.Save(recipe)

	client := newClient()
	loginUser(t, srv, client, "checkout_u", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	resp, _ := client.Get(srv.URL + fmt.Sprintf("/cart/add/%d", recipe.ID))
	resp.Body.Close()

	body := url.Values{}
	body.Set("csrf_token", csrfToken)
	resp2, err := client.PostForm(srv.URL+"/cart/checkout", body)
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	resp2.Body.Close()

	var orders []database.Order
	database.DB.Where("user_id = ?", user.ID).Find(&orders)
	if len(orders) == 0 {
		t.Error("order should be created on checkout")
	}

	var items []database.CartItem
	database.DB.Where("user_id = ?", user.ID).Find(&items)
	if len(items) != 0 {
		t.Error("cart should be empty after checkout")
	}
}

func TestOrderTracker(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "track_user", "password123", false, false, "")
	order := database.Order{UserID: user.ID, Status: "PENDING"}
	database.DB.Create(&order)

	resp, _ := newClient().Get(srv.URL + "/orders/track/" + order.TrackingToken)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestOrderTrackerInvalidToken(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	resp, _ := newClient().Get(srv.URL + "/orders/track/nonexistent-token")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestRateRecipe(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "rate_user", "password123", false, false, "")
	adminUser := createTestUser(t, "rate_admin", "password123", true, true, "admin")
	recipe := createTestRecipe(t, "RateRecipe", adminUser, true)

	client := newClient()
	loginUser(t, srv, client, "rate_user", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	body := url.Values{}
	body.Set("score", "8.5")
	body.Set("comment", "Great recipe!")
	body.Set("csrf_token", csrfToken)

	resp, err := client.PostForm(srv.URL+fmt.Sprintf("/recipes/rate/%d", recipe.ID), body)
	if err != nil {
		t.Fatalf("rate request failed: %v", err)
	}
	resp.Body.Close()

	var ratings []database.Rating
	database.DB.Where("recipe_id = ? AND user_id = ?", recipe.ID, user.ID).Find(&ratings)
	if len(ratings) == 0 {
		t.Error("rating should be created")
	} else if ratings[0].Score != 8.5 {
		t.Errorf("expected score 8.5, got %.1f", ratings[0].Score)
	}
}

func TestUserProfile(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "profile_u", "password123", false, false, "")

	client := newClient()
	loginUser(t, srv, client, "profile_u", "password123")

	resp, err := client.Get(srv.URL + "/profile")
	if err != nil {
		t.Fatalf("profile request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestUserProfileUpdate(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "profile_up", "password123", false, false, "")

	client := newClient()
	loginUser(t, srv, client, "profile_up", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	body := url.Values{}
	body.Set("first_name", "Updated")
	body.Set("last_name", "Name")
	body.Set("email", "updated@test.com")
	body.Set("bio", "New bio")
	body.Set("csrf_token", csrfToken)

	resp, err := client.PostForm(srv.URL+"/profile", body)
	if err != nil {
		t.Fatalf("profile update failed: %v", err)
	}
	resp.Body.Close()

	var updated database.User
	database.DB.First(&updated, user.ID)
	if updated.FirstName != "Updated" {
		t.Errorf("expected first_name 'Updated', got '%s'", updated.FirstName)
	}
	if updated.Bio != "New bio" {
		t.Errorf("expected bio 'New bio', got '%s'", updated.Bio)
	}
}

func TestRSSFeed(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	adminUser := createTestUser(t, "rss_admin", "password123", true, true, "admin")
	createTestRecipe(t, "RSSRecipe", adminUser, true)

	resp, _ := newClient().Get(srv.URL + "/feeds/latest")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/rss+xml") {
		t.Errorf("expected RSS content type, got %q", contentType)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "RSSRecipe") {
		t.Error("RSS feed should contain recipe title")
	}
}

func TestCommunitySubmission(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	user := createTestUser(t, "comm_sub", "password123", false, false, "community")

	client := newClient()
	loginUser(t, srv, client, "comm_sub", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	body := url.Values{}
	body.Set("title", "MyCommRecipe")
	body.Set("description", "Community test")
	body.Set("ingredients", "1 cup flour")
	body.Set("instructions", "Mix well")
	body.Set("servings", "4")
	body.Set("csrf_token", csrfToken)

	resp, err := client.PostForm(srv.URL+"/community", body)
	if err != nil {
		t.Fatalf("community submit failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}

	var recipe database.Recipe
	if err := database.DB.Where("title = ?", "MyCommRecipe").First(&recipe).Error; err != nil {
		t.Fatal("community recipe was not created")
	}
	if recipe.IsApproved {
		t.Error("community recipe should not be auto-approved")
	}
	if recipe.UploadedByID == nil || *recipe.UploadedByID != user.ID {
		t.Error("recipe should have correct uploader")
	}
}

func TestAdminDashboard(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "dash_admin", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "dash_admin", "password123")

	resp, err := client.Get(srv.URL + "/dashboard")
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAdminDashboardRequiresStaff(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "regular_u", "password123", false, false, "")

	client := newClient()
	loginUser(t, srv, client, "regular_u", "password123")

	resp, err := client.Get(srv.URL + "/dashboard/")
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("non-staff user should not access dashboard")
	}
}

func TestAdminRecipeCRUD(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "crud_admin", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "crud_admin", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	body := url.Values{}
	body.Set("title", "AdminCreated")
	body.Set("description", "Admin test")
	body.Set("ingredients", "1 cup sugar")
	body.Set("instructions", "Bake")
	body.Set("servings", "3")
	body.Set("price", "5.99")
	body.Set("csrf_token", csrfToken)

	resp, err := client.PostForm(srv.URL+"/dashboard/recipes/add", body)
	if err != nil {
		t.Fatalf("recipe add failed: %v", err)
	}
	resp.Body.Close()

	var recipe database.Recipe
	if err := database.DB.Where("title = ?", "AdminCreated").First(&recipe).Error; err != nil {
		t.Fatal("admin recipe was not created")
	}

	csrfToken = getCSRFToken(t, srv, client)

	editBody := url.Values{}
	editBody.Set("title", "AdminUpdated")
	editBody.Set("description", "Updated description")
	editBody.Set("ingredients", "2 cups sugar")
	editBody.Set("instructions", "Bake longer")
	editBody.Set("servings", "5")
	editBody.Set("price", "10.00")
	editBody.Set("csrf_token", csrfToken)

	resp2, _ := client.PostForm(srv.URL+fmt.Sprintf("/dashboard/recipes/%d/edit", recipe.ID), editBody)
	resp2.Body.Close()

	var updated database.Recipe
	database.DB.First(&updated, recipe.ID)
	if updated.Title != "AdminUpdated" {
		t.Errorf("expected title 'AdminUpdated', got '%s'", updated.Title)
	}

	csrfToken = getCSRFToken(t, srv, client)

	delBody := url.Values{}
	delBody.Set("csrf_token", csrfToken)

	resp3, _ := client.PostForm(srv.URL+fmt.Sprintf("/dashboard/recipes/%d/delete", recipe.ID), delBody)
	resp3.Body.Close()

	var deleted database.Recipe
	if err := database.DB.First(&deleted, recipe.ID).Error; err == nil {
		t.Error("recipe should be deleted")
	}
}

func TestAdminRecipeApproval(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	communityUser := createTestUser(t, "approve_u", "password123", false, false, "community")
	recipe := createTestRecipe(t, "ApprovalRec", communityUser, false)

	createTestUser(t, "approve_a", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "approve_a", "password123")

	resp, err := client.Get(srv.URL + fmt.Sprintf("/dashboard/recipes/%d/approve", recipe.ID))
	if err != nil {
		t.Fatalf("approval request failed: %v", err)
	}
	resp.Body.Close()

	var updated database.Recipe
	database.DB.First(&updated, recipe.ID)
	if !updated.IsApproved {
		t.Error("recipe should be approved")
	}
}

func TestAdminOrderList(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "order_a", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "order_a", "password123")

	resp, err := client.Get(srv.URL + "/dashboard/orders")
	if err != nil {
		t.Fatalf("order list request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAdminUserManagement(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "manage_u", "password123", false, false, "")
	createTestUser(t, "manage_a", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "manage_a", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	resp, _ := client.Get(srv.URL + "/dashboard/users")
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var targetUser database.User
	database.DB.Where("username = ?", "manage_u").First(&targetUser)

	csrfToken = getCSRFToken(t, srv, client)

	editBody := url.Values{}
	editBody.Set("username", "manage_u")
	editBody.Set("email", "updated@test.com")
	editBody.Set("first_name", "UpdatedFirst")
	editBody.Set("last_name", "UpdatedLast")
	editBody.Set("is_staff", "true")
	editBody.Set("csrf_token", csrfToken)

	resp2, _ := client.PostForm(srv.URL+fmt.Sprintf("/dashboard/users/%d/edit", targetUser.ID), editBody)
	resp2.Body.Close()

	var updated database.User
	database.DB.First(&updated, targetUser.ID)
	if updated.FirstName != "UpdatedFirst" {
		t.Errorf("expected first_name 'UpdatedFirst', got '%s'", updated.FirstName)
	}
	if !updated.IsStaff {
		t.Error("user should be staff after update")
	}
}

func TestAdminTagManagement(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "tag_a", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "tag_a", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	addBody := url.Values{}
	addBody.Set("name", "vegetarian")
	addBody.Set("csrf_token", csrfToken)

	resp, _ := client.PostForm(srv.URL+"/dashboard/tags/add", addBody)
	resp.Body.Close()

	var tag database.Tag
	if err := database.DB.Where("name = ?", "vegetarian").First(&tag).Error; err != nil {
		t.Fatal("tag was not created")
	}

	csrfToken = getCSRFToken(t, srv, client)

	editBody := url.Values{}
	editBody.Set("name", "vegan")
	editBody.Set("csrf_token", csrfToken)

	resp2, _ := client.PostForm(srv.URL+fmt.Sprintf("/dashboard/tags/%d/edit", tag.ID), editBody)
	resp2.Body.Close()

	var updated database.Tag
	database.DB.First(&updated, tag.ID)
	if updated.Name != "vegan" {
		t.Errorf("expected tag name 'vegan', got '%s'", updated.Name)
	}
}

func TestAdminSettings(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	createTestUser(t, "set_a", "password123", true, true, "admin")

	client := newClient()
	loginUser(t, srv, client, "set_a", "password123")
	csrfToken := getCSRFToken(t, srv, client)

	body := url.Values{}
	body.Set("site_name", "My Sandwitches")
	body.Set("site_description", "Best sandwiches ever")
	body.Set("email", "contact@test.com")
	body.Set("log_level", "DEBUG")
	body.Set("csrf_token", csrfToken)

	resp, _ := client.PostForm(srv.URL+"/dashboard/settings", body)
	resp.Body.Close()

	var setting database.Setting
	database.DB.First(&setting)
	if setting.SiteName != "My Sandwitches" {
		t.Errorf("expected site name 'My Sandwitches', got '%s'", setting.SiteName)
	}
	if setting.LogLevel != "DEBUG" {
		t.Errorf("expected log level 'DEBUG', got '%s'", setting.LogLevel)
	}
}
