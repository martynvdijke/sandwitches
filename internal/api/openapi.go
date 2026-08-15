package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// buildOpenAPISpec returns the OpenAPI 3.1 document describing the /api/v1
// endpoints. Hand-maintained: keep in sync with RegisterRoutes and the DTOs.
func buildOpenAPISpec() map[string]interface{} {
	ref := func(name string) map[string]interface{} {
		return map[string]interface{}{"$ref": "#/components/schemas/" + name}
	}
	sch := func(typ string) map[string]interface{} { return map[string]interface{}{"type": typ} }

	auth := []map[string]interface{}{{"sessionAuth": []interface{}{}}}

	responses := func(desc string, schema interface{}) map[string]interface{} {
		out := map[string]interface{}{
			"200": map[string]interface{}{"description": desc, "content": map[string]interface{}{
				"application/json": map[string]interface{}{"schema": schema},
			}},
		}
		if schema == nil {
			out["200"] = map[string]interface{}{"description": desc}
		}
		return out
	}

	messageSchema := func() map[string]interface{} {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message":    map[string]interface{}{"type": "string"},
				"code":       map[string]interface{}{"type": "string", "description": "Machine-readable error code (bad_request, validation_error, not_found, unauthorized, forbidden, rate_limited, internal)"},
				"request_id": map[string]interface{}{"type": "string", "description": "Correlates with the X-Request-Id response header"},
			},
		}
	}
	errorResponses := map[string]interface{}{
		"400": map[string]interface{}{"description": "Bad request", "content": map[string]interface{}{"application/json": map[string]interface{}{"schema": messageSchema()}}},
		"401": map[string]interface{}{"description": "Authentication required", "content": map[string]interface{}{"application/json": map[string]interface{}{"schema": messageSchema()}}},
		"403": map[string]interface{}{"description": "Forbidden", "content": map[string]interface{}{"application/json": map[string]interface{}{"schema": messageSchema()}}},
		"404": map[string]interface{}{"description": "Not found", "content": map[string]interface{}{"application/json": map[string]interface{}{"schema": messageSchema()}}},
		"422": map[string]interface{}{"description": "Validation error", "content": map[string]interface{}{"application/json": map[string]interface{}{"schema": messageSchema()}}},
	}

	idParam := []map[string]interface{}{{
		"name": "id", "in": "path", "required": true, "description": "Resource ID",
		"schema": map[string]interface{}{"type": "integer"},
	}}

	// queryParam builds a non-required query string parameter.
	queryParam := func(name, desc, typ string) map[string]interface{} {
		return map[string]interface{}{"name": name, "in": "query", "required": false, "description": desc,
			"schema": map[string]interface{}{"type": typ}}
	}

	paginationParams := []map[string]interface{}{
		queryParam("limit", "Maximum number of results (default 50, max 200)", "integer"),
		queryParam("offset", "Number of results to skip (default 0)", "integer"),
	}

	recipeFilterParams := []map[string]interface{}{
		queryParam("search", "Case-insensitive search over title, description and ingredients", "string"),
		queryParam("tag", "Filter by tag slug", "string"),
		queryParam("is_approved", "Filter by approval status (true or false)", "boolean"),
		queryParam("uploaded_by", "Filter by uploader user id", "integer"),
	}

	jsonBody := func(schema interface{}) map[string]interface{} {
		return map[string]interface{}{"required": true, "content": map[string]interface{}{
			"application/json": map[string]interface{}{"schema": schema},
		}}
	}

	return map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "Sandwitches API",
			"description": "Recipe collection, ordering and community API (v2.x parity). Every response carries an X-Request-Id header; every error is a JSON object with message, code and request_id fields. List endpoints accept limit/offset pagination and return {items, total} when pagination parameters are present.",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{{"url": "/api/v1"}},
		"paths": map[string]interface{}{
			"/ping": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check",
					"operationId": "ping",
					"tags":        []string{"System"},
					"responses": responses("Pong", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"status":  map[string]interface{}{"type": "string"},
							"message": map[string]interface{}{"type": "string"},
						},
					}),
				},
			},
			"/settings": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get site settings",
					"operationId": "getSettings",
					"tags":        []string{"Settings"},
					"responses":   responses("Settings", ref("Setting")),
				},
				"post": map[string]interface{}{
					"summary":     "Update site settings (staff only)",
					"operationId": "updateSettings",
					"tags":        []string{"Settings"},
					"security":    auth,
					"requestBody": jsonBody(ref("Setting")),
					"responses": map[string]interface{}{
						"200": responses("Updated settings", ref("Setting"))["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
					},
				},
			},
			"/me": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Current user",
					"operationId": "me",
					"tags":        []string{"Users"},
					"security":    auth,
					"responses": map[string]interface{}{
						"200": responses("Current user", ref("User"))["200"],
						"401": errorResponses["401"],
					},
				},
			},
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List users",
					"operationId": "users",
					"tags":        []string{"Users"},
					"parameters":  paginationParams,
					"responses":   responses("Users", map[string]interface{}{"type": "array", "items": ref("User")}),
				},
			},
			"/recipes": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List recipes",
					"operationId": "getRecipes",
					"tags":        []string{"Recipes"},
					"parameters":  append(paginationParams, recipeFilterParams...),
					"responses":   responses("Recipes", map[string]interface{}{"type": "array", "items": ref("Recipe")}),
				},
				"post": map[string]interface{}{
					"summary":     "Create recipe",
					"operationId": "createRecipe",
					"tags":        []string{"Recipes"},
					"security":    auth,
					"requestBody": jsonBody(ref("RecipeCreate")),
					"responses": map[string]interface{}{
						"201": responses("Created recipe", ref("Recipe"))["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
					},
				},
			},
			"/recipes/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get recipe",
					"operationId": "getRecipe",
					"tags":        []string{"Recipes"},
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"200": responses("Recipe", ref("Recipe"))["200"],
						"404": errorResponses["404"],
					},
				},
				"patch": map[string]interface{}{
					"summary":     "Update recipe (owner or staff)",
					"operationId": "updateRecipe",
					"tags":        []string{"Recipes"},
					"security":    auth,
					"parameters":  idParam,
					"requestBody": jsonBody(ref("RecipeUpdate")),
					"responses": map[string]interface{}{
						"200": responses("Updated recipe", ref("Recipe"))["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
						"404": errorResponses["404"],
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete recipe (owner or staff)",
					"operationId": "deleteRecipe",
					"tags":        []string{"Recipes"},
					"security":    auth,
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"204": map[string]interface{}{"description": "Deleted"},
						"403": errorResponses["403"],
						"404": errorResponses["404"],
					},
				},
			},
			"/recipes/{id}/scale-ingredients": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Scale ingredient lines to a target serving count",
					"operationId": "scaleIngredients",
					"tags":        []string{"Recipes"},
					"parameters": append(idParam, map[string]interface{}{
						"name": "target_servings", "in": "query", "required": true,
						"description": "Target number of servings",
						"schema":      map[string]interface{}{"type": "integer"},
					}),
					"responses": map[string]interface{}{
						"200": responses("Scaled ingredients", map[string]interface{}{"type": "array", "items": ref("ScaledIngredient")})["200"],
						"404": errorResponses["404"],
						"422": errorResponses["422"],
					},
				},
			},
			"/recipes/{id}/rating": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Recipe rating summary",
					"operationId": "getRecipeRating",
					"tags":        []string{"Recipes"},
					"parameters":  idParam,
					"responses": responses("Rating summary", map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"average": map[string]interface{}{"type": "number"},
							"count":   map[string]interface{}{"type": "integer"},
						},
					}),
				},
			},
			"/recipes/{id}/ratings": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Create or update a rating",
					"operationId": "createRating",
					"tags":        []string{"Recipes"},
					"security":    auth,
					"parameters":  idParam,
					"requestBody": jsonBody(ref("RatingCreate")),
					"responses": map[string]interface{}{
						"201": responses("Created rating", ref("Rating"))["200"],
						"400": errorResponses["400"],
						"401": errorResponses["401"],
					},
				},
			},
			"/recipe-of-the-day": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Deterministic recipe for today",
					"operationId": "recipeOfTheDay",
					"tags":        []string{"Recipes"},
					"responses": map[string]interface{}{
						"200": responses("Recipe of the day (or null)", ref("Recipe"))["200"],
					},
				},
			},
			"/tags": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List tags",
					"operationId": "getTags",
					"tags":        []string{"Tags"},
					"parameters":  paginationParams,
					"responses":   responses("Tags", map[string]interface{}{"type": "array", "items": ref("Tag")}),
				},
				"post": map[string]interface{}{
					"summary":     "Create tag (staff only)",
					"operationId": "createTag",
					"tags":        []string{"Tags"},
					"security":    auth,
					"requestBody": jsonBody(map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
						"required":   []string{"name"},
					}),
					"responses": map[string]interface{}{
						"201": responses("Created tag", ref("Tag"))["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
					},
				},
			},
			"/tags/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get tag",
					"operationId": "getTag",
					"tags":        []string{"Tags"},
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"200": responses("Tag", ref("Tag"))["200"],
						"404": errorResponses["404"],
					},
				},
				"patch": map[string]interface{}{
					"summary":     "Update tag (staff only)",
					"operationId": "updateTag",
					"tags":        []string{"Tags"},
					"security":    auth,
					"parameters":  idParam,
					"requestBody": jsonBody(map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
					}),
					"responses": map[string]interface{}{
						"200": responses("Updated tag", ref("Tag"))["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
						"404": errorResponses["404"],
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete tag (staff only)",
					"operationId": "deleteTag",
					"tags":        []string{"Tags"},
					"security":    auth,
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"204": map[string]interface{}{"description": "Deleted"},
						"403": errorResponses["403"],
					},
				},
			},
			"/orders": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List orders (staff: all, otherwise own)",
					"operationId": "getOrders",
					"tags":        []string{"Orders"},
					"security":    auth,
					"parameters":  paginationParams,
					"responses": map[string]interface{}{
						"200": responses("Orders", map[string]interface{}{"type": "array", "items": ref("Order")})["200"],
						"401": errorResponses["401"],
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create order for a recipe",
					"operationId": "createOrder",
					"tags":        []string{"Orders"},
					"security":    auth,
					"requestBody": jsonBody(ref("OrderCreate")),
					"responses": map[string]interface{}{
						"201": responses("Created order", ref("Order"))["200"],
						"400": errorResponses["400"],
						"401": errorResponses["401"],
					},
				},
			},
			"/orders/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get order with items (staff: any, otherwise own)",
					"operationId": "getOrder",
					"tags":        []string{"Orders"},
					"security":    auth,
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"200": responses("Order detail", ref("OrderDetail"))["200"],
						"401": errorResponses["401"],
						"404": errorResponses["404"],
					},
				},
				"patch": map[string]interface{}{
					"summary":     "Update order status (staff only)",
					"operationId": "updateOrderStatus",
					"tags":        []string{"Orders"},
					"security":    auth,
					"parameters":  idParam,
					"requestBody": jsonBody(map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{"status": map[string]interface{}{"type": "string"}},
						"required":   []string{"status"},
					}),
					"responses": map[string]interface{}{
						"200": responses("Updated status", map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{"status": map[string]interface{}{"type": "string"}},
						})["200"],
						"400": errorResponses["400"],
						"403": errorResponses["403"],
					},
				},
			},
			"/cart": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List cart items",
					"operationId": "getCart",
					"tags":        []string{"Cart"},
					"security":    auth,
					"parameters":  paginationParams,
					"responses": map[string]interface{}{
						"200": responses("Cart items", map[string]interface{}{"type": "array", "items": ref("CartItem")})["200"],
						"401": errorResponses["401"],
					},
				},
				"post": map[string]interface{}{
					"summary":     "Add recipe to cart (increments quantity)",
					"operationId": "addToCartAPI",
					"tags":        []string{"Cart"},
					"security":    auth,
					"requestBody": jsonBody(ref("CartItemCreate")),
					"responses": map[string]interface{}{
						"201": responses("Created cart item", ref("CartItem"))["200"],
						"400": errorResponses["400"],
						"401": errorResponses["401"],
					},
				},
			},
			"/cart/{id}": map[string]interface{}{
				"patch": map[string]interface{}{
					"summary":     "Update cart item quantity",
					"operationId": "updateCartAPI",
					"tags":        []string{"Cart"},
					"security":    auth,
					"parameters":  idParam,
					"requestBody": jsonBody(ref("CartItemUpdate")),
					"responses": map[string]interface{}{
						"200": responses("Updated cart item", ref("CartItem"))["200"],
						"400": errorResponses["400"],
						"401": errorResponses["401"],
						"404": errorResponses["404"],
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Remove cart item",
					"operationId": "deleteCartAPI",
					"tags":        []string{"Cart"},
					"security":    auth,
					"parameters":  idParam,
					"responses": map[string]interface{}{
						"204": map[string]interface{}{"description": "Deleted"},
						"401": errorResponses["401"],
					},
				},
			},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"sessionAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "cookie",
					"name": "sandwitches_session",
				},
			},
			"schemas": map[string]interface{}{
				"Setting": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"site_name":           sch("string"),
						"site_description":    sch("string"),
						"email":               sch("string"),
						"ai_connection_point": sch("string"),
						"ai_model":            sch("string"),
						"ai_api_key":          sch("string"),
						"log_level":           sch("string"),
						"gotify_url":          sch("string"),
						"gotify_token":        sch("string"),
						"otel_endpoint":       sch("string"),
						"otel_enabled":        sch("boolean"),
					},
				},
				"User": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":         sch("integer"),
						"username":   sch("string"),
						"first_name": sch("string"),
						"last_name":  sch("string"),
						"avatar":     sch("string"),
						"bio":        sch("string"),
						"language":   sch("string"),
						"theme":      sch("string"),
					},
				},
				"UserPublic": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"username":   sch("string"),
						"first_name": sch("string"),
						"last_name":  sch("string"),
						"avatar":     sch("string"),
					},
				},
				"Tag": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":   sch("integer"),
						"name": sch("string"),
						"slug": sch("string"),
					},
				},
				"Recipe": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":                 sch("integer"),
						"title":              sch("string"),
						"slug":               sch("string"),
						"description":        sch("string"),
						"ingredients":        sch("string"),
						"instructions":       sch("string"),
						"servings":           sch("integer"),
						"price":              map[string]interface{}{"type": "number", "nullable": true},
						"image":              sch("string"),
						"image_thumbnail":    sch("string"),
						"image_small":        sch("string"),
						"image_medium":       sch("string"),
						"image_large":        sch("string"),
						"is_highlighted":     sch("boolean"),
						"is_approved":        sch("boolean"),
						"prep_time":          map[string]interface{}{"type": "integer", "nullable": true},
						"cook_time":          map[string]interface{}{"type": "integer", "nullable": true},
						"calories":           map[string]interface{}{"type": "integer", "nullable": true},
						"uploaded_by":        map[string]interface{}{"type": "integer", "nullable": true},
						"tags":               map[string]interface{}{"type": "array", "items": ref("Tag")},
						"favorited_by":       map[string]interface{}{"type": "array", "items": sch("integer")},
						"average_rating":     sch("number"),
						"daily_orders_count": sch("integer"),
						"created_at":         sch("string"),
						"updated_at":         sch("string"),
					},
				},
				"RecipeCreate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title":        sch("string"),
						"description":  sch("string"),
						"ingredients":  sch("string"),
						"instructions": sch("string"),
						"servings":     sch("integer"),
						"price":        map[string]interface{}{"type": "number", "nullable": true},
						"tags":         map[string]interface{}{"type": "array", "items": sch("string")},
					},
					"required": []string{"title"},
				},
				"RecipeUpdate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title":        sch("string"),
						"description":  sch("string"),
						"ingredients":  sch("string"),
						"instructions": sch("string"),
						"servings":     sch("integer"),
						"price":        map[string]interface{}{"type": "number", "nullable": true},
						"tags":         map[string]interface{}{"type": "array", "items": sch("string")},
					},
				},
				"ScaledIngredient": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"original_line": sch("string"),
						"scaled_line":   sch("string"),
						"quantity":      map[string]interface{}{"type": "number", "nullable": true},
						"unit":          sch("string"),
						"name":          sch("string"),
					},
				},
				"Rating": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":         sch("integer"),
						"recipe":     sch("integer"),
						"score":      sch("number"),
						"comment":    sch("string"),
						"user":       ref("UserPublic"),
						"created_at": sch("string"),
						"updated_at": sch("string"),
					},
				},
				"RatingCreate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"score":   sch("number"),
						"comment": sch("string"),
					},
					"required": []string{"score"},
				},
				"Order": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          sch("integer"),
						"status":      sch("string"),
						"total_price": sch("number"),
						"created_at":  sch("string"),
						"updated_at":  sch("string"),
					},
				},
				"OrderItem": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":       sch("integer"),
						"quantity": sch("integer"),
						"price":    sch("number"),
						"recipe":   ref("Recipe"),
					},
				},
				"OrderDetail": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          sch("integer"),
						"status":      sch("string"),
						"total_price": sch("number"),
						"created_at":  sch("string"),
						"updated_at":  sch("string"),
						"items":       map[string]interface{}{"type": "array", "items": ref("OrderItem")},
					},
				},
				"OrderCreate": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"recipe_id": sch("integer")},
					"required":   []string{"recipe_id"},
				},
				"CartItem": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":         sch("integer"),
						"recipe":     ref("Recipe"),
						"quantity":   sch("integer"),
						"created_at": sch("string"),
						"updated_at": sch("string"),
					},
				},
				"CartItemCreate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"recipe_id": sch("integer"),
						"quantity":  sch("integer"),
					},
					"required": []string{"recipe_id"},
				},
				"CartItemUpdate": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{"quantity": sch("integer")},
					"required":   []string{"quantity"},
				},
			},
		},
	}
}

func openapiJSON(c *gin.Context) {
	c.JSON(http.StatusOK, buildOpenAPISpec())
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sandwitches API — Swagger UI</title>
<link rel="stylesheet" href="/static/vendor/swagger-ui/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="/static/vendor/swagger-ui/swagger-ui-bundle.js"></script>
<script src="/static/vendor/swagger-ui/swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "/api/openapi.json",
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    layout: "StandaloneLayout",
  });
};
</script>
</body>
</html>`

const redocUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sandwitches API — ReDoc</title>
</head>
<body>
<div id="redoc"></div>
<script src="/static/vendor/redoc/redoc.standalone.js"></script>
<script>
Redoc.init("/api/openapi.json", { scrollYOffset: 0 }, document.getElementById("redoc"));
</script>
</body>
</html>`

func swaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerUIHTML)
}

func redocUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, redocUIHTML)
}
