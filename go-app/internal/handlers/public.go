package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"gorm.io/gorm"
)

func Index(c *gin.Context) {
	user := middleware.GetUser(c)

	var recipes []database.Recipe
	query := database.DB.Preload("FavoritedBy").Preload("Tags").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ?", "admin").
		Where("recipes.is_approved = ?", true)

	if q := c.Query("q"); q != "" {
		query = query.Where("recipes.title LIKE ? OR tags.name LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if uploader := c.Query("uploader"); uploader != "" {
		query = query.Where("uploaded_by_id IN (SELECT id FROM users WHERE username = ?)", uploader)
	}
	if tags := c.QueryArray("tag"); len(tags) > 0 {
		query = query.Where("recipes.id IN (SELECT recipe_id FROM recipe_tags JOIN tags ON recipe_tags.tag_id = tags.id WHERE tags.name IN ?)", tags)
	}

	sort := c.Query("sort")
	switch sort {
	case "date_asc":
		query = query.Order("recipes.created_at ASC")
	case "rating":
		query = query.Order("(SELECT COALESCE(AVG(score), 0) FROM ratings WHERE ratings.recipe_id = recipes.id) DESC, recipes.created_at DESC")
	default:
		query = query.Order("recipes.created_at DESC")
	}

	query.Find(&recipes)

	var highlighted []database.Recipe
	database.DB.Preload("FavoritedBy").Where("is_highlighted = ?", true).Find(&highlighted)

	var uploaders []string
	database.DB.Model(&database.User{}).
		Joins("JOIN recipes ON recipes.uploaded_by_id = users.id").
		Distinct("username").Pluck("username", &uploaders)

	var allTags []database.Tag
	database.DB.Find(&allTags)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"recipes":             recipes,
		"highlighted_recipes": highlighted,
		"uploaders":           uploaders,
		"tags":                allTags,
		"selected_tags":       c.QueryArray("tag"),
		"user":                user,
	})
}

func RecipeDetail(c *gin.Context) {
	slug := c.Param("slug")
	user := middleware.GetUser(c)

	var recipe database.Recipe
	if err := database.DB.Preload("Tags").Preload("FavoritedBy").Preload("UploadedBy").
		Preload("Ratings.User").Where("slug = ?", slug).First(&recipe).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Recipe not found"})
		return
	}

	var avgRating float64
	var ratingCount int64
	database.DB.Model(&database.Rating{}).Where("recipe_id = ?", recipe.ID).
		Select("COALESCE(AVG(score), 0)").Scan(&avgRating)
	database.DB.Model(&database.Rating{}).Where("recipe_id = ?", recipe.ID).Count(&ratingCount)

	var userRating *database.Rating
	if user != nil {
		database.DB.Where("recipe_id = ? AND user_id = ?", recipe.ID, user.ID).First(&userRating)
	}

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"recipe":       recipe,
		"avg_rating":   math.Round(avgRating*10) / 10,
		"rating_count": ratingCount,
		"user_rating":  userRating,
		"all_ratings":  recipe.Ratings,
		"user":         user,
	})
}

func ToggleFavorite(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	var count int64
	database.DB.Table("user_favorites").Where("user_id = ? AND recipe_id = ?", user.ID, recipe.ID).Count(&count)

	if count > 0 {
		database.DB.Exec("DELETE FROM user_favorites WHERE user_id = ? AND recipe_id = ?", user.ID, recipe.ID)
	} else {
		database.DB.Exec("INSERT INTO user_favorites (user_id, recipe_id) VALUES (?, ?)", user.ID, recipe.ID)
	}

	referer := c.GetHeader("Referer")
	if referer != "" {
		c.Redirect(http.StatusFound, referer)
		return
	}
	c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
}

func RecipeRate(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	score, _ := strconv.ParseFloat(c.PostForm("score"), 64)
	comment := c.PostForm("comment")

	var rating database.Rating
	database.DB.Where("recipe_id = ? AND user_id = ?", recipe.ID, user.ID).First(&rating)

	rating.RecipeID = recipe.ID
	rating.UserID = user.ID
	rating.Score = score
	rating.Comment = comment

	database.DB.Save(&rating)

	c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
}

func Favorites(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var recipes []database.Recipe
	query := database.DB.Preload("FavoritedBy").Preload("Tags").
		Joins("JOIN user_favorites ON user_favorites.recipe_id = recipes.id").
		Where("user_favorites.user_id = ?", user.ID)

	if q := c.Query("q"); q != "" {
		query = query.Where("recipes.title LIKE ?", "%"+q+"%")
	}

	sort := c.Query("sort")
	switch sort {
	case "date_asc":
		query = query.Order("recipes.created_at ASC")
	default:
		query = query.Order("recipes.created_at DESC")
	}

	query.Find(&recipes)

	c.HTML(http.StatusOK, "favorites.html", gin.H{
		"recipes": recipes,
		"user":    user,
	})
}

