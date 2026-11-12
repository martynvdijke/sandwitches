package notifications

import (
	"log"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/database"
	"gorm.io/gorm"
)

// Service decides who receives a notification and delivers it over the web
// hub and (optionally) Firebase Cloud Messaging.
type Service struct {
	db  *gorm.DB
	hub *Hub
}

var defaultService *Service

// Init wires the process-wide notification service. Call it after
// database.Init and before routes are registered.
func Init(db *gorm.DB) *Service {
	defaultService = NewService(db, NewHub())
	return defaultService
}

// Default returns the process-wide service, or nil when Init has not run.
// Callers must guard for nil; a nil service silently disables notifications.
func Default() *Service {
	return defaultService
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	if hub == nil {
		hub = NewHub()
	}
	return &Service{db: db, hub: hub}
}

// Hub exposes the web event hub to the SSE handler.
func (s *Service) Hub() *Hub {
	if s == nil {
		return nil
	}
	return s.hub
}

// NotifyNewRecipe tells every active user (except the author) who opted in
// that a recipe was published. Best-effort: it never returns an error.
func (s *Service) NotifyNewRecipe(actorID uint, recipeID uint, title string, urlPath string) {
	if s == nil || s.db == nil {
		return
	}
	setting, ok := s.loadSetting()
	if !ok {
		return
	}

	var userIDs []uint
	s.db.Model(&database.User{}).
		Where("is_active = ? AND id <> ? AND notify_new_recipes = ?", true, actorID, true).
		Pluck("id", &userIDs)
	if len(userIDs) == 0 {
		return
	}

	ev := Event{
		Type:  "recipe.new",
		Title: "New recipe",
		Body:  title,
		URL:   urlPath,
		Data:  map[string]string{"recipe_id": uintToString(recipeID)},
		Time:  time.Now(),
	}
	s.hub.Publish(userIDs, ev)
	s.push(setting, userIDs, ev)
}

// NotifyNewOrder tells staff members who opted in that an order was placed.
func (s *Service) NotifyNewOrder(orderID uint, summary string, urlPath string) {
	if s == nil || s.db == nil {
		return
	}
	setting, ok := s.loadSetting()
	if !ok {
		return
	}

	var userIDs []uint
	s.db.Model(&database.User{}).
		Where("is_staff = ? AND is_active = ? AND notify_new_orders = ?", true, true, true).
		Pluck("id", &userIDs)
	if len(userIDs) == 0 {
		return
	}

	ev := Event{
		Type:  "order.new",
		Title: "New order",
		Body:  summary,
		URL:   urlPath,
		Data:  map[string]string{"order_id": uintToString(orderID)},
		Time:  time.Now(),
	}
	s.hub.Publish(userIDs, ev)
	s.push(setting, userIDs, ev)
}

// loadSetting reads the singleton setting and reports whether notification
// delivery is enabled at all.
func (s *Service) loadSetting() (database.Setting, bool) {
	var setting database.Setting
	if err := s.db.First(&setting).Error; err != nil {
		return setting, false
	}
	if !setting.NotificationsEnabled {
		return setting, false
	}
	return setting, true
}

// push sends an FCM notification to every push-enabled target user, in the
// background so a slow Firebase call never delays the HTTP response.
func (s *Service) push(setting database.Setting, userIDs []uint, ev Event) {
	if !setting.FCMEnabled || setting.FCMProjectID == "" || setting.FCMServiceAccountJSON == "" {
		return
	}

	var tokens []string
	if err := s.db.Table("device_tokens").
		Joins("JOIN users ON users.id = device_tokens.user_id").
		Where("device_tokens.user_id IN ? AND device_tokens.platform = ? AND users.notify_push = ? AND users.is_active = ?",
			userIDs, "android", true, true).
		Pluck("device_tokens.token", &tokens).Error; err != nil {
		log.Printf("notifications: could not load device tokens: %v", err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	projectID := setting.FCMProjectID
	serviceAccount := setting.FCMServiceAccountJSON
	db := s.db
	go func() {
		if err := sendFCM(projectID, serviceAccount, tokens, ev, func(token string) {
			if err := db.Where("token = ?", token).Delete(&database.DeviceToken{}).Error; err != nil {
				log.Printf("notifications: could not remove stale device token: %v", err)
			}
		}); err != nil {
			// Only the aggregate failure is logged: tokens and credentials must
			// never reach the logs.
			log.Printf("notifications: FCM delivery failed for %d device(s): %v", len(tokens), err)
		}
	}()
}

func uintToString(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
