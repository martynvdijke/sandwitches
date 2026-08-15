package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

// createTagDirect inserts a tag straight into the database (bypasses the API).
func createTagDirect(t *testing.T, name string) *database.Tag {
	t.Helper()
	tag := database.Tag{Name: name}
	if err := database.DB.Create(&tag).Error; err != nil {
		t.Fatalf("failed to create tag %q: %v", name, err)
	}
	return &tag
}

func TestAPIPaginationTags(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	for i := 0; i < 55; i++ {
		createTagDirect(t, "page-tag-"+strconv.Itoa(i))
	}

	// No params: plain array (TRMNL compat), still capped at default limit 50.
	w := apiGet(t, r, "/api/v1/tags", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var plain []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &plain); err != nil {
		t.Fatalf("no-param response should be a plain array: %v (%s)", err, w.Body.String())
	}
	if len(plain) != 50 {
		t.Errorf("expected default cap of 50, got %d", len(plain))
	}

	// With params: {items, total} envelope.
	w2 := apiGet(t, r, "/api/v1/tags?limit=10&offset=5", "")
	var env struct {
		Items []map[string]interface{} `json:"items"`
		Total int64                    `json:"total"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &env); err != nil {
		t.Fatalf("paginated response should be an envelope: %v (%s)", err, w2.Body.String())
	}
	if len(env.Items) != 10 {
		t.Errorf("expected 10 items, got %d", len(env.Items))
	}
	if env.Total != 55 {
		t.Errorf("expected total 55, got %d", env.Total)
	}
}

func TestAPIPaginationValidation(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	for _, path := range []string{
		"/api/v1/tags?limit=-1",
		"/api/v1/tags?limit=abc",
		"/api/v1/tags?offset=-3",
		"/api/v1/tags?offset=xyz",
	} {
		w := apiGet(t, r, path, "")
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: expected 422, got %d", path, w.Code)
		}
		var body map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &body)
		if body["code"] != "validation_error" {
			t.Errorf("%s: expected code validation_error, got %v", path, body["code"])
		}
	}
}

func TestAPIPaginationOffsetPastEnd(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		createTagDirect(t, "end-tag-"+strconv.Itoa(i))
	}

	w := apiGet(t, r, "/api/v1/tags?offset=1000", "")
	var env struct {
		Items []map[string]interface{} `json:"items"`
		Total int64                    `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("expected envelope: %v", err)
	}
	if len(env.Items) != 0 {
		t.Errorf("expected empty items past the end, got %d", len(env.Items))
	}
	if env.Total != 5 {
		t.Errorf("expected total 5, got %d", env.Total)
	}
}

func TestAPILimitCapAtMax(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	for i := 0; i < 250; i++ {
		createTagDirect(t, "cap-tag-"+strconv.Itoa(i))
	}

	w := apiGet(t, r, "/api/v1/tags?limit=1000", "")
	var env struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("expected envelope: %v", err)
	}
	if len(env.Items) != 200 {
		t.Errorf("expected cap at 200, got %d", len(env.Items))
	}
}

func TestAPIRecipeFilters(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	chef := createAPIUser(t, "filter_chef", true)
	cookie := loginAPIUser(t, r, "filter_chef")

	recipe := map[string]interface{}{
		"title":        "Alpine Pancake Stack",
		"description":  "Fluffy breakfast pancakes",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     2,
		"tags":         []string{"breakfast"},
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	recipe2 := map[string]interface{}{
		"title":        "Waffle Iron Special",
		"description":  "Crispy waffles",
		"ingredients":  "2 cups flour",
		"instructions": "Cook",
		"servings":     4,
	}
	w2 := apiPost(t, r, "/api/v1/recipes", recipe2, cookie)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &created)
	// Approve the waffle recipe directly so is_approved filter is meaningful.
	var waffle database.Recipe
	database.DB.First(&waffle, created["id"])
	database.DB.Model(&waffle).Update("is_approved", true)

	// search (case-insensitive)
	w3 := apiGet(t, r, "/api/v1/recipes?search=pancake", "")
	var recipes []map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &recipes)
	if len(recipes) != 1 || recipes[0]["title"] != "Alpine Pancake Stack" {
		t.Errorf("search=pancake should match 1 recipe, got %d: %s", len(recipes), w3.Body.String())
	}

	// tag filter
	w4 := apiGet(t, r, "/api/v1/recipes?tag=breakfast", "")
	json.Unmarshal(w4.Body.Bytes(), &recipes)
	if len(recipes) != 1 {
		t.Errorf("tag=breakfast should match 1 recipe, got %d", len(recipes))
	}

	// nonexistent tag → empty, no error
	w5 := apiGet(t, r, "/api/v1/recipes?tag=nope", "")
	if w5.Code != http.StatusOK {
		t.Errorf("nonexistent tag should be 200, got %d", w5.Code)
	}
	json.Unmarshal(w5.Body.Bytes(), &recipes)
	if len(recipes) != 0 {
		t.Errorf("nonexistent tag should return empty, got %d", len(recipes))
	}

	// is_approved
	w6 := apiGet(t, r, "/api/v1/recipes?is_approved=true", "")
	json.Unmarshal(w6.Body.Bytes(), &recipes)
	if len(recipes) != 1 || recipes[0]["title"] != "Waffle Iron Special" {
		t.Errorf("is_approved=true should match the waffle recipe, got %d: %s", len(recipes), w6.Body.String())
	}

	// uploaded_by
	w7 := apiGet(t, r, "/api/v1/recipes?uploaded_by="+itoa(int(chef.ID)), "")
	json.Unmarshal(w7.Body.Bytes(), &recipes)
	if len(recipes) != 2 {
		t.Errorf("uploaded_by=chef should match 2 recipes, got %d", len(recipes))
	}

	// AND combined
	w8 := apiGet(t, r, "/api/v1/recipes?search=waffle&tag=breakfast", "")
	json.Unmarshal(w8.Body.Bytes(), &recipes)
	if len(recipes) != 0 {
		t.Errorf("AND combination should match 0, got %d", len(recipes))
	}

	// invalid is_approved / uploaded_by → 422
	w9 := apiGet(t, r, "/api/v1/recipes?is_approved=maybe", "")
	if w9.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid is_approved should be 422, got %d", w9.Code)
	}
	w10 := apiGet(t, r, "/api/v1/recipes?uploaded_by=abc", "")
	if w10.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid uploaded_by should be 422, got %d", w10.Code)
	}
}

