package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
	os.Remove("test_api_db.sqlite3")
	os.Exit(code)
}

func setupAPITest(t *testing.T) (*gin.Engine, func()) {
	t.Helper()

	tmp := t.TempDir()
	dbPath := tmp + "/test.db"

	database.Init(&config.Config{
		DatabaseFile: dbPath,
		Debug:        false,
		SecretKey:    "api-test-key",
		LanguageCode: "en",
	})

	r := gin.New()
	store := cookie.NewStore([]byte("api-test-key"))
	r.Use(sessions.Sessions("sandwitches_session", store))
	r.Use(middleware.OptionalAuth())

	r.GET("/login", func(c *gin.Context) { c.String(200, "") })
	r.POST("/login", func(c *gin.Context) {
		var user database.User
		if err := database.DB.Where("username = ?", c.PostForm("username")).First(&user).Error; err != nil {
			c.String(200, "invalid")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.PostForm("password"))) != nil {
			c.String(200, "invalid")
			return
		}
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Save()
		c.String(200, "ok")
	})

	RegisterRoutes(r.Group("/api"))

	cleanup := func() {
		sqlDB, _ := database.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return r, cleanup
}

func createAPIUser(t *testing.T, username string, isStaff bool) *database.User {
	t.Helper()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := database.User{
		Username:    username,
		Password:    string(hashed),
		Email:       username + "@test.com",
		Language:    "en",
		Theme:       "light",
		IsActive:    true,
		IsStaff:     isStaff,
		DateJoined:  time.Now(),
	}
	database.DB.Create(&user)
	return &user
}

func loginAPIUser(t *testing.T, r *gin.Engine, username string) string {
	t.Helper()
	body := url.Values{}
	body.Set("username", username)
	body.Set("password", "password123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	for _, c := range w.Result().Cookies() {
		if c.Name == "sandwitches_session" {
			return c.Value
		}
	}
	return ""
}

func apiGet(t *testing.T, r *gin.Engine, path string, sessionCookie string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("Content-Type", "application/json")
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: "sandwitches_session", Value: sessionCookie})
	}
	r.ServeHTTP(w, req)
	return w
}

func apiPost(t *testing.T, r *gin.Engine, path string, body interface{}, sessionCookie string) *httptest.ResponseRecorder {
	t.Helper()
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", path, strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: "sandwitches_session", Value: sessionCookie})
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAPIPing(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	w := apiGet(t, r, "/api/v1/ping", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["status"] != "ok" {
		t.Error("ping should return status ok")
	}
}

func TestAPISettings(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	w := apiGet(t, r, "/api/v1/settings", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIMeUnauthorized(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	w := apiGet(t, r, "/api/v1/me", "")
	if w.Code != http.StatusFound {
		t.Errorf("expected 302 redirect, got %d", w.Code)
	}
}

func TestAPIMeAuthorized(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "api_user", false)
	cookie := loginAPIUser(t, r, "api_user")

	w := apiGet(t, r, "/api/v1/me", cookie)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIUsers(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "api_user1", false)
	createAPIUser(t, "api_user2", false)

	w := apiGet(t, r, "/api/v1/users", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAPIRecipes(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "api_chef", true)
	cookie := loginAPIUser(t, r, "api_chef")

	recipe := map[string]interface{}{
		"title":        "API Recipe",
		"description":  "API test",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     2,
		"tags":         []string{"baking"},
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, cookie)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w2 := apiGet(t, r, "/api/v1/recipes", "")
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}

	w3 := apiGet(t, r, "/api/v1/recipe-of-the-day", "")
	if w3.Code != http.StatusOK || w3.Body.String() == "null" {
		t.Error("recipe-of-the-day should return a recipe")
	}
}

func TestAPITags(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "tag_chef", true)
	cookie := loginAPIUser(t, r, "tag_chef")

	w := apiPost(t, r, "/api/v1/tags", map[string]string{"name": "seafood"}, cookie)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w2 := apiGet(t, r, "/api/v1/tags", "")
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}

func TestAPIOrders(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "order_api", false)
	cookie := loginAPIUser(t, r, "order_api")

	createAPIUser(t, "chef_api", true)
	chefCookie := loginAPIUser(t, r, "chef_api")

	recipe := map[string]interface{}{
		"title":        "Orderable Recipe",
		"description":  "For orders",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     1,
		"price":        7.5,
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiPost(t, r, "/api/v1/orders", map[string]int{"recipe_id": recipeID}, cookie)
	if w2.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}

	w3 := apiGet(t, r, "/api/v1/orders", cookie)
	if w3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w3.Code)
	}
}

func TestAPICart(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "cart_api", false)
	cookie := loginAPIUser(t, r, "cart_api")

	createAPIUser(t, "cart_chef", true)
	chefCookie := loginAPIUser(t, r, "cart_chef")

	recipe := map[string]interface{}{
		"title":        "Cartable Recipe",
		"description":  "For cart",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     1,
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiPost(t, r, "/api/v1/cart", map[string]int{"recipe_id": recipeID, "quantity": 2}, cookie)
	if w2.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}

	w3 := apiGet(t, r, "/api/v1/cart", cookie)
	if w3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w3.Code)
	}
}

func TestAPIRatingUnauthorized(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	w := apiPost(t, r, "/api/v1/recipes/1/ratings", map[string]interface{}{
		"score":   8.0,
		"comment": "Delicious!",
	}, "")
	if w.Code != http.StatusFound {
		t.Errorf("rating without auth should return 302 redirect, got %d", w.Code)
	}
}
