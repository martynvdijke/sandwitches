package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// TestAPISettingsShape: settings response uses snake_case keys.
func TestAPISettingsShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	w := apiGet(t, r, "/api/v1/settings", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var s map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &s)
	if _, ok := s["site_name"]; !ok {
		t.Error("settings should expose site_name (snake_case)")
	}
	if _, ok := s["ai_connection_point"]; !ok {
		t.Error("settings should expose ai_connection_point (snake_case)")
	}
	if _, ok := s["SiteName"]; ok {
		t.Error("settings must not leak PascalCase SiteName")
	}
}

// TestAPIMeShape: /me excludes email, exposes public fields.
func TestAPIMeShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "me_shape", false)
	cookie := loginAPIUser(t, r, "me_shape")

	w := apiGet(t, r, "/api/v1/me", cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)
	if _, ok := m["email"]; ok {
		t.Error("/me must not expose email")
	}
	for _, k := range []string{"id", "username", "first_name", "last_name", "avatar", "bio", "language", "theme"} {
		if _, ok := m[k]; !ok {
			t.Errorf("/me should expose %s", k)
		}
	}
}

// TestAPIUsersShape: /users includes id, bio, language, theme.
func TestAPIUsersShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "users_shape_1", false)
	createAPIUser(t, "users_shape_2", false)

	w := apiGet(t, r, "/api/v1/users", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var users []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &users)
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	for _, u := range users {
		if _, ok := u["id"]; !ok {
			t.Error("users should include id")
		}
		if _, ok := u["language"]; !ok {
			t.Error("users should include language")
		}
		if _, ok := u["theme"]; !ok {
			t.Error("users should include theme")
		}
		if _, ok := u["email"]; ok {
			t.Error("users must not expose email")
		}
	}
}

// TestAPITagsShape: tag responses use snake_case and include slug.
func TestAPITagsShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "tags_shape", true)
	cookie := loginAPIUser(t, r, "tags_shape")

	w := apiPost(t, r, "/api/v1/tags", map[string]string{"name": "spicy"}, cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var tag map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &tag)
	if _, ok := tag["slug"]; !ok {
		t.Error("tag should include slug")
	}
	if _, ok := tag["Name"]; ok {
		t.Error("tag must not leak PascalCase Name")
	}
}

// TestAPIOrdersShape: orders list hides tracking_token, user_id, completed.
func TestAPIOrdersShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "orders_shape_user", false)
	cookie := loginAPIUser(t, r, "orders_shape_user")
	createAPIUser(t, "orders_shape_chef", true)
	chefCookie := loginAPIUser(t, r, "orders_shape_chef")

	recipe := map[string]interface{}{
		"title": "Order Shape Recipe", "description": "d", "servings": 1, "price": 9.99,
	}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	apiPost(t, r, "/api/v1/orders", map[string]int{"recipe_id": recipeID}, cookie)

	w2 := apiGet(t, r, "/api/v1/orders", cookie)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var orders []map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &orders)
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	o := orders[0]
	for _, k := range []string{"tracking_token", "user_id", "completed", "UserID", "TrackingToken"} {
		if _, ok := o[k]; ok {
			t.Errorf("order must not expose %s", k)
		}
	}
	for _, k := range []string{"id", "status", "total_price", "created_at", "updated_at"} {
		if _, ok := o[k]; !ok {
			t.Errorf("order should expose %s", k)
		}
	}

	// Order detail includes items with full recipe payload.
	orderID := int(o["id"].(float64))
	w3 := apiGet(t, r, "/api/v1/orders/"+itoa(orderID), cookie)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 for order detail, got %d", w3.Code)
	}
	var detail map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &detail)
	items, ok := detail["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("order detail should include items")
	}
	item := items[0].(map[string]interface{})
	if _, ok := item["recipe"]; !ok {
		t.Error("order item should include full recipe")
	}
	if _, ok := detail["tracking_token"]; ok {
		t.Error("order detail must not expose tracking_token")
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

// TestAPICartShape: cart items are snake_case with full recipe.
func TestAPICartShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "cart_shape_user", false)
	cookie := loginAPIUser(t, r, "cart_shape_user")
	createAPIUser(t, "cart_shape_chef", true)
	chefCookie := loginAPIUser(t, r, "cart_shape_chef")

	recipe := map[string]interface{}{"title": "Cart Shape Recipe", "description": "d", "servings": 1}
	w := apiPost(t, r, "/api/v1/recipes", recipe, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiPost(t, r, "/api/v1/cart", map[string]int{"recipe_id": recipeID, "quantity": 2}, cookie)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}
	var item map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &item)
	if _, ok := item["recipe"]; !ok {
		t.Error("cart item should include recipe")
	}
	if _, ok := item["RecipeID"]; ok {
		t.Error("cart item must not leak RecipeID")
	}
	if item["quantity"].(float64) != 2 {
		t.Errorf("expected quantity 2, got %v", item["quantity"])
	}
}

