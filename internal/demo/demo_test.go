package demo_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/demo"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func demoConfig() *config.Config {
	return &config.Config{
		Debug:            true,
		SecretKey:        "demo-test",
		LanguageCode:     "en",
		DemoUserEnabled:  true,
		DemoUserUsername: "demo",
		DemoUserPassword: "demo",
		DemoUserEmail:    "demo@sandwitches.local",
	}
}

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	tmp := t.TempDir()
	cfg := &config.Config{
		Debug:        true,
		SecretKey:    "demo-test",
		DatabaseFile: filepath.Join(tmp, "test.db"),
		LanguageCode: "en",
	}
	database.Init(cfg)
	t.Cleanup(func() {
		if sqlDB, err := database.DB.DB(); err == nil {
			sqlDB.Close()
		}
		_ = os.Remove(cfg.DatabaseFile)
	})
	return database.DB
}

func TestEnsureDisabledIsNoop(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()
	cfg.DemoUserEnabled = false

	if err := demo.Ensure(context.Background(), db, cfg); err != nil {
		t.Fatalf("disabled Ensure returned error: %v", err)
	}

	var count int64
	db.Model(&database.User{}).Count(&count)
	if count != 0 {
		t.Fatalf("disabled demo created %d users", count)
	}
	db.Model(&database.Recipe{}).Count(&count)
	if count != 0 {
		t.Fatalf("disabled demo seeded %d recipes", count)
	}
}

func TestEnsureCreatesDemoUserAndSeeds(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()

	if err := demo.Ensure(context.Background(), db, cfg); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	var user database.User
	if err := db.Where("username = ?", "demo").First(&user).Error; err != nil {
		t.Fatalf("demo user not created: %v", err)
	}
	if !user.IsDemo {
		t.Error("IsDemo = false")
	}
	if !user.IsActive {
		t.Error("demo user is not active")
	}
	if user.IsSuperuser || user.IsStaff {
		t.Error("demo user must not have admin rights")
	}
	if user.AcceptedTermsAt == nil {
		t.Error("demo user should have accepted the terms")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("demo")) != nil {
		t.Error("demo password does not verify")
	}

	var memberships int64
	db.Table("user_groups").Where("user_id = ?", user.ID).Count(&memberships)
	if memberships == 0 {
		t.Error("demo user is not in the community group")
	}

	var recipes []database.Recipe
	if err := db.Preload("Tags").Where("uploaded_by_id = ?", user.ID).Find(&recipes).Error; err != nil {
		t.Fatalf("load seeded recipes: %v", err)
	}
	if len(recipes) != 2 {
		t.Fatalf("expected 2 seeded recipes, got %d", len(recipes))
	}
	for _, r := range recipes {
		if !r.IsApproved {
			t.Errorf("seeded recipe %q is not approved", r.Title)
		}
		if r.Ingredients == "" || r.Instructions == "" {
			t.Errorf("seeded recipe %q is missing text content", r.Title)
		}
		var tags int
		for _, tag := range r.Tags {
			if tag.Name == "demo" || tag.Name == "vegetarian" {
				tags++
			}
		}
		if tags != 2 {
			t.Errorf("seeded recipe %q has tags %v, want demo and vegetarian", r.Title, r.Tags)
		}
	}
}

func TestEnsureIsIdempotent(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()

	for i := 0; i < 2; i++ {
		if err := demo.Ensure(context.Background(), db, cfg); err != nil {
			t.Fatalf("Ensure call %d: %v", i+1, err)
		}
	}

	var users, recipes int64
	db.Model(&database.User{}).Where("username = ?", "demo").Count(&users)
	db.Model(&database.Recipe{}).Count(&recipes)
	if users != 1 {
		t.Errorf("expected 1 demo user after two calls, got %d", users)
	}
	if recipes != 2 {
		t.Errorf("expected 2 seeded recipes after two calls, got %d", recipes)
	}
}

func TestEnsureReenablesInactiveDemo(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()

	if err := demo.Ensure(context.Background(), db, cfg); err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	db.Model(&database.User{}).Where("username = ?", "demo").Update("is_active", false)

	if err := demo.Ensure(context.Background(), db, cfg); err != nil {
		t.Fatalf("Ensure again: %v", err)
	}
	var user database.User
	if err := db.Where("username = ?", "demo").First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if !user.IsActive {
		t.Error("inactive demo user was not re-enabled")
	}
	// The password must still verify after the repair pass.
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("demo")) != nil {
		t.Error("demo password no longer verifies")
	}
}

func TestEnsureRefusesExistingNormalUser(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()

	hash, _ := bcrypt.GenerateFromPassword([]byte("real-password"), bcrypt.DefaultCost)
	existing := database.User{
		Username:   "demo",
		Email:      "real@example.com",
		Password:   string(hash),
		IsActive:   true,
		DateJoined: time.Now(),
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}

	if err := demo.Ensure(context.Background(), db, cfg); err == nil {
		t.Fatal("expected Ensure to refuse an existing normal account")
	}

	var user database.User
	if err := db.Where("username = ?", "demo").First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if user.IsDemo {
		t.Error("Ensure must not mark a normal account as demo")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("real-password")) != nil {
		t.Error("Ensure reset a normal account's password")
	}
}

func TestEnsureRefusesExistingSuperuser(t *testing.T) {
	db := setupDB(t)
	cfg := demoConfig()

	hash, _ := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.DefaultCost)
	admin := database.User{
		Username:    "demo",
		Email:       "admin@example.com",
		Password:    string(hash),
		IsActive:    true,
		IsSuperuser: true,
		IsStaff:     true,
		DateJoined:  time.Now(),
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}

	if err := demo.Ensure(context.Background(), db, cfg); err == nil {
		t.Fatal("expected Ensure to refuse an existing superuser")
	}

	var user database.User
	if err := db.Where("username = ?", "demo").First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if !user.IsSuperuser || !user.IsStaff {
		t.Error("Ensure must not change the admin's role")
	}
	if user.IsDemo {
		t.Error("Ensure must not mark an admin account as demo")
	}
}
