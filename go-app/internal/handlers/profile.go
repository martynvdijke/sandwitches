package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

func UserProfile(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if c.Request.Method == "POST" {
		user.FirstName = c.PostForm("first_name")
		user.LastName = c.PostForm("last_name")
		user.Email = c.PostForm("email")
		user.Bio = c.PostForm("bio")
		database.DB.Save(user)
		c.Redirect(http.StatusFound, "/profile")
		return
	}

	var orders []database.Order
	status := c.Query("status")
	query := database.DB.Preload("Items.Recipe").Where("user_id = ?", user.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	sort := c.Query("sort")
	switch sort {
	case "date_asc":
		query = query.Order("created_at ASC")
	case "price_asc":
		query = query.Order("total_price ASC")
	case "price_desc":
		query = query.Order("total_price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	query.Find(&orders)

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"user":           user,
		"orders":         orders,
		"current_status": status,
		"current_sort":   c.Query("sort"),
		"status_choices": OrderStatusChoices,
	})
}

func UserSettings(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if c.Request.Method == "POST" {
		user.Language = c.PostForm("language")
		user.Theme = c.PostForm("theme")
		database.DB.Save(user)
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	c.HTML(http.StatusOK, "settings.html", gin.H{
		"user":    user,
		"version": "2.12.15", // TODO: version from build
	})
}

func UserOrderDetail(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	var order database.Order
	if err := database.DB.Preload("Items.Recipe").Where("id = ? AND user_id = ?", id, user.ID).First(&order).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Order not found"})
		return
	}

	c.HTML(http.StatusOK, "order_detail.html", gin.H{
		"order": order,
		"user":  user,
	})
}

var OrderStatusChoices = []struct {
	Code  string
	Label string
}{
	{"PENDING", "Pending"},
	{"PREPARING", "Preparing"},
	{"MADE", "Made"},
	{"SHIPPED", "Shipped"},
	{"COMPLETED", "Completed"},
	{"CANCELLED", "Cancelled"},
}
