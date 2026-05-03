package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
	"gorm.io/gorm"
)

func AdminDashboard(c *gin.Context) {
	td := utils.NewTemplateData(c)
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

	var recentOrders []database.Order
	database.DB.Preload("User").Preload("Items.Recipe").Order("created_at DESC").Limit(5).Find(&recentOrders)

	c.HTML(http.StatusOK, "admin/dashboard.html", td.With("recipe_count", recipeCount).
		With("user_count", userCount).
		With("tag_count", tagCount).
		With("recent_recipes", recentRecipes).
		With("pending_recipes", pendingRecipes).
		With("recent_orders", recentOrders).
		ToGinH())
}

func AdminRecipeList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var recipes []database.Recipe
	query := database.DB.Preload("Tags").Preload("UploadedBy")

	sort := c.Query("sort")
	switch sort {
	case "title":
		query = query.Order("title ASC")
	case "title_desc":
		query = query.Order("title DESC")
	case "date_asc":
		query = query.Order("created_at ASC")
	case "price_asc":
		query = query.Order("COALESCE(price, 0) ASC")
	case "price_desc":
		query = query.Order("COALESCE(price, 0) DESC")
	case "approved":
		query = query.Order("is_approved ASC, created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	page, perPage, offset := utils.GetPagination(c, 20)
	var total int64
	query.Session(&gorm.Session{}).Count(&total)
	pagination := utils.NewPagination(page, perPage, total)

	query.Limit(perPage).Offset(offset).Find(&recipes)

	c.HTML(http.StatusOK, "admin/recipe_list.html", td.With("recipes", recipes).
		With("pagination", pagination).
		ToGinH())
}

func AdminRecipeAdd(c *gin.Context) {
	td := utils.NewTemplateData(c)

	if c.Request.Method == "POST" {
		recipe := database.Recipe{
			Title:        c.PostForm("title"),
			Description:  c.PostForm("description"),
			Ingredients:  c.PostForm("ingredients"),
			Instructions: c.PostForm("instructions"),
			IsApproved:   c.PostForm("is_approved") != "",
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
		if pt, err := strconv.Atoi(c.PostForm("prep_time")); err == nil {
			recipe.PrepTime = &pt
		}
		if ct, err := strconv.Atoi(c.PostForm("cook_time")); err == nil {
			recipe.CookTime = &ct
		}
		if cal, err := strconv.Atoi(c.PostForm("calories")); err == nil {
			recipe.Calories = &cal
		}
		if d, err := strconv.Atoi(c.PostForm("max_daily_orders")); err == nil {
			recipe.MaxDailyOrders = &d
		}

		file, err := c.FormFile("image")
		if err == nil {
			imgPath, err := utils.SaveUploadedFile(file, filepath.Join(mediaRoot, "recipes"), "recipes")
			if err == nil {
				recipe.Image = imgPath
			}
		}

		recipe.Slug = database.UniqueSlug(recipe.Title, nil)
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

		utils.AddFlash(c, "success", "Recipe created")
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	c.HTML(http.StatusOK, "admin/recipe_form.html", td.With("title", "Add Recipe").
		With("recipe", nil).
		ToGinH())
}

func AdminRecipeEdit(c *gin.Context) {
	td := utils.NewTemplateData(c)
	user := middleware.GetUser(c)

	id, _ := strconv.Atoi(c.Param("id"))
	var recipe database.Recipe
	if err := database.DB.Preload("Tags").First(&recipe, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	if c.Request.Method == "POST" {
		changes := map[string]string{}

		if v := c.PostForm("title"); v != recipe.Title {
			changes["Title"] = recipe.Title
			recipe.Title = v
			recipe.Slug = database.UniqueSlug(v, &recipe.ID)
		}
		if v := c.PostForm("description"); v != recipe.Description {
			changes["Description"] = recipe.Description
			recipe.Description = v
		}
		if v := c.PostForm("ingredients"); v != recipe.Ingredients {
			changes["Ingredients"] = recipe.Ingredients
			recipe.Ingredients = v
		}
		if v := c.PostForm("instructions"); v != recipe.Instructions {
			changes["Instructions"] = recipe.Instructions
			recipe.Instructions = v
		}
		recipe.IsHighlighted = c.PostForm("is_highlighted") != ""
		recipe.IsApproved = c.PostForm("is_approved") != ""

		if s, err := strconv.Atoi(c.PostForm("servings")); err == nil {
			recipe.Servings = s
		}
		if p, err := strconv.ParseFloat(c.PostForm("price"), 64); err == nil {
			recipe.Price = &p
		}
		if pt, err := strconv.Atoi(c.PostForm("prep_time")); err == nil {
			recipe.PrepTime = &pt
		}
		if ct, err := strconv.Atoi(c.PostForm("cook_time")); err == nil {
			recipe.CookTime = &ct
		}
		if cal, err := strconv.Atoi(c.PostForm("calories")); err == nil {
			recipe.Calories = &cal
		}
		if d, err := strconv.Atoi(c.PostForm("max_daily_orders")); err == nil {
			recipe.MaxDailyOrders = &d
		}

		file, err := c.FormFile("image")
		if err == nil {
			imgPath, err := utils.SaveUploadedFile(file, filepath.Join(mediaRoot, "recipes"), "recipes")
			if err == nil {
				recipe.Image = imgPath
			}
		}

		database.DB.Save(&recipe)

		for field, oldVal := range changes {
			newVal := ""
			switch field {
			case "Title":
				newVal = recipe.Title
			case "Description":
				newVal = recipe.Description
			case "Ingredients":
				newVal = recipe.Ingredients
			case "Instructions":
				newVal = recipe.Instructions
			}
			database.RecordRecipeHistory(database.DB, recipe.ID, &user.ID, field, oldVal, newVal)
		}

		tagStr := c.PostForm("tags_string")
		if tagStr != "" {
			database.DB.Exec("DELETE FROM recipe_tags WHERE recipe_id = ?", recipe.ID)
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

		utils.AddFlash(c, "success", "Recipe updated")
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	var tagNames []string
	for _, t := range recipe.Tags {
		tagNames = append(tagNames, t.Name)
	}

	c.HTML(http.StatusOK, "admin/recipe_form.html", td.With("title", "Edit Recipe").
		With("recipe", recipe).
		With("tag_names", strings.Join(tagNames, ", ")).
		ToGinH())
}

func AdminRecipeApprove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Model(&database.Recipe{}).Where("id = ?", id).Update("is_approved", true)
	utils.AddFlash(c, "success", "Recipe approved")
	c.Redirect(http.StatusFound, "/dashboard/approvals")
}

func AdminRecipeDelete(c *gin.Context) {
	td := utils.NewTemplateData(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Recipe{}, id)
		utils.AddFlash(c, "success", "Recipe deleted")
		c.Redirect(http.StatusFound, "/dashboard/recipes")
		return
	}

	var recipe database.Recipe
	database.DB.First(&recipe, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", td.With("object", recipe).
		With("type", "recipe").
		ToGinH())
}

func AdminUserList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var users []database.User
	database.DB.Order("date_joined DESC").Find(&users)
	c.HTML(http.StatusOK, "admin/user_list.html", td.With("users", users).ToGinH())
}

func AdminUserEdit(c *gin.Context) {
	td := utils.NewTemplateData(c)
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
		utils.AddFlash(c, "success", "User updated")
		c.Redirect(http.StatusFound, "/dashboard/users")
		return
	}

	c.HTML(http.StatusOK, "admin/user_form.html", td.With("user_obj", user).ToGinH())
}

func AdminUserDelete(c *gin.Context) {
	td := utils.NewTemplateData(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.User{}, id)
		utils.AddFlash(c, "success", "User deleted")
		c.Redirect(http.StatusFound, "/dashboard/users")
		return
	}

	var user database.User
	database.DB.First(&user, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", td.With("object", user).
		With("type", "user").
		ToGinH())
}

func AdminTagList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var tags []database.Tag
	database.DB.Order("name ASC").Find(&tags)
	c.HTML(http.StatusOK, "admin/tag_list.html", td.With("tags", tags).ToGinH())
}

