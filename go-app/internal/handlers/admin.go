package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func AdminDashboard(c *gin.Context) {
	var recipeCount, userCount, tagCount int64
	database.DB.Model(&database.Recipe{}).Count(&recipeCount)
	database.DB.Model(&database.User{}).Count(&userCount)
	database.DB.Model(&database.Tag{}).Count(&tagCount)

	var recentRecipes []database.Recipe
	database.DB.Order("created_at DESC").Limit(5).Find(&recentRecipes)

	var pendingRecipes []database.Recipe
	database.DB.Preload("UploadedBy").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ? AND recipes.is_approved = ?", "community", false).
		Order("recipes.created_at DESC").Find(&pendingRecipes)

	c.HTML(http.StatusOK, "admin/dashboard.html", gin.H{
		"recipe_count":    recipeCount,
		"user_count":      userCount,
		"tag_count":       tagCount,
		"recent_recipes":  recentRecipes,
		"pending_recipes": pendingRecipes,
	})
}

func AdminRecipeList(c *gin.Context) {
	var recipes []database.Recipe
	database.DB.Preload("Tags").Preload("UploadedBy").
		Order("created_at DESC").Find(&recipes)

	c.HTML(http.StatusOK, "admin/recipe_list.html", gin.H{
		"recipes": recipes,
	})
}

func AdminRecipeAdd(c *gin.Context) {
	if c.Request.Method == "POST" {
		recipe := database.Recipe{
			Title:        c.PostForm("title"),
			Description:  c.PostForm("description"),
			Ingredients:  c.PostForm("ingredients"),
			Instructions: c.PostForm("instructions"),
			IsApproved:   true,
		}

		if s, err := strconv.Atoi(c.PostForm("servings")); err == nil {
			recipe.Servings = s
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			recipe.Price = &p
		}
		if hl := c.PostForm("is_highlighted"); hl != "" {
			recipe.IsHighlighted = true
		}

		database.DB.Create(&recipe)
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	c.HTML(http.StatusOK, "admin/recipe_form.html", gin.H{
		"title": "Add Recipe",
	})
}

func AdminRecipeEdit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.Preload("Tags").First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	if c.Request.Method == "POST" {
		recipe.Title = c.PostForm("title")
		recipe.Description = c.PostForm("description")
		recipe.Ingredients = c.PostForm("ingredients")
		recipe.Instructions = c.PostForm("instructions")
		recipe.IsHighlighted = c.PostForm("is_highlighted") != ""
		recipe.IsApproved = c.PostForm("is_approved") != ""

		if s, err := strconv.Atoi(c.PostForm("servings")); err == nil {
			recipe.Servings = s
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			recipe.Price = &p
		}

		database.DB.Save(&recipe)
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	c.HTML(http.StatusOK, "admin/recipe_form.html", gin.H{
		"title":  "Edit Recipe",
		"recipe": recipe,
	})
}

func AdminRecipeApprove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Model(&database.Recipe{}).Where("id = ?", id).Update("is_approved", true)
	c.Redirect(http.StatusFound, "/dashboard/approvals")
}

func AdminRecipeDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Recipe{}, id)
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	var recipe database.Recipe
	database.DB.First(&recipe, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", gin.H{
		"object": recipe,
		"type":   "recipe",
	})
}

func AdminUserList(c *gin.Context) {
	var users []database.User
	database.DB.Find(&users)
	c.HTML(http.StatusOK, "admin/user_list.html", gin.H{"users": users})
}

func AdminUserEdit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user database.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/users")
		return
	}

	if c.Request.Method == "POST" {
		user.Username = c.PostForm("username")
		user.Email = c.PostForm("email")
		user.FirstName = c.PostForm("first_name")
		user.LastName = c.PostForm("last_name")
		user.Bio = c.PostForm("bio")
		user.IsStaff = c.PostForm("is_staff") != ""
		user.IsActive = c.PostForm("is_active") != ""
		database.DB.Save(&user)
		c.Redirect(http.StatusFound, "/dashboard/users")
		return
	}

	c.HTML(http.StatusOK, "admin/user_form.html", gin.H{
		"user_obj": user,
	})
}

func AdminUserDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.User{}, id)
		c.Redirect(http.StatusFound, "/dashboard/users")
		return
	}

	var user database.User
	database.DB.First(&user, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", gin.H{
		"object": user,
		"type":   "user",
	})
}

func AdminTagList(c *gin.Context) {
	var tags []database.Tag
	database.DB.Find(&tags)
	c.HTML(http.StatusOK, "admin/tag_list.html", gin.H{"tags": tags})
}

func AdminTagAdd(c *gin.Context) {
	if c.Request.Method == "POST" {
		database.DB.Create(&database.Tag{Name: c.PostForm("name")})
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}
	c.HTML(http.StatusOK, "admin/tag_form.html", gin.H{"title": "Add Tag"})
}

func AdminTagEdit(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var tag database.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	if c.Request.Method == "POST" {
		tag.Name = c.PostForm("name")
		database.DB.Save(&tag)
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	c.HTML(http.StatusOK, "admin/tag_form.html", gin.H{"tag": tag, "title": "Edit Tag"})
}

func AdminTagDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Tag{}, id)
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	var tag database.Tag
	database.DB.First(&tag, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", gin.H{
		"object": tag,
		"type":   "tag",
	})
}

func AdminOrderList(c *gin.Context) {
	var orders []database.Order
	database.DB.Preload("User").Preload("Items.Recipe").Order("created_at DESC").Find(&orders)

	c.HTML(http.StatusOK, "admin/order_list.html", gin.H{
		"orders":         orders,
		"status_choices": OrderStatusChoices,
	})
}

func AdminOrderUpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var order database.Order
	if err := database.DB.First(&order, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/orders")
		return
	}

	if order.Status == "COMPLETED" || order.Status == "CANCELLED" {
		c.Redirect(http.StatusFound, "/dashboard/orders")
		return
	}

	newStatus := c.PostForm("status")
	valid := false
	for _, s := range OrderStatusChoices {
		if s.Code == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		c.Redirect(http.StatusFound, "/dashboard/orders")
		return
	}

	order.Status = newStatus
	if newStatus == "COMPLETED" {
		order.Completed = true
	}
	database.DB.Save(&order)

	c.Redirect(http.StatusFound, "/dashboard/orders")
}

func AdminSettings(c *gin.Context) {
	var setting database.Setting
	database.DB.First(&setting)

	if c.Request.Method == "POST" {
		setting.SiteName = c.PostForm("site_name")
		setting.SiteDescription = c.PostForm("site_description")
		setting.Email = c.PostForm("email")
		setting.LogLevel = c.PostForm("log_level")
		setting.GotifyURL = c.PostForm("gotify_url")
		setting.GotifyToken = c.PostForm("gotify_token")
		database.DB.Save(&setting)
		c.Redirect(http.StatusFound, "/dashboard/settings")
		return
	}

	c.HTML(http.StatusOK, "admin/settings.html", gin.H{
		"setting": setting,
	})
}

func AdminRatingList(c *gin.Context) {
	var ratings []database.Rating
	database.DB.Preload("Recipe").Preload("User").Order("updated_at DESC").Find(&ratings)
	c.HTML(http.StatusOK, "admin/rating_list.html", gin.H{"ratings": ratings})
}

func AdminRatingDelete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Rating{}, id)
		c.Redirect(http.StatusFound, "/dashboard/ratings")
		return
	}

	var rating database.Rating
	database.DB.Preload("Recipe").Preload("User").First(&rating, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", gin.H{
		"object": rating,
		"type":   "rating",
	})
}

func AdminRecipeApprovalList(c *gin.Context) {
	var recipes []database.Recipe
	database.DB.Preload("UploadedBy").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ? AND recipes.is_approved = ?", "community", false).
		Order("recipes.created_at DESC").Find(&recipes)

	c.HTML(http.StatusOK, "admin/recipe_approval_list.html", gin.H{"recipes": recipes})
}
