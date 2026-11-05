// Package demo provisions the shared demo account used by the hosted
// Sandwitches instance. The account is deliberately public: its username and
// password are advertised on the login screen so an app-store reviewer or a
// visitor can try the app without registering.
package demo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	demoBio = "Shared demo account for app-store review. Feel free to look around."
)

type seedRecipe struct {
	Title        string
	Description  string
	Ingredients  string
	Instructions string
	Servings     int
}

// The seeds are intentionally small and text-only so a fresh install has
// public content without needing uploaded images.
var seedRecipes = []seedRecipe{
	{
		Title:        "Demo Cheese Sandwich",
		Description:  "A simple toasted cheese sandwich to try the app.",
		Ingredients:  "2 slices of bread\n2 slices of cheese\n1 tablespoon butter",
		Instructions: "Butter one side of each slice of bread.\nPut the cheese between the slices, buttered sides out.\nToast in a pan over medium heat until golden on both sides.",
		Servings:     1,
	},
	{
		Title:        "Demo Tomato Sandwich",
		Description:  "A fresh vegetarian sandwich with tomato and herbs.",
		Ingredients:  "2 slices of bread\n1 ripe tomato, sliced\nA few basil leaves\nSalt and pepper\nOlive oil",
		Instructions: "Slice the tomato and season it with salt and pepper.\nLayer tomato and basil on one slice of bread.\nDrizzle with olive oil, close the sandwich and serve.",
		Servings:     1,
	},
}

// Ensure creates or repairs the demo account and, on a database with no
// approved recipes, seeds two public demo recipes. It is safe to call on every
// boot.
func Ensure(ctx context.Context, db *gorm.DB, cfg *config.Config) error {
	if cfg == nil || !cfg.DemoUserEnabled {
		return nil
	}
	if db == nil {
		return fmt.Errorf("demo: database handle is nil")
	}

	username := strings.ToLower(strings.TrimSpace(cfg.DemoUserUsername))
	if username == "" {
		return fmt.Errorf("DEMO_USER_USERNAME must not be empty")
	}
	email := strings.ToLower(strings.TrimSpace(cfg.DemoUserEmail))
	if email == "" {
		return fmt.Errorf("DEMO_USER_EMAIL must not be empty")
	}
	if cfg.DemoUserPassword == "" {
		return fmt.Errorf("DEMO_USER_PASSWORD must not be empty")
	}

	db = db.WithContext(ctx)

	var user database.User
	err := db.Where("username = ?", username).First(&user).Error
	created := false
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("demo user: %w", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(cfg.DemoUserPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			return hashErr
		}
		now := time.Now()
		user = database.User{
			Username:        username,
			Email:           email,
			Password:        string(hash),
			FirstName:       "Demo",
			Bio:             demoBio,
			Language:        "en",
			Theme:           "light",
			IsActive:        true,
			IsDemo:          true,
			AcceptedTermsAt: &now,
			DateJoined:      now,
		}
		if createErr := db.Create(&user).Error; createErr != nil {
			return fmt.Errorf("create demo user: %w", createErr)
		}
		created = true
	}

	if !created {
		// Never take over an account with administrative rights. A demo login
		// must never expose an admin, even if an operator reuses the username.
		if user.IsSuperuser || user.IsStaff {
			return fmt.Errorf("DEMO_USER_USERNAME %q belongs to a staff account; refusing to expose it as the demo login", username)
		}
		// Only repair an account this package already owns. Resetting a real
		// user's password because their username collides with the configured
		// demo name would hand their account to anyone who reads the login page.
		if !user.IsDemo {
			return fmt.Errorf("DEMO_USER_USERNAME %q belongs to a normal account; refusing to take it over", username)
		}

		updates := map[string]interface{}{
			"is_demo":   true,
			"is_active": true,
		}
		if strings.TrimSpace(user.Email) == "" {
			updates["email"] = email
		}
		if strings.TrimSpace(user.FirstName) == "" {
			updates["first_name"] = "Demo"
		}
		if strings.TrimSpace(user.Bio) == "" {
			updates["bio"] = demoBio
		}
		if user.AcceptedTermsAt == nil {
			now := time.Now()
			updates["accepted_terms_at"] = now
		}
		// Only re-hash when the configured password no longer matches, so the
		// stored hash is not rewritten on every boot.
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cfg.DemoUserPassword)) != nil {
			hash, hashErr := bcrypt.GenerateFromPassword([]byte(cfg.DemoUserPassword), bcrypt.DefaultCost)
			if hashErr != nil {
				return hashErr
			}
			updates["password"] = string(hash)
		}
		if updateErr := db.Model(&database.User{}).Where("id = ?", user.ID).Updates(updates).Error; updateErr != nil {
			return fmt.Errorf("repair demo user: %w", updateErr)
		}
		// Reflect the updates on the in-memory row for the seed steps below.
		db.First(&user, user.ID)
	}

	// Community membership, without duplicating an existing join row.
	var community database.Group
	if err := db.Where("name = ?", "community").FirstOrCreate(&community, database.Group{Name: "community"}).Error; err != nil {
		return fmt.Errorf("demo community group: %w", err)
	}
	var memberships int64
	if err := db.Table("user_groups").
		Where("user_id = ? AND group_id = ?", user.ID, community.ID).
		Count(&memberships).Error; err != nil {
		return fmt.Errorf("demo membership: %w", err)
	}
	if memberships == 0 {
		if err := db.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", user.ID, community.ID).Error; err != nil {
			return fmt.Errorf("demo membership: %w", err)
		}
	}

	return seedPublicRecipes(db, &user)
}

func seedPublicRecipes(db *gorm.DB, user *database.User) error {
	// Only populate a database that has no approved content at all, so an
	// existing instance is never given reviewer-only demo recipes.
	var approved int64
	if err := db.Model(&database.Recipe{}).Where("is_approved = ?", true).Count(&approved).Error; err != nil {
		return fmt.Errorf("count recipes: %w", err)
	}
	if approved > 0 {
		return nil
	}

	for _, seed := range seedRecipes {
		var existing int64
		if err := db.Model(&database.Recipe{}).Where("title = ?", seed.Title).Count(&existing).Error; err != nil {
			return fmt.Errorf("check demo recipe: %w", err)
		}
		if existing > 0 {
			continue
		}

		recipe := database.Recipe{
			Title:        seed.Title,
			Description:  seed.Description,
			Ingredients:  seed.Ingredients,
			Instructions: seed.Instructions,
			Servings:     seed.Servings,
			IsApproved:   true,
		}
		recipe.UploadedByID = &user.ID
		if err := db.Create(&recipe).Error; err != nil {
			return fmt.Errorf("seed demo recipe %q: %w", seed.Title, err)
		}

		for _, name := range []string{"demo", "vegetarian"} {
			var tag database.Tag
			if err := db.Where("name = ?", name).FirstOrCreate(&tag, database.Tag{Name: name}).Error; err != nil {
				return fmt.Errorf("seed demo tag %q: %w", name, err)
			}
			if err := db.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", recipe.ID, tag.ID).Error; err != nil {
				return fmt.Errorf("seed demo recipe tag %q: %w", name, err)
			}
		}
	}
	return nil
}