// TestAPIRecipeIsApprovedFalse: new recipes default to is_approved=false.
func TestAPIRecipeIsApprovedFalse(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "approval_chef", true)
	cookie := loginAPIUser(t, r, "approval_chef")

	w := apiPost(t, r, "/api/v1/recipes", map[string]interface{}{
		"title": "Pending Approval", "description": "d", "servings": 1,
	}, cookie)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var recipe map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &recipe)
	if recipe["is_approved"].(bool) {
		t.Error("new recipes must default to is_approved=false")
	}
}

// TestAPIScaleIngredientsValidation: missing target_servings → 422, explicit <1 clamps to 1.
func TestAPIScaleIngredientsValidation(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "scale_chef", true)
	cookie := loginAPIUser(t, r, "scale_chef")

	w := apiPost(t, r, "/api/v1/recipes", map[string]interface{}{
		"title": "Scale Recipe", "description": "d",
		"ingredients": "2 cups flour", "servings": 2,
	}, cookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiGet(t, r, "/api/v1/recipes/"+itoa(recipeID)+"/scale-ingredients", "")
	if w2.Code != http.StatusUnprocessableEntity {
		t.Errorf("missing target_servings should be 422, got %d", w2.Code)
	}

	w3 := apiGet(t, r, "/api/v1/recipes/"+itoa(recipeID)+"/scale-ingredients?target_servings=1", "")
	if w3.Code != http.StatusOK {
		t.Errorf("target_servings=1 should be 200, got %d", w3.Code)
	}
}

// TestAPIOrderWithoutPrice: ordering a recipe with no price must not panic and
// returns 201 with total_price 0.
func TestAPIOrderWithoutPrice(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "noprice_user", false)
	cookie := loginAPIUser(t, r, "noprice_user")
	createAPIUser(t, "noprice_chef", true)
	chefCookie := loginAPIUser(t, r, "noprice_chef")

	w := apiPost(t, r, "/api/v1/recipes", map[string]interface{}{
		"title": "Free Recipe", "description": "d", "servings": 1,
	}, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiPost(t, r, "/api/v1/orders", map[string]int{"recipe_id": recipeID}, cookie)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}
	var order map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &order)
	if order["total_price"].(float64) != 0 {
		t.Errorf("expected total_price 0, got %v", order["total_price"])
	}
}

// TestAPIRatingShape: created rating includes created_at/updated_at and a full public user.
func TestAPIRatingShape(t *testing.T) {
	r, cleanup := setupAPITest(t)
	defer cleanup()

	createAPIUser(t, "rating_shape_user", false)
	cookie := loginAPIUser(t, r, "rating_shape_user")
	createAPIUser(t, "rating_shape_chef", true)
	chefCookie := loginAPIUser(t, r, "rating_shape_chef")

	w := apiPost(t, r, "/api/v1/recipes", map[string]interface{}{
		"title": "Rating Shape Recipe", "description": "d", "servings": 1,
	}, chefCookie)
	var created map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &created)
	recipeID := int(created["id"].(float64))

	w2 := apiPost(t, r, "/api/v1/recipes/"+itoa(recipeID)+"/ratings", map[string]interface{}{
		"score": 8.0, "comment": "Great",
	}, cookie)
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}
	var rating map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &rating)
	for _, k := range []string{"created_at", "updated_at"} {
		if _, ok := rating[k]; !ok {
			t.Errorf("rating should expose %s", k)
		}
	}
	user, ok := rating["user"].(map[string]interface{})
	if !ok {
		t.Fatal("rating should expose user object")
	}
	if _, ok := user["first_name"]; !ok {
		t.Error("rating user should be full public schema (first_name)")
	}
}