func Community(c *gin.Context) {
	user := middleware.GetUser(c)

	var recipes []database.Recipe
	database.DB.Preload("FavoritedBy").Preload("UploadedBy").Preload("Tags").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ?", "community").
		Order("recipes.created_at DESC").Find(&recipes)

	if c.Request.Method == "POST" {
		recipe := database.Recipe{
			Title:        c.PostForm("title"),
			Description:  c.PostForm("description"),
			Ingredients:  c.PostForm("ingredients"),
			Instructions: c.PostForm("instructions"),
			IsApproved:   false,
		}
		if user != nil {
			recipe.UploadedByID = &user.ID
		}

		if s, err := strconv.Atoi(c.PostForm("servings")); err == nil {
			recipe.Servings = s
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			recipe.Price = &p
		}

		database.DB.Create(&recipe)

		tagStr := c.PostForm("tags_string")
		if tagStr != "" {
			for _, name := range strings.Split(tagStr, ",") {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				var tag database.Tag
				database.DB.Where("name = ?", name).FirstOrCreate(&tag, database.Tag{Name: name})
				database.DB.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", recipe.ID, tag.ID)
			}
		}

		c.Redirect(http.StatusFound, "/profile")
		return
	}

	var uploaders []string
	database.DB.Model(&database.User{}).Distinct("username").Pluck("username", &uploaders)
	var allTags []database.Tag
	database.DB.Find(&allTags)

	c.HTML(http.StatusOK, "community.html", gin.H{
		"recipes":   recipes,
		"user":      user,
		"tags":      allTags,
		"uploaders": uploaders,
	})
}

func OrderTracker(c *gin.Context) {
	token := c.Param("token")
	var order database.Order
	if err := database.DB.Preload("Items.Recipe").Where("tracking_token = ?", token).First(&order).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Order not found"})
		return
	}

	c.HTML(http.StatusOK, "order_tracker.html", gin.H{"order": order})
}

func OrderRecipe(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	order := database.Order{
		UserID: user.ID,
		Status: "PENDING",
	}
	if recipe.Price != nil {
		order.TotalPrice = *recipe.Price
	}
	database.DB.Create(&order)

	database.DB.Create(&database.OrderItem{
		OrderID:  order.ID,
		RecipeID: recipe.ID,
		Quantity: 1,
		Price:    *recipe.Price,
	})

	c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
}

func ViewCart(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var items []database.CartItem
	database.DB.Preload("Recipe").Preload("Recipe.FavoritedBy").Preload("Recipe.Tags").
		Where("user_id = ?", user.ID).Find(&items)

	var total float64
	for _, item := range items {
		if item.Recipe.Price != nil {
			total += *item.Recipe.Price * float64(item.Quantity)
		}
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"cart_items": items,
		"total":      total,
		"user":       user,
	})
}

func AddToCart(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	var item database.CartItem
	if err := database.DB.Where("user_id = ? AND recipe_id = ?", user.ID, recipe.ID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			database.DB.Create(&database.CartItem{UserID: user.ID, RecipeID: recipe.ID, Quantity: 1})
		}
	} else {
		item.Quantity++
		database.DB.Save(&item)
	}

	c.Redirect(http.StatusFound, "/cart")
}

func RemoveFromCart(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&database.CartItem{})
	c.Redirect(http.StatusFound, "/cart")
}

func UpdateCartQuantity(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	qty, _ := strconv.Atoi(c.PostForm("quantity"))

	if qty <= 0 {
		database.DB.Where("id = ? AND user_id = ?", id, user.ID).Delete(&database.CartItem{})
	} else {
		database.DB.Model(&database.CartItem{}).Where("id = ? AND user_id = ?", id, user.ID).Update("quantity", qty)
	}
	c.Redirect(http.StatusFound, "/cart")
}

func Checkout(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var items []database.CartItem
	database.DB.Preload("Recipe").Where("user_id = ?", user.ID).Find(&items)
	if len(items) == 0 {
		c.Redirect(http.StatusFound, "/cart")
		return
	}

	tx := database.DB.Begin()

	order := database.Order{UserID: user.ID, Status: "PENDING"}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.Redirect(http.StatusFound, "/cart")
		return
	}

	var total float64
	for _, item := range items {
		price := 0.0
		if item.Recipe.Price != nil {
			price = *item.Recipe.Price
		}
		total += price * float64(item.Quantity)

		tx.Create(&database.OrderItem{
			OrderID:  order.ID,
			RecipeID: item.RecipeID,
			Quantity: item.Quantity,
			Price:    price,
		})
	}
	tx.Model(&order).Update("total_price", total)
	tx.Where("user_id = ?", user.ID).Delete(&database.CartItem{})
	tx.Commit()

	c.Redirect(http.StatusFound, "/profile")
}