func AdminTagAdd(c *gin.Context) {
	td := utils.NewTemplateData(c)
	if c.Request.Method == "POST" {
		database.DB.Create(&database.Tag{Name: c.PostForm("name")})
		utils.AddFlash(c, "success", "Tag created")
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}
	c.HTML(http.StatusOK, "admin/tag_form.html", td.With("title", "Add Tag").ToGinH())
}

func AdminTagEdit(c *gin.Context) {
	td := utils.NewTemplateData(c)
	id, _ := strconv.Atoi(c.Param("id"))
	var tag database.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	if c.Request.Method == "POST" {
		tag.Name = c.PostForm("name")
		database.DB.Save(&tag)
		utils.AddFlash(c, "success", "Tag updated")
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	c.HTML(http.StatusOK, "admin/tag_form.html", td.With("tag", tag).
		With("title", "Edit Tag").
		ToGinH())
}

func AdminTagDelete(c *gin.Context) {
	td := utils.NewTemplateData(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Tag{}, id)
		utils.AddFlash(c, "success", "Tag deleted")
		c.Redirect(http.StatusFound, "/dashboard/tags")
		return
	}

	var tag database.Tag
	database.DB.First(&tag, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", td.With("object", tag).
		With("type", "tag").
		ToGinH())
}

func AdminOrderList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var orders []database.Order
	query := database.DB.Preload("User").Preload("Items.Recipe")

	sort := c.Query("sort")
	switch sort {
	case "date_asc":
		query = query.Order("created_at ASC")
	case "status":
		query = query.Order("status ASC, created_at DESC")
	case "price_asc":
		query = query.Order("total_price ASC")
	case "price_desc":
		query = query.Order("total_price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	page, perPage, offset := utils.GetPagination(c, 20)
	var total int64
	query.Session(&gorm.Session{}).Count(&total)
	pagination := utils.NewPagination(page, perPage, total)

	query.Limit(perPage).Offset(offset).Find(&orders)

	c.HTML(http.StatusOK, "admin/order_list.html", td.With("orders", orders).
		With("status_choices", OrderStatusChoices).
		With("pagination", pagination).
		ToGinH())
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

	oldStatus := order.Status
	order.Status = newStatus
	if newStatus == "COMPLETED" {
		order.Completed = true
	}
	database.DB.Save(&order)

	utils.AddFlash(c, "success", fmt.Sprintf("Order #%d status updated: %s → %s", order.ID, oldStatus, newStatus))
	c.Redirect(http.StatusFound, "/dashboard/orders")
}

func AdminSettings(c *gin.Context) {
	td := utils.NewTemplateData(c)
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
		utils.AddFlash(c, "success", "Settings saved")
		c.Redirect(http.StatusFound, "/dashboard/settings")
		return
	}

	c.HTML(http.StatusOK, "admin/settings.html", td.With("setting", setting).ToGinH())
}

