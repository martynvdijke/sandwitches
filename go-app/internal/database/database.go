package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/martynvdijke/sandwitches-go/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var err error

	logLevel := logger.Warn
	if cfg.Debug {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(cfg.DatabaseFile), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := DB.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
		log.Printf("Warning: could not set WAL mode: %v", err)
	}
	if err := DB.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		log.Printf("Warning: could not enable foreign keys: %v", err)
	}

	models := []interface{}{
		&Setting{},
		&User{},
		&Group{},
		&Tag{},
		&Recipe{},
		&Rating{},
		&Order{},
		&OrderItem{},
		&CartItem{},
		&RecipeHistory{},
	}
	for _, m := range models {
		if err := DB.AutoMigrate(m); err != nil {
			log.Fatalf("Failed to migrate %T: %v", m, err)
		}
	}

	var count int64
	DB.Model(&Setting{}).Count(&count)
	if count == 0 {
		DB.Create(&Setting{SiteName: "Sandwitches", LogLevel: "INFO"})
	}
	DB.Model(&Group{}).Where("name = ?", "admin").FirstOrCreate(&Group{Name: "admin"})
	DB.Model(&Group{}).Where("name = ?", "community").FirstOrCreate(&Group{Name: "community"})

	fmt.Println("Database initialized")
}

func Slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	runes := []rune(s)
	if len(runes) > 60 {
		s = string(runes[:60])
	}
	return strings.TrimFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

func UniqueSlug(base string, existingID *uint) string {
	slug := Slugify(base)
	if slug == "" {
		slug = "untitled"
	}
	if existingID != nil && *existingID > 0 {
		var count int64
		DB.Model(&Recipe{}).Where("slug = ? AND id != ?", slug, *existingID).Count(&count)
		if count == 0 {
			return slug
		}
	} else {
		var count int64
		DB.Model(&Recipe{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
	}
	for i := 1; i <= 100; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		var count int64
		query := DB.Model(&Recipe{}).Where("slug = ?", candidate)
		if existingID != nil && *existingID > 0 {
			query = query.Where("id != ?", *existingID)
		}
		query.Count(&count)
		if count == 0 {
			return candidate
		}
	}
	return slug + fmt.Sprintf("-%d", time.Now().Unix())
}

func RecordRecipeHistory(tx *gorm.DB, recipeID uint, changedByID *uint, fieldName, oldValue, newValue string) {
	tx.Create(&RecipeHistory{
		RecipeID:    recipeID,
		ChangedByID: changedByID,
		FieldName:   fieldName,
		OldValue:    oldValue,
		NewValue:    newValue,
		ChangedAt:   time.Now(),
	})
}
