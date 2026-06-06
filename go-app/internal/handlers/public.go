package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/tasks"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
	"gorm.io/gorm"
)

func Index(c *gin.Context) {
	td := utils.NewTemplateData(c)
	user := middleware.GetUser(c)

	var superuserCount int64
	database.DB.Model(&database.User{}).Where("is_superuser = ?", true).Count(&superuserCount)
	if superuserCount == 0 {
		c.Redirect(http.StatusFound, "/setup")
		return
	}

	var recipes []database.Recipe
	query := database.DB.Preload("FavoritedBy").Preload("Tags").Preload("UploadedBy").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ?", "admin").
		Where("recipes.is_approved = ?", true)

	if q := c.Query("q"); q != "" {
		query = query.Where("recipes.title LIKE ? OR recipes.id IN (SELECT recipe_id FROM recipe_tags JOIN tags ON recipe_tags.tag_id = tags.id WHERE tags.name LIKE ?)", "%"+q+"%", "%"+q+"%")
	}
	if uploader := c.Query("uploader"); uploader != "" {
		query = query.Where("uploaded_by_id IN (SELECT id FROM users WHERE username = ?)", uploader)
	}
	if tags := c.QueryArray("tag"); len(tags) > 0 {
		query = query.Where("recipes.id IN (SELECT recipe_id FROM recipe_tags JOIN tags ON recipe_tags.tag_id = tags.id WHERE tags.name IN ?)", tags)
	}
	if ds := c.Query("date_start"); ds != "" {
		if t, err := time.Parse("2006-01-02", ds); err == nil {
			query = query.Where("recipes.created_at >= ?", t)
		}
	}
	if de := c.Query("date_end"); de != "" {
		if t, err := time.Parse("2006-01-02", de); err == nil {
			query = query.Where("recipes.created_at <= ?", t.Add(24*time.Hour))
		}
	}
	if user != nil && c.Query("favorites") == "on" {
		query = query.Where("recipes.id IN (SELECT recipe_id FROM user_favorites WHERE user_id = ?)", user.ID)
	}

	sort := c.Query("sort")
	switch sort {
	case "date_asc":
		query = query.Order("recipes.created_at ASC")
	case "rating":
		query = query.Order("(SELECT COALESCE(AVG(score), 0) FROM ratings WHERE ratings.recipe_id = recipes.id) DESC, recipes.created_at DESC")
	case "user":
		query = query.Order("(SELECT username FROM users WHERE users.id = recipes.uploaded_by_id) ASC, recipes.created_at DESC")
	default:
		query = query.Order("recipes.created_at DESC")
	}

	page, perPage, offset := utils.GetPagination(c, 12)
	var total int64
	query.Session(&gorm.Session{}).Count(&total)
	pagination := utils.NewPagination(page, perPage, total)

	query.Limit(perPage).Offset(offset).Find(&recipes)

	var highlighted []database.Recipe
	database.DB.Preload("FavoritedBy").Where("is_highlighted = ?", true).Find(&highlighted)

	var uploaders []string
	database.DB.Model(&database.User{}).
		Joins("JOIN recipes ON recipes.uploaded_by_id = users.id").
		Distinct("username").Pluck("username", &uploaders)

	var allTags []database.Tag
	database.DB.Find(&allTags)

	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusOK, "partials/recipe_list.html", gin.H{
			"recipes":    recipes,
			"user":       user,
			"pagination": pagination,
			"q":          c.Query("q"),
			"sort":       sort,
			"date_start": c.Query("date_start"),
			"date_end":   c.Query("date_end"),
			"uploader":   c.Query("uploader"),
			"favorites":  c.Query("favorites"),
		})
		return
	}

	c.HTML(http.StatusOK, "index.html", td.With("recipes", recipes).
		With("highlighted_recipes", highlighted).
		With("uploaders", uploaders).
		With("tags", allTags).
		With("selected_tags", c.QueryArray("tag")).
		With("pagination", pagination).
		With("q", c.Query("q")).
		With("sort", sort).
		With("date_start", c.Query("date_start")).
		With("date_end", c.Query("date_end")).
		With("uploader", c.Query("uploader")).
		With("favorites", c.Query("favorites")).
		ToGinH())
}

