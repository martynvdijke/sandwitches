package api

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/openapi.json", openapiJSON)
	r.GET("/docs", swaggerUI)
	r.GET("/redoc", redocUI)

	api := r.Group("/v1")
	{
		api.GET("/ping", ping)
		api.GET("/settings", getSettings)
		api.POST("/settings", middleware.AuthRequired(), updateSettings)

		api.GET("/me", middleware.AuthRequired(), me)

		api.GET("/users", users)

		api.GET("/recipes", getRecipes)
		api.POST("/recipes", middleware.AuthRequired(), createRecipe)
		api.GET("/recipes/:id", getRecipe)
		api.PATCH("/recipes/:id", middleware.AuthRequired(), updateRecipe)
		api.DELETE("/recipes/:id", middleware.AuthRequired(), deleteRecipe)
		api.GET("/recipes/:id/scale-ingredients", scaleIngredients)
		api.GET("/recipes/:id/rating", getRecipeRating)
		api.POST("/recipes/:id/ratings", middleware.AuthRequired(), createRating)

		api.GET("/recipe-of-the-day", recipeOfTheDay)

		api.GET("/tags", getTags)
		api.POST("/tags", middleware.AuthRequired(), createTag)
		api.GET("/tags/:id", getTag)
		api.PATCH("/tags/:id", middleware.AuthRequired(), updateTag)
		api.DELETE("/tags/:id", middleware.AuthRequired(), deleteTag)

		api.GET("/orders", middleware.AuthRequired(), getOrders)
		api.POST("/orders", middleware.AuthRequired(), createOrder)
		api.GET("/orders/:id", middleware.AuthRequired(), getOrder)
		api.PATCH("/orders/:id", middleware.AuthRequired(), updateOrderStatus)

		api.GET("/cart", middleware.AuthRequired(), getCart)
		api.POST("/cart", middleware.AuthRequired(), addToCartAPI)
		api.PATCH("/cart/:id", middleware.AuthRequired(), updateCartAPI)
		api.DELETE("/cart/:id", middleware.AuthRequired(), deleteCartAPI)
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "pong"})
}

func getSettings(c *gin.Context) {
	var s database.Setting
	database.DB.First(&s)
	c.JSON(http.StatusOK, settingToDTO(&s))
}

func updateSettings(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil || !user.IsStaff {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	var s database.Setting
	database.DB.First(&s)
	var input SettingUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	applySettingUpdates(&s, input)
	database.DB.Save(&s)
	c.JSON(http.StatusOK, settingToDTO(&s))
}

func me(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	c.JSON(http.StatusOK, userToDTO(user))
}

func users(c *gin.Context) {
	query := database.DB.Model(&database.User{})
	var total int64
	query.Count(&total)

	paginated := wantsPagination(c)
	pg, ok := parsePagination(c)
	if !ok {
		return
	}

	var users []database.User
	query.Limit(pg.limit).Offset(pg.offset).Find(&users)

	items := usersToDTOs(users)
	if paginated {
		c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
		return
	}
	c.JSON(http.StatusOK, items)
}

type RecipeCreateInput struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description"`
	Ingredients  string   `json:"ingredients"`
	Instructions string   `json:"instructions"`
	Servings     int      `json:"servings"`
	Price        *float64 `json:"price"`
	Tags         []string `json:"tags"`
}

type RecipeUpdateInput struct {
	Title        *string   `json:"title"`
	Description  *string   `json:"description"`
	Ingredients  *string   `json:"ingredients"`
	Instructions *string   `json:"instructions"`
	Servings     *int      `json:"servings"`
	Price        *float64  `json:"price"`
	Tags         *[]string `json:"tags"`
}

func recipeToJSON(r *database.Recipe) gin.H {
	tags := make([]gin.H, len(r.Tags))
	for i, t := range r.Tags {
		tags[i] = gin.H{"id": t.ID, "name": t.Name, "slug": t.Slug}
	}
	favoriteIDs := make([]uint, len(r.FavoritedBy))
	for i, u := range r.FavoritedBy {
		favoriteIDs[i] = u.ID
	}
	var avgRating float64
	database.DB.Model(&database.Rating{}).Where("recipe_id = ?", r.ID).Select("COALESCE(AVG(score), 0)").Scan(&avgRating)

	return gin.H{
		"id":                 r.ID,
		"title":              r.Title,
		"slug":               r.Slug,
		"description":        r.Description,
		"ingredients":        r.Ingredients,
		"instructions":       r.Instructions,
		"servings":           r.Servings,
		"price":              r.Price,
		"image":              r.Image,
		"image_thumbnail":    r.ImageThumbnail,
		"image_small":        r.ImageSmall,
		"image_medium":       r.ImageMedium,
		"image_large":        r.ImageLarge,
		"is_highlighted":     r.IsHighlighted,
		"is_approved":        r.IsApproved,
		"prep_time":          r.PrepTime,
		"cook_time":          r.CookTime,
		"calories":           r.Calories,
		"uploaded_by":        r.UploadedByID,
		"tags":               tags,
		"favorited_by":       favoriteIDs,
		"average_rating":     math.Round(avgRating*10) / 10,
		"daily_orders_count": r.DailyOrdersCount,
		"created_at":         r.CreatedAt,
		"updated_at":         r.UpdatedAt,
	}
}

