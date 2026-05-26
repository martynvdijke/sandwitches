package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Debug                       bool
	SecretKey                   string
	AllowedHosts                string
	CSRFTrustedOrigins          string
	DatabaseFile                string
	MediaRoot                   string
	SendEmail                   bool
	SMTPHost                    string
	SMTPPort                    string
	SMTPUser                    string
	SMTPPassword                string
	SMTPFromEmail               string
	GotifyURL                   string
	GotifyToken                 string
	UmamiHost                   string
	UmamiWebsiteID              string
	LanguageCode                string
	StaticRoot                  string
	BaseDir                     string
}

func Load() *Config {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	cfg := &Config{
		Debug:              os.Getenv("DEBUG") == "true",
		SecretKey:          os.Getenv("SECRET_KEY"),
		AllowedHosts:       os.Getenv("ALLOWED_HOSTS"),
		CSRFTrustedOrigins: os.Getenv("CSRF_TRUSTED_ORIGINS"),
		DatabaseFile:       os.Getenv("DATABASE_FILE"),
		MediaRoot:          os.Getenv("MEDIA_ROOT"),
		SendEmail:          os.Getenv("SEND_EMAIL") == "true",
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           os.Getenv("SMTP_PORT"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:      envOr("SMTP_FROM_EMAIL", "noreply@sandwitches.local"),
		GotifyURL:          os.Getenv("GOTIFY_URL"),
		GotifyToken:        os.Getenv("GOTIFY_TOKEN"),
		UmamiHost:          os.Getenv("UMAMI_HOST"),
		UmamiWebsiteID:     os.Getenv("UMAMI_WEBSITE_ID"),
		LanguageCode:       envOr("LANGUAGE_CODE", "en"),
		BaseDir:            baseDir,
	}

	if cfg.DatabaseFile == "" {
		cfg.DatabaseFile = filepath.Join(baseDir, "db.sqlite3")
	}
	if cfg.MediaRoot == "" {
		cfg.MediaRoot = filepath.Join(baseDir, "media")
	}
	if cfg.SecretKey == "" && !cfg.Debug {
		panic("SECRET_KEY must be set in production")
	}

	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