func RecipeDetail(c *gin.Context) {
	slug := c.Param("slug")
	user := middleware.GetUser(c)
	td := utils.NewTemplateData(c)

	var recipe database.Recipe
	if err := database.DB.Preload("Tags").Preload("FavoritedBy").Preload("UploadedBy").
		Preload("Ratings.User").Where("slug = ?", slug).First(&recipe).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Recipe not found"})
		return
	}

	if recipe.UploadedBy != nil {
		var communityGroup database.Group
		if err := database.DB.Where("name = ?", "community").First(&communityGroup).Error; err == nil {
			var count int64
			database.DB.Table("user_groups").Where("user_id = ? AND group_id = ?", recipe.UploadedBy.ID, communityGroup.ID).Count(&count)
			isCommunity := count > 0

			if isCommunity && !recipe.IsApproved {
				isOwner := user != nil && recipe.UploadedByID != nil && *recipe.UploadedByID == user.ID
				isStaff := user != nil && user.IsStaff
				if !isOwner && !isStaff {
					c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Recipe not found"})
					return
				}
			}
		}
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

	c.HTML(http.StatusOK, "detail.html", td.With("recipe", recipe).
		With("avg_rating", math.Round(avgRating*10)/10).
		With("rating_count", ratingCount).
		With("user_rating", userRating).
		With("all_ratings", recipe.Ratings).
		ToGinH())
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
	if score < 0 || score > 10 {
		utils.AddFlash(c, "error", "Rating must be between 0 and 10")
		c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
		return
	}
	comment := c.PostForm("comment")

	var rating database.Rating
	database.DB.Where("recipe_id = ? AND user_id = ?", recipe.ID, user.ID).First(&rating)

	rating.RecipeID = recipe.ID
	rating.UserID = user.ID
	rating.Score = score
	rating.Comment = comment

	database.DB.Save(&rating)

	utils.AddFlash(c, "success", "Rating saved")
	c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
}

func Favorites(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	td := utils.NewTemplateData(c)

	var recipes []database.Recipe
	query := database.DB.Preload("FavoritedBy").Preload("Tags").Preload("UploadedBy").
		Joins("JOIN user_favorites ON user_favorites.recipe_id = recipes.id").
		Where("user_favorites.user_id = ?", user.ID)

	if q := c.Query("q"); q != "" {
		query = query.Where("recipes.title LIKE ?", "%"+q+"%")
	}
	if ds := c.Query("date_start"); ds != "" {
		if t, err := time.Parse("2006-01-02", ds); err == nil {
			query = query.Where("recipes.created_at >= ?", t)
		}
	}
	if de := c.Query("date_end"); de != "" {
		if t, err := time.Parse("2006-01-02", de); err == nil {
			query = query.Where("recipes.created_at <= ?", t.Add(24*time.Hour))
		}
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
	case "user":
		query = query.Order("(SELECT username FROM users WHERE users.id = recipes.uploaded_by_id) ASC, recipes.created_at DESC")
	default:
		query = query.Order("recipes.created_at DESC")
	}

	page, perPage, offset := utils.GetPagination(c, 12)
	var total int64
	query.Session(&gorm.Session{}).Count(&total)
	pagination := utils.NewPagination(page, perPage, total)

	query.Limit(perPage).Offset(offset).Find(&recipes)

	var uploaders []string
	database.DB.Model(&database.User{}).
		Joins("JOIN recipes ON recipes.uploaded_by_id = users.id").
		Joins("JOIN user_favorites ON user_favorites.recipe_id = recipes.id").
		Where("user_favorites.user_id = ?", user.ID).
		Distinct("username").Pluck("username", &uploaders)

	var allTags []database.Tag
	database.DB.Table("tags").
		Joins("JOIN recipe_tags ON recipe_tags.tag_id = tags.id").
		Joins("JOIN user_favorites ON user_favorites.recipe_id = recipe_tags.recipe_id").
		Where("user_favorites.user_id = ?", user.ID).
		Group("tags.id").
		Find(&allTags)

	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusOK, "partials/recipe_list.html", gin.H{
			"recipes":    recipes,
			"user":       user,
			"pagination": pagination,
		})
		return
	}

	c.HTML(http.StatusOK, "favorites.html", td.With("recipes", recipes).
		With("pagination", pagination).
		With("uploaders", uploaders).
		With("tags", allTags).
		With("selected_tags", c.QueryArray("tag")).
		With("q", c.Query("q")).
		With("sort", sort).
		With("date_start", c.Query("date_start")).
		With("date_end", c.Query("date_end")).
		With("uploader", c.Query("uploader")).
		ToGinH())
}

