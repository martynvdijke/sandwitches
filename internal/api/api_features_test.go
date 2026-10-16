package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetRecipeRatings(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "rater1", false)
	cookie1 := loginAPIUser(t, r, "rater1")
	createAPIUser(t, "rater2", false)
	cookie2 := loginAPIUser(t, r, "rater2")
	createAPIUser(t, "chef_ratings", true)
	chefCookie := loginAPIUser(t, r, "chef_ratings")

	recipe := map[string]interface{}{
		"title":        "Ratings Recipe",
		"description":  "For ratings list",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     1,
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	// Create two ratings
	w = apiPost(t, r, "/api/v1/recipes/1/ratings", map[string]interface{}{"score": 4.0, "comment": "good"}, cookie1)
	// Actually use correct id
	_ = id
	// recipe id is 1 since it's first recipe in fresh DB
	w = apiPost(t, r, "/api/v1/recipes/1/ratings", map[string]interface{}{"score": 4.0, "comment": "good"}, cookie1)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	w = apiPost(t, r, "/api/v1/recipes/1/ratings", map[string]interface{}{"score": 5.0, "comment": "great"}, cookie2)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// List ratings
	w = apiGet(t, r, "/api/v1/recipes/1/ratings", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 ratings, got %d: %s", len(list), w.Body.String())
	}

	// Pagination
	w = apiGet(t, r, "/api/v1/recipes/1/ratings?limit=1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var paged map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &paged)
	if _, ok := paged["items"]; !ok {
		t.Fatalf("expected items key in paginated response")
	}
	if int(paged["total"].(float64)) != 2 {
		t.Fatalf("expected total 2, got %v", paged["total"])
	}

	// 404 for missing recipe
	w = apiGet(t, r, "/api/v1/recipes/9999/ratings", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestOrderTrackAndTrackingToken(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "order_user", false)
	cookie := loginAPIUser(t, r, "order_user")
	createAPIUser(t, "chef_order", true)
	chefCookie := loginAPIUser(t, r, "chef_order")

	recipe := map[string]interface{}{
		"title":        "Trackable Recipe",
		"description":  "For track",
		"ingredients":  "1 cup flour",
		"instructions": "Bake",
		"servings":     1,
		"price":        5.0,
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w = apiPost(t, r, "/api/v1/orders", map[string]int{"recipe_id": recipeID}, cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var orderResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &orderResp)
	token, ok := orderResp["tracking_token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected non-empty tracking_token, got %v", orderResp)
	}

	// Check orderToDTO via helper
	// Also test track endpoint with valid token
	w = apiGet(t, r, "/api/v1/orders/track/"+token, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid token, got %d: %s", w.Code, w.Body.String())
	}
	var detail map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &detail)
	if detail["tracking_token"] != token {
		t.Fatalf("expected tracking_token %s, got %v", token, detail["tracking_token"])
	}
	if _, ok := detail["items"]; !ok {
		t.Fatalf("expected items in detail")
	}

	// 404 for unknown token
	w = apiGet(t, r, "/api/v1/orders/track/00000000-0000-0000-0000-000000000000", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown token, got %d", w.Code)
	}
}