func AdminRatingList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var ratings []database.Rating
	query := database.DB.Preload("Recipe").Preload("User")

	page, perPage, offset := utils.GetPagination(c, 20)
	var total int64
	query.Session(&gorm.Session{}).Count(&total)
	pagination := utils.NewPagination(page, perPage, total)

	query.Order("updated_at DESC").Limit(perPage).Offset(offset).Find(&ratings)

	c.HTML(http.StatusOK, "admin/rating_list.html", td.With("ratings", ratings).
		With("pagination", pagination).
		ToGinH())
}

func AdminRatingDelete(c *gin.Context) {
	td := utils.NewTemplateData(c)
	id, _ := strconv.Atoi(c.Param("id"))
	if c.Request.Method == "POST" {
		database.DB.Delete(&database.Rating{}, id)
		utils.AddFlash(c, "success", "Rating deleted")
		c.Redirect(http.StatusFound, "/dashboard/ratings")
		return
	}

	var rating database.Rating
	database.DB.Preload("Recipe").Preload("User").First(&rating, id)
	c.HTML(http.StatusOK, "admin/confirm_delete.html", td.With("object", rating).
		With("type", "rating").
		ToGinH())
}

func AdminRecipeApprovalList(c *gin.Context) {
	td := utils.NewTemplateData(c)
	var recipes []database.Recipe
	database.DB.Preload("UploadedBy").
		Joins("JOIN user_groups ON recipes.uploaded_by_id = user_groups.user_id").
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("groups.name = ? AND recipes.is_approved = ?", "community", false).
		Order("recipes.created_at DESC").Find(&recipes)

	c.HTML(http.StatusOK, "admin/recipe_approval_list.html", td.With("recipes", recipes).ToGinH())
}

var mediaRoot string

func SetMediaRoot(root string) {
	mediaRoot = root
}

func AdminLogs(c *gin.Context) {
	td := utils.NewTemplateData(c)
	logFile := "sandwitches.log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		c.HTML(http.StatusOK, "admin/logs.html", td.With("logs", "No log file found").ToGinH())
		return
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		c.HTML(http.StatusOK, "admin/logs.html", td.With("logs", "Error reading log file").ToGinH())
		return
	}

	lines := strings.Split(string(data), "\n")
	tail := c.Query("tail")
	n := 100
	if tail != "" {
		if v, err := strconv.Atoi(tail); err == nil && v > 0 {
			n = v
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	c.HTML(http.StatusOK, "admin/logs.html", td.With("logs", strings.Join(lines, "\n")).ToGinH())
}

func AdminTasks(c *gin.Context) {
	td := utils.NewTemplateData(c)
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		c.HTML(http.StatusOK, "admin/tasks.html", td.With("processes", "Error listing processes").ToGinH())
		return
	}
	c.HTML(http.StatusOK, "admin/tasks.html", td.With("processes", strings.TrimSpace(string(output))).ToGinH())
}

func init() {
	_ = io.Discard
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