func Community(c *gin.Context) {
	user := middleware.GetUser(c)
	td := utils.NewTemplateData(c)

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

		if recipe.Slug == "" {
			recipe.Slug = database.Slugify(recipe.Title)
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

		utils.AddFlash(c, "success", "Recipe submitted for approval")
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	var recipes []database.Recipe
	query := database.DB.Preload("FavoritedBy").Preload("UploadedBy").Preload("Tags").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ?", "community")
	if user == nil {
		query = query.Where("recipes.is_approved = ?", true)
	} else if user.IsStaff {
		// Staff see all community recipes
	} else {
		query = query.Where("(recipes.is_approved = ? OR recipes.uploaded_by_id = ?)", true, user.ID)
	}
	query.Order("recipes.created_at DESC").Find(&recipes)

	var uploaders []string
	database.DB.Model(&database.User{}).Distinct("username").Pluck("username", &uploaders)
	var allTags []database.Tag
	database.DB.Find(&allTags)

	c.HTML(http.StatusOK, "community.html", td.With("recipes", recipes).
		With("tags", allTags).
		With("uploaders", uploaders).
		ToGinH())
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

func ViewCart(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	td := utils.NewTemplateData(c)

	var items []database.CartItem
	database.DB.Preload("Recipe").Preload("Recipe.FavoritedBy").Preload("Recipe.Tags").
		Where("user_id = ?", user.ID).Find(&items)

	var total float64
	for _, item := range items {
		if item.Recipe.Price != nil {
			total += *item.Recipe.Price * float64(item.Quantity)
		}
	}

	c.HTML(http.StatusOK, "cart.html", td.With("cart_items", items).
		With("total", total).
		ToGinH())
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

	if recipe.Price == nil {
		utils.AddFlash(c, "error", "This recipe cannot be ordered because it has no price.")
		c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
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

	if recipe.Price == nil {
		utils.AddFlash(c, "error", "This recipe cannot be ordered because it has no price.")
		c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
		return
	}

	tx := database.DB.Begin()
	order := database.Order{UserID: user.ID, Status: "PENDING"}
	if recipe.Price != nil {
		order.TotalPrice = *recipe.Price
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		utils.AddFlash(c, "error", "Could not create order")
		c.Redirect(http.StatusFound, "/recipes/"+recipe.Slug)
		return
	}

	tx.Create(&database.OrderItem{
		OrderID:  order.ID,
		RecipeID: recipe.ID,
		Quantity: 1,
		Price:    *recipe.Price,
	})

	tx.Model(&database.Recipe{}).Where("id = ?", recipe.ID).
		UpdateColumn("daily_orders_count", gorm.Expr("daily_orders_count + 1"))
	tx.Commit()

	tasks.EnqueueNotifyOrder(order.ID)
	utils.AddFlash(c, "success", "Order placed successfully!")
	c.Redirect(http.StatusFound, "/profile")
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

	for _, item := range items {
		if item.Recipe.MaxDailyOrders != nil && *item.Recipe.MaxDailyOrders > 0 {
			if item.Recipe.DailyOrdersCount+item.Quantity > *item.Recipe.MaxDailyOrders {
				utils.AddFlash(c, "error", item.Recipe.Title+" has reached its daily order limit")
				c.Redirect(http.StatusFound, "/cart")
				return
			}
		}
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

		tx.Model(&database.Recipe{}).Where("id = ?", item.RecipeID).
			UpdateColumn("daily_orders_count", gorm.Expr("daily_orders_count + ?", item.Quantity))
	}
	tx.Model(&order).Update("total_price", total)
	tx.Where("user_id = ?", user.ID).Delete(&database.CartItem{})
	tx.Commit()

	tasks.EnqueueNotifyOrder(order.ID)
	utils.AddFlash(c, "success", "Order placed successfully!")
	c.Redirect(http.StatusFound, "/profile")
}