func TestAPIETagRecipeOfTheDay(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "etag_chef", true)
	cookie := loginAPIUser(t, r, "etag_chef")

	recipe := map[string]interface{}{
		"title":        "ETag Daily Special",
		"description":  "Cached daily pick",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     2,
	}
	if w := apiPost(t, r, "/api/v1/recipes", recipe, cookie); w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w1 := apiGet(t, r, "/api/v1/recipe-of-the-day", "")
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header on recipe-of-the-day")
	}

	req, _ := http.NewRequest("GET", "/api/v1/recipe-of-the-day", nil)
	req.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code != http.StatusNotModified {
		t.Errorf("expected 304 with matching If-None-Match, got %d", w2.Code)
	}
}

func TestAPIETagRecipesList(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "etag_list_chef", true)
	cookie := loginAPIUser(t, r, "etag_list_chef")

	recipe := map[string]interface{}{
		"title":        "ETag List Item",
		"description":  "For list etag",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     2,
	}
	if w := apiPost(t, r, "/api/v1/recipes", recipe, cookie); w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w1 := apiGet(t, r, "/api/v1/recipes", "")
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header on recipes list")
	}

	req, _ := http.NewRequest("GET", "/api/v1/recipes", nil)
	req.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code != http.StatusNotModified {
		t.Errorf("expected 304 with matching If-None-Match, got %d", w2.Code)
	}
}

func TestAPIRequestID(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/ok", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/err", func(c *gin.Context) {
		c.JSON(400, gin.H{"message": "bad", "code": "bad_request", "request_id": middleware.GetRequestID(c)})
	})

	req, _ := http.NewRequest("GET", "/ok", nil)
	req.Header.Set("X-Request-Id", "client-supplied-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-Id"); got != "client-supplied-id" {
		t.Errorf("expected client-supplied X-Request-Id echoed, got %q", got)
	}

	// Generated id when none supplied.
	req2, _ := http.NewRequest("GET", "/ok", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if got := w2.Header().Get("X-Request-Id"); got == "" {
		t.Error("expected generated X-Request-Id header")
	}
}

func TestAPIRateLimit(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RateLimit(0, 2))
	r.GET("/limited", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	var lastCode int
	var lastRetry string
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/limited", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		lastCode = w.Code
		lastRetry = w.Header().Get("Retry-After")
	}
	if lastCode != http.StatusTooManyRequests {
		t.Errorf("3rd request should be 429, got %d", lastCode)
	}
	if lastRetry == "" {
		t.Error("expected Retry-After header on 429")
	}
}

func TestAPICORS(t *testing.T) {
	r := gin.New()
	r.Use(middleware.CORS("https://allowed.example"))
	r.GET("/cors", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Allowed origin → ACAO echoed.
	req, _ := http.NewRequest("GET", "/cors", nil)
	req.Header.Set("Origin", "https://allowed.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example" {
		t.Errorf("expected ACAO for allowed origin, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Allow-Credentials true, got %q", got)
	}

	// Preflight → 204 with allowed methods.
	pre, _ := http.NewRequest("OPTIONS", "/cors", nil)
	pre.Header.Set("Origin", "https://allowed.example")
	pre.Header.Set("Access-Control-Request-Method", "GET")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, pre)
	if w2.Code != http.StatusNoContent {
		t.Errorf("expected preflight 204, got %d", w2.Code)
	}
	if got := w2.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Errorf("expected GET in Allow-Methods, got %q", got)
	}

	// Disallowed origin → no ACAO header.
	bad, _ := http.NewRequest("GET", "/cors", nil)
	bad.Header.Set("Origin", "https://evil.example")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, bad)
	if got := w3.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no ACAO for disallowed origin, got %q", got)
	}
}
