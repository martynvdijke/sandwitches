package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"
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

func slugify(s string) string {
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
