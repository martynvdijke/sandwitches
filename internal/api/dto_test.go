package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/database"
)

func TestSettingToDTO(t *testing.T) {
	s := &database.Setting{
		SiteName:          "Sandwitches",
		SiteDescription:   "desc",
		Email:             "a@b.c",
		AIConnectionPoint: "http://ai",
		AIModel:           "gpt-4",
		AIAPIKey:          "secret",
		LogLevel:          "INFO",
		GotifyURL:         "http://gotify",
		GotifyToken:       "tok",
		OTelEndpoint:      "http://otel",
		OTelEnabled:       true,
	}
	dto := settingToDTO(s)
	if dto.SiteName != "Sandwitches" || dto.AIConnectionPoint != "http://ai" || !dto.OTelEnabled {
		t.Error("settingToDTO did not map fields")
	}
	if dto.SiteName == "" || dto.AIAPIKey == "" {
		t.Error("settingToDTO lost data")
	}
}

func TestApplySettingUpdates(t *testing.T) {
	s := &database.Setting{SiteName: "Old", LogLevel: "INFO"}
	email := "new@x.y"
	applySettingUpdates(s, SettingUpdateInput{Email: &email})
	if s.Email != "new@x.y" {
		t.Error("applySettingUpdates should set email")
	}
	if s.SiteName != "Old" {
		t.Error("applySettingUpdates must not touch unset fields")
	}
	if s.LogLevel != "INFO" {
		t.Error("applySettingUpdates must not touch unset fields (log_level)")
	}
}

func TestUserToDTO(t *testing.T) {
	u := &database.User{
		ID: 7, Username: "chef", Email: "leak@me", FirstName: "F", LastName: "L",
		Avatar: "a.png", Bio: "bio", Language: "nl", Theme: "dark",
	}
	dto := userToDTO(u)
	if dto.ID != 7 || dto.Username != "chef" || dto.Language != "nl" || dto.Theme != "dark" {
		t.Error("userToDTO did not map public fields")
	}
	b, _ := json.Marshal(dto)
	var raw map[string]interface{}
	json.Unmarshal(b, &raw)
	if _, ok := raw["email"]; ok {
		t.Error("userToDTO must not expose email")
	}
}

func TestUserPublicToDTO(t *testing.T) {
	dto := userPublicToDTO(&database.User{Username: "u", FirstName: "F", LastName: "L", Avatar: "a"})
	if dto.Username != "u" || dto.FirstName != "F" || dto.LastName != "L" || dto.Avatar != "a" {
		t.Error("userPublicToDTO mismatch")
	}
}

func TestTagToDTO(t *testing.T) {
	dto := tagToDTO(&database.Tag{ID: 3, Name: "Spicy", Slug: "spicy"})
	if dto.ID != 3 || dto.Name != "Spicy" || dto.Slug != "spicy" {
		t.Error("tagToDTO mismatch")
	}
}

func TestOrderDTOsHideInternalFields(t *testing.T) {
	o := &database.Order{
		ID: 1, UserID: 99, Status: "PENDING", Completed: true,
		TotalPrice: 12.5, TrackingToken: "tok-123",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		Items: []database.OrderItem{{
			ID: 1, RecipeID: 2, Quantity: 1, Price: 12.5,
			Recipe: database.Recipe{ID: 2, Title: "R", Slug: "r", Servings: 1},
		}},
	}
	dto := orderToDTO(o)
	b, _ := json.Marshal(dto)
	var raw map[string]interface{}
	json.Unmarshal(b, &raw)
	for _, k := range []string{"tracking_token", "user_id", "completed", "UserID", "TrackingToken"} {
		if _, ok := raw[k]; ok {
			t.Errorf("orderToDTO must not expose %s", k)
		}
	}
	if dto.Status != "PENDING" || dto.TotalPrice != 12.5 {
		t.Error("orderToDTO lost public fields")
	}

	detail := orderToDetailDTO(o)
	if len(detail.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Recipe["title"] != "R" {
		t.Error("order item recipe should be full recipe payload")
	}
}

func TestRatingToDTO(t *testing.T) {
	r := &database.Rating{
		ID: 1, RecipeID: 2, Score: 9, Comment: "yum",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		User: database.User{Username: "u", FirstName: "F"},
	}
	dto := ratingToDTO(r)
	if dto.Recipe != 2 || dto.Score != 9 || dto.User.Username != "u" {
		t.Error("ratingToDTO mismatch")
	}
	if dto.CreatedAt.IsZero() {
		t.Error("ratingToDTO must carry created_at")
	}
}
