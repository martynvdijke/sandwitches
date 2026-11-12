package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/notifications"
	"gorm.io/gorm"
)

// notificationPingInterval keeps proxies and browsers from dropping an idle
// SSE connection without sending a full event.
const notificationPingInterval = 25 * time.Second

// notificationsStream is the Server-Sent Events endpoint for web clients.
func notificationsStream(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}

	service := notifications.Default()
	if service == nil || service.Hub() == nil {
		apiError(c, http.StatusServiceUnavailable, "Notifications are not available")
		return
	}

	events, cancel := service.Hub().Register(user.ID)
	defer cancel()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if _, err := fmt.Fprint(c.Writer, ": connected\n\n"); err != nil {
		return
	}
	c.Writer.Flush()

	ticker := time.NewTicker(notificationPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case ev := <-events:
			payload, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", payload); err != nil {
				return
			}
			c.Writer.Flush()
		case <-ticker.C:
			if _, err := fmt.Fprint(c.Writer, "event: ping\ndata: {}\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

type deviceTokenInput struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
}

// registerDeviceToken upserts an FCM token for the authenticated user.
func registerDeviceToken(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	var input deviceTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	if input.Token == "" {
		apiError(c, http.StatusUnprocessableEntity, "token is required")
		return
	}
	platform := input.Platform
	if platform == "" {
		platform = "android"
	}

	now := time.Now()
	var device database.DeviceToken
	err := database.DB.Where("token = ?", input.Token).First(&device).Error
	if err == gorm.ErrRecordNotFound {
		device = database.DeviceToken{
			UserID:     user.ID,
			Token:      input.Token,
			Platform:   platform,
			LastSeenAt: now,
		}
		if err := database.DB.Create(&device).Error; err != nil {
			apiError(c, http.StatusInternalServerError, "Could not register device")
			return
		}
	} else if err != nil {
		apiError(c, http.StatusInternalServerError, "Could not register device")
		return
	} else {
		// The token may have been issued to another account on the same install;
		// reassigning it is the intended behaviour.
		if err := database.DB.Model(&device).Updates(map[string]interface{}{
			"user_id":      user.ID,
			"platform":     platform,
			"last_seen_at": now,
			"updated_at":   now,
		}).Error; err != nil {
			apiError(c, http.StatusInternalServerError, "Could not register device")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"registered": true}})
}

type deviceTokenDeleteInput struct {
	Token string `json:"token" binding:"required"`
}

// unregisterDeviceToken removes a token owned by the authenticated user.
func unregisterDeviceToken(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	var input deviceTokenDeleteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := database.DB.Where("token = ? AND user_id = ?", input.Token, user.ID).
		Delete(&database.DeviceToken{}).Error; err != nil {
		apiError(c, http.StatusInternalServerError, "Could not remove device")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"removed": true}})
}

func notificationPreferences(data gin.H) gin.H {
	return gin.H{"data": data}
}

// getNotificationPreferences returns the four per-user notification switches.
func getNotificationPreferences(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	c.JSON(http.StatusOK, notificationPreferences(gin.H{
		"notify_new_recipes": user.NotifyNewRecipes,
		"notify_new_orders":  user.NotifyNewOrders,
		"notify_web":         user.NotifyWeb,
		"notify_push":        user.NotifyPush,
	}))
}

type notificationPreferencesInput struct {
	NotifyNewRecipes *bool `json:"notify_new_recipes"`
	NotifyNewOrders  *bool `json:"notify_new_orders"`
	NotifyWeb        *bool `json:"notify_web"`
	NotifyPush       *bool `json:"notify_push"`
}

// updateNotificationPreferences patches only the supplied switches.
func updateNotificationPreferences(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		apiError(c, http.StatusForbidden, "Please sign in")
		return
	}
	var input notificationPreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.NotifyNewRecipes != nil {
		updates["notify_new_recipes"] = *input.NotifyNewRecipes
	}
	if input.NotifyNewOrders != nil {
		updates["notify_new_orders"] = *input.NotifyNewOrders
	}
	if input.NotifyWeb != nil {
		updates["notify_web"] = *input.NotifyWeb
	}
	if input.NotifyPush != nil {
		updates["notify_push"] = *input.NotifyPush
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&database.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
			apiError(c, http.StatusInternalServerError, "Could not update preferences")
			return
		}
	}

	var fresh database.User
	if err := database.DB.First(&fresh, user.ID).Error; err != nil {
		apiError(c, http.StatusInternalServerError, "Could not update preferences")
		return
	}
	c.JSON(http.StatusOK, notificationPreferences(gin.H{
		"notify_new_recipes": fresh.NotifyNewRecipes,
		"notify_new_orders":  fresh.NotifyNewOrders,
		"notify_web":         fresh.NotifyWeb,
		"notify_push":        fresh.NotifyPush,
	}))
}
