package database

import (
	"os"
	"testing"

	"github.com/martynvdijke/sandwitches-go/internal/config"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	dbPath := tmp + "/test.db"
	cfg := &config.Config{
		DatabaseFile: dbPath,
		Debug:        false,
		SecretKey:    "testkey",
		LanguageCode: "en",
	}
	Init(cfg)
}

func TestInit(t *testing.T) {
	setupTestDB(t)

	var count int64
	DB.Model(&Setting{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 setting, got %d", count)
	}

	var adminGroup, communityGroup Group
	if err := DB.Where("name = ?", "admin").First(&adminGroup).Error; err != nil {
		t.Error("admin group not created")
	}
	if err := DB.Where("name = ?", "community").First(&communityGroup).Error; err != nil {
		t.Error("community group not created")
	}
}

func TestCreateUser(t *testing.T) {
	setupTestDB(t)

	user := User{
		Username: "testuser",
		Password: "hashedpw",
		Email:    "test@test.com",
		Language: "en",
		Theme:    "light",
		IsActive: true,
	}
	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	if user.ID == 0 {
		t.Error("user ID not set after create")
	}
}

func TestCreateRecipe(t *testing.T) {
	setupTestDB(t)

	user := User{Username: "chef", Password: "pw", Language: "en", Theme: "light", IsActive: true}
	DB.Create(&user)

	recipe := Recipe{
		Title:        "Test Recipe",
		Description:  "A test recipe",
		Ingredients:  "1 cup flour\n2 eggs",
		Instructions: "Mix and bake",
		Servings:     2,
		IsApproved:   true,
	}
	recipe.UploadedByID = &user.ID
	if err := DB.Create(&recipe).Error; err != nil {
		t.Fatalf("failed to create recipe: %v", err)
	}

	if recipe.Slug == "" {
		t.Error("slug not auto-generated")
	}

	var count int64
	DB.Model(&Rating{}).Where("recipe_id = ?", recipe.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 ratings, got %d", count)
	}
}

func TestOrderWithTrackingToken(t *testing.T) {
	setupTestDB(t)

	user := User{Username: "buyer", Password: "pw", Language: "en", Theme: "light", IsActive: true}
	DB.Create(&user)

	order := Order{UserID: user.ID, Status: "PENDING"}
	if err := DB.Create(&order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	if order.TrackingToken == "" {
		t.Error("tracking token not auto-generated")
	}

	var found Order
	DB.Where("tracking_token = ?", order.TrackingToken).First(&found)
	if found.ID != order.ID {
		t.Error("order not found by tracking token")
	}
}

func TestCartItemUniqueness(t *testing.T) {
	setupTestDB(t)

	user := User{Username: "cartuser", Password: "pw", Language: "en", Theme: "light", IsActive: true}
	DB.Create(&user)

	recipe := Recipe{Title: "Cart Recipe", IsApproved: true}
	DB.Create(&recipe)

	DB.Create(&CartItem{UserID: user.ID, RecipeID: recipe.ID, Quantity: 1})

	dup := CartItem{UserID: user.ID, RecipeID: recipe.ID, Quantity: 1}
	err := DB.Create(&dup).Error
	if err == nil {
		t.Error("should not allow duplicate cart item for same user and recipe")
	}
}

func TestHashedFilename(t *testing.T) {
	name, err := HashedFilename("test.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name == "" {
		t.Error("empty filename")
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	dbFile := os.Getenv("DATABASE_FILE")
	if dbFile != "" {
		os.Remove(dbFile)
	}
	os.Exit(code)
}
