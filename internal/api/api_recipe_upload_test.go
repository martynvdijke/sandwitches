package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE,
		0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54,
		0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F, 0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59, 0xE7,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

func apiPostMultipart(t *testing.T, r *gin.Engine, path string, fields map[string]string, tags []string, image []byte, sessionCookie string) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, tag := range tags {
		if err := writer.WriteField("tags", tag); err != nil {
			t.Fatal(err)
		}
	}
	if image != nil {
		part, err := writer.CreateFormFile("image", "recipe.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(image); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if sessionCookie != "" {
		req.AddCookie(&http.Cookie{Name: "sandwitches_session", Value: sessionCookie})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAPIStaffCanCreateRecipeWithImage(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	oldMediaRoot := mediaRoot
	SetMediaRoot(t.TempDir())
	defer SetMediaRoot(oldMediaRoot)

	createAPIUser(t, "upload_chef", true)
	cookie := loginAPIUser(t, r, "upload_chef")

	w := apiPostMultipart(t, r, "/api/v1/recipes", map[string]string{
		"title":        "Photo Recipe",
		"description":  "with a photo",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     "2",
		"price":        "4.50",
	}, []string{"upload,photo", "upload"}, tinyPNG(), cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var recipe map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &recipe); err != nil {
		t.Fatal(err)
	}
	if recipe["is_approved"] != true {
		t.Error("staff-created recipe should be auto-approved")
	}
	image, _ := recipe["image"].(string)
	if !strings.HasPrefix(image, "/media/recipes/") {
		t.Fatalf("expected recipe image path, got %q", image)
	}
	if _, err := os.Stat(filepath.Join(mediaRoot, "recipes", filepath.Base(image))); err != nil {
		t.Errorf("expected saved image file: %v", err)
	}
	tags, _ := recipe["tags"].([]interface{})
	if len(tags) != 2 {
		t.Errorf("expected 2 deduplicated tags, got %d: %v", len(tags), tags)
	}
	if int(recipe["uploaded_by"].(float64)) == 0 {
		t.Error("expected uploaded_by to be set")
	}
}

func TestAPIStaffCanCreateRecipeWithoutImage(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "upload_no_image", true)
	cookie := loginAPIUser(t, r, "upload_no_image")

	w := apiPostMultipart(t, r, "/api/v1/recipes", map[string]string{
		"title":    "No Photo Recipe",
		"servings": "3",
	}, nil, nil, cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var recipe map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &recipe)
	if recipe["image"] != "" {
		t.Errorf("expected empty image, got %v", recipe["image"])
	}
	if int(recipe["servings"].(float64)) != 3 {
		t.Errorf("expected servings 3, got %v", recipe["servings"])
	}
}

func TestAPINonStaffCannotCreateRecipe(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "upload_user", false)
	cookie := loginAPIUser(t, r, "upload_user")

	w := apiPost(t, r, "/api/v1/recipes", map[string]interface{}{
		"title": "Not Allowed", "servings": 1,
	}, cookie)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-staff, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAPIMultipartRecipeRejectsBadFields(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "upload_bad_fields", true)
	cookie := loginAPIUser(t, r, "upload_bad_fields")

	w := apiPostMultipart(t, r, "/api/v1/recipes", map[string]string{
		"title": "Bad Price", "price": "not-a-number",
	}, nil, nil, cookie)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad price, got %d: %s", w.Code, w.Body.String())
	}
}