func getRecipes(c *gin.Context) {
	query := database.DB.Model(&database.Recipe{})

	// Filters (AND semantics).
	if s := c.Query("search"); s != "" {
		like := "%" + s + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(ingredients) LIKE ?",
			strings.ToLower(like), strings.ToLower(like), strings.ToLower(like))
	}
	if tagSlug := c.Query("tag"); tagSlug != "" {
		query = query.Joins("JOIN recipe_tags rt ON rt.recipe_id = recipes.id").
			Joins("JOIN tags t ON t.id = rt.tag_id").
			Where("t.slug = ?", tagSlug)
	}
	if v := c.Query("is_approved"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			apiError(c, http.StatusUnprocessableEntity, "is_approved must be true or false")
			return
		}
		query = query.Where("is_approved = ?", b)
	}
	if v := c.Query("uploaded_by"); v != "" {
		uid, err := strconv.Atoi(v)
		if err != nil {
			apiError(c, http.StatusUnprocessableEntity, "uploaded_by must be an integer")
			return
		}
		query = query.Where("uploaded_by_id = ?", uid)
	}

	var total int64
	query.Count(&total)

	paginated := wantsPagination(c)
	pg, ok := parsePagination(c)
	if !ok {
		return
	}

	var recipes []database.Recipe
	query.Preload("Tags").Preload("FavoritedBy").Limit(pg.limit).Offset(pg.offset).Find(&recipes)

	result := make([]gin.H, len(recipes))
	for i, r := range recipes {
		result[i] = recipeToJSON(&r)
	}

	if paginated {
		serveWithETag(c, gin.H{"items": result, "total": total})
		return
	}
	serveWithETag(c, result)
}

func createRecipe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	var input RecipeCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}

	r := database.Recipe{
		Title:        input.Title,
		Description:  input.Description,
		Ingredients:  input.Ingredients,
		Instructions: input.Instructions,
		Servings:     input.Servings,
		Price:        input.Price,
	}
	r.UploadedByID = &user.ID
	database.DB.Create(&r)

	for _, tagName := range input.Tags {
		var tag database.Tag
		database.DB.Where("name = ?", tagName).FirstOrCreate(&tag, database.Tag{Name: tagName})
		database.DB.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", r.ID, tag.ID)
	}

	database.DB.Preload("Tags").Preload("FavoritedBy").First(&r, r.ID)
	c.JSON(http.StatusCreated, recipeToJSON(&r))
}

func getRecipe(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var r database.Recipe
	if err := database.DB.Preload("Tags").Preload("FavoritedBy").First(&r, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Recipe not found")
		return
	}
	c.JSON(http.StatusOK, recipeToJSON(&r))
}

func updateRecipe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var r database.Recipe
	if err := database.DB.First(&r, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Recipe not found")
		return
	}
	if !user.IsStaff && (r.UploadedByID == nil || *r.UploadedByID != user.ID) {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}

	var input RecipeUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Ingredients != nil {
		updates["ingredients"] = *input.Ingredients
	}
	if input.Instructions != nil {
		updates["instructions"] = *input.Instructions
	}
	if input.Servings != nil {
		updates["servings"] = *input.Servings
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	database.DB.Model(&r).Updates(updates)

	if input.Tags != nil {
		database.DB.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", r.ID)
		for _, tagName := range *input.Tags {
			var tag database.Tag
			database.DB.Where("name = ?", tagName).FirstOrCreate(&tag, database.Tag{Name: tagName})
			database.DB.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", r.ID, tag.ID)
		}
	}

	database.DB.Preload("Tags").Preload("FavoritedBy").First(&r, id)
	c.JSON(http.StatusOK, recipeToJSON(&r))
}

func deleteRecipe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var r database.Recipe
	if err := database.DB.First(&r, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Recipe not found")
		return
	}
	if !user.IsStaff && (r.UploadedByID == nil || *r.UploadedByID != user.ID) {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	database.DB.Delete(&r)
	c.JSON(http.StatusNoContent, nil)
}

type ScaledIngredient struct {
	OriginalLine string   `json:"original_line"`
	ScaledLine   string   `json:"scaled_line"`
	Quantity     *float64 `json:"quantity"`
	Unit         string   `json:"unit"`
	Name         string   `json:"name"`
}

func scaleIngredients(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	targetStr := c.Query("target_servings")
	if targetStr == "" {
		apiError(c, http.StatusUnprocessableEntity, "target_servings is required")
		return
	}
	target, err := strconv.Atoi(targetStr)
	if err != nil {
		apiError(c, http.StatusUnprocessableEntity, "target_servings must be an integer")
		return
	}
	if target < 1 {
		target = 1
	}

	var r database.Recipe
	if err := database.DB.First(&r, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Recipe not found")
		return
	}

	current := r.Servings
	if current <= 0 {
		current = 1
	}

	lines := utils.ParseIngredientLines(r.Ingredients)
	result := make([]ScaledIngredient, len(lines))
	for i, p := range lines {
		scaled := utils.ScaleIngredient(p, float64(current), float64(target))
		result[i] = ScaledIngredient{
			OriginalLine: p.OriginalLine,
			ScaledLine:   utils.FormatScaledIngredient(scaled),
			Quantity:     &scaled.Quantity,
			Unit:         scaled.Unit,
			Name:         scaled.Name,
		}
	}
	c.JSON(http.StatusOK, result)
}

func getRecipeRating(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var avg float64
	var count int64
	database.DB.Model(&database.Rating{}).Where("recipe_id = ?", id).Select("COALESCE(AVG(score), 0)").Scan(&avg)
	database.DB.Model(&database.Rating{}).Where("recipe_id = ?", id).Count(&count)
	c.JSON(http.StatusOK, gin.H{"average": math.Round(avg*10) / 10, "count": count})
}

type RatingInput struct {
	Score   float64 `json:"score" binding:"required"`
	Comment string  `json:"comment"`
}

func createRating(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var input RatingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}

	var rating database.Rating
	database.DB.Where("recipe_id = ? AND user_id = ?", id, user.ID).First(&rating)
	rating.RecipeID = uint(id)
	rating.UserID = user.ID
	rating.Score = input.Score
	rating.Comment = input.Comment
	database.DB.Save(&rating)

	rating.User = *user
	c.JSON(http.StatusCreated, ratingToDTO(&rating))
}

func recipeOfTheDay(c *gin.Context) {
	now := time.Now()
	dayKey := strconv.Itoa(now.Year()) + "-" + strconv.Itoa(now.YearDay())

	var count int64
	database.DB.Model(&database.Recipe{}).Count(&count)

	recipeOfTheDayCache.mu.Lock()
	if recipeOfTheDayCache.day == dayKey && recipeOfTheDayCache.count == count && recipeOfTheDayCache.body != nil {
		body := recipeOfTheDayCache.body
		recipeOfTheDayCache.mu.Unlock()
		serveETagBody(c, body)
		return
	}
	recipeOfTheDayCache.mu.Unlock()

	var payload interface{}
	var recipes []database.Recipe
	database.DB.Preload("Tags").Preload("FavoritedBy").Find(&recipes)
	if len(recipes) == 0 {
		payload = nil
	} else {
		// Seed deterministically from the local calendar day so the "recipe of the
		// day" is stable for all requests within one local day (v2.x seeded with
		// date.today().toordinal()).
		seed := int64(now.Year())*1000 + int64(now.YearDay())
		rng := rand.New(rand.NewSource(seed))
		r := recipes[rng.Intn(len(recipes))]
		payload = recipeToJSON(&r)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		apiError(c, http.StatusInternalServerError, "Failed to encode response")
		return
	}

	recipeOfTheDayCache.mu.Lock()
	recipeOfTheDayCache.day = dayKey
	recipeOfTheDayCache.count = count
	recipeOfTheDayCache.body = body
	recipeOfTheDayCache.mu.Unlock()

	serveETagBody(c, body)
}

func getTags(c *gin.Context) {
	query := database.DB.Model(&database.Tag{})
	var total int64
	query.Count(&total)

	paginated := wantsPagination(c)
	pg, ok := parsePagination(c)
	if !ok {
		return
	}

	var tags []database.Tag
	query.Limit(pg.limit).Offset(pg.offset).Find(&tags)

	items := tagsToDTOs(tags)
	if paginated {
		c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
		return
	}
	c.JSON(http.StatusOK, items)
}

func createTag(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil || !user.IsStaff {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	tag := database.Tag{Name: input.Name}
	database.DB.Create(&tag)
	c.JSON(http.StatusCreated, tagToDTO(&tag))
}

func getTag(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var tag database.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Tag not found")
		return
	}
	c.JSON(http.StatusOK, tagToDTO(&tag))
}

func updateTag(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil || !user.IsStaff {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var tag database.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Tag not found")
		return
	}
	var input struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	tag.Name = input.Name
	database.DB.Save(&tag)
	c.JSON(http.StatusOK, tagToDTO(&tag))
}

func deleteTag(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil || !user.IsStaff {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Delete(&database.Tag{}, id)
	c.JSON(http.StatusNoContent, nil)
}

func getOrders(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	query := database.DB.Model(&database.Order{}).Preload("Items.Recipe")
	if !user.IsStaff {
		query = query.Where("user_id = ?", user.ID)
	}
	var total int64
	query.Count(&total)

	paginated := wantsPagination(c)
	pg, ok := parsePagination(c)
	if !ok {
		return
	}

	var orders []database.Order
	query.Limit(pg.limit).Offset(pg.offset).Find(&orders)

	items := ordersToDTOs(orders)
	if paginated {
		c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
		return
	}
	c.JSON(http.StatusOK, items)
}

type CreateOrderInput struct {
	RecipeID int `json:"recipe_id" binding:"required"`
}

func createOrder(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}

	var recipe database.Recipe
	if err := database.DB.First(&recipe, input.RecipeID).Error; err != nil {
		apiError(c, http.StatusBadRequest, "Recipe not found")
		return
	}

	price := 0.0
	if recipe.Price != nil {
		price = *recipe.Price
	}

	tx := database.DB.Begin()
	order := database.Order{UserID: user.ID, Status: "PENDING", TotalPrice: price}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		apiError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}
	if err := tx.Create(&database.OrderItem{OrderID: order.ID, RecipeID: recipe.ID, Quantity: 1, Price: price}).Error; err != nil {
		tx.Rollback()
		apiError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}
	if err := tx.Commit().Error; err != nil {
		apiError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}

	c.JSON(http.StatusCreated, orderToDTO(&order))
}

func getOrder(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var order database.Order
	query := database.DB.Preload("Items.Recipe")
	if !user.IsStaff {
		query = query.Where("user_id = ?", user.ID)
	}
	if err := query.First(&order, id).Error; err != nil {
		apiError(c, http.StatusNotFound, "Order not found")
		return
	}
	c.JSON(http.StatusOK, orderToDetailDTO(&order))
}

func updateOrderStatus(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil || !user.IsStaff {
		apiError(c, http.StatusForbidden, "Not authorized")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	database.DB.Model(&database.Order{}).Where("id = ?", id).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"status": input.Status})
}

func getCart(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	query := database.DB.Model(&database.CartItem{}).
		Preload("Recipe.Tags").Preload("Recipe.FavoritedBy").
		Where("user_id = ?", user.ID)
	var total int64
	query.Count(&total)

	paginated := wantsPagination(c)
	pg, ok := parsePagination(c)
	if !ok {
		return
	}

	var items []database.CartItem
	query.Limit(pg.limit).Offset(pg.offset).Find(&items)

	dto := cartItemsToDTOs(items)
	if paginated {
		c.JSON(http.StatusOK, gin.H{"items": dto, "total": total})
		return
	}
	c.JSON(http.StatusOK, dto)
}

type CartItemInput struct {
	RecipeID int `json:"recipe_id" binding:"required"`
	Quantity int `json:"quantity"`
}

func addToCartAPI(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	var input CartItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	if input.Quantity < 1 {
		input.Quantity = 1
	}

	var item database.CartItem
	result := database.DB.Where("user_id = ? AND recipe_id = ?", user.ID, input.RecipeID).First(&item)
	if result.Error == gorm.ErrRecordNotFound {
		item = database.CartItem{UserID: user.ID, RecipeID: uint(input.RecipeID), Quantity: input.Quantity}
		database.DB.Create(&item)
	} else {
		item.Quantity += input.Quantity
		database.DB.Save(&item)
	}

	database.DB.Preload("Recipe.Tags").Preload("Recipe.FavoritedBy").First(&item, item.ID)
	c.JSON(http.StatusCreated, cartItemToDTO(&item))
}

type CartItemUpdateInput struct {
	Quantity int `json:"quantity" binding:"required"`
}

func updateCartAPI(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var item database.CartItem
	if err := database.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&item).Error; err != nil {
		apiError(c, http.StatusNotFound, "Cart item not found")
		return
	}
	var input CartItemUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	item.Quantity = input.Quantity
	database.DB.Save(&item)

	database.DB.Preload("Recipe.Tags").Preload("Recipe.FavoritedBy").First(&item, item.ID)
	c.JSON(http.StatusOK, cartItemToDTO(&item))
}

func deleteCartAPI(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&database.CartItem{})
	c.JSON(http.StatusNoContent, nil)
}
