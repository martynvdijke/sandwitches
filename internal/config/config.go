package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Debug              bool
	SecretKey          string
	AllowedHosts       string
	CSRFTrustedOrigins string
	DatabaseFile       string
	MediaRoot          string
	LogFile            string
	SendEmail          bool
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPassword       string
	SMTPFromEmail      string
	GotifyURL          string
	GotifyToken        string
	UmamiHost          string
	UmamiWebsiteID     string
	LanguageCode       string
	StaticRoot         string
	BaseDir            string
	APICORSOrigins     string
	APIRateLimitRate   float64
	APIRateLimitBurst  float64

	// Legal identity published on /privacy, /terms and /account/delete. Empty
	// values fall back to neutral wording so a self-hosted instance can publish
	// the pages before its operator fills them in.
	LegalControllerName string
	LegalContactEmail   string
	LegalContactAddress string
	LegalHostingRegion  string
	LegalBaseURL        string
	LegalDPOEmail       string
	// LegalSupervisoryAuthority names the data-protection authority a person
	// may complain to. Empty falls back to neutral wording.
	LegalSupervisoryAuthority string

	// Demo login. The hosted instance at sandwitches.vandijke.xyz enables this
	// by default; self-hosters opt in with DEMO_USER_ENABLED=true. The account
	// is deliberately public and shared, so it must never have admin rights.
	DemoUserEnabled  bool
	DemoUserUsername string
	DemoUserPassword string
	DemoUserEmail    string
}

func Load() *Config {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}

	cfg := &Config{
		Debug:                     os.Getenv("DEBUG") == "true",
		SecretKey:                 os.Getenv("SECRET_KEY"),
		AllowedHosts:              os.Getenv("ALLOWED_HOSTS"),
		CSRFTrustedOrigins:        os.Getenv("CSRF_TRUSTED_ORIGINS"),
		DatabaseFile:              os.Getenv("DATABASE_FILE"),
		MediaRoot:                 os.Getenv("MEDIA_ROOT"),
		LogFile:                   os.Getenv("LOG_FILE"),
		SendEmail:                 os.Getenv("SEND_EMAIL") == "true",
		SMTPHost:                  os.Getenv("SMTP_HOST"),
		SMTPPort:                  os.Getenv("SMTP_PORT"),
		SMTPUser:                  os.Getenv("SMTP_USER"),
		SMTPPassword:              os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:             envOr("SMTP_FROM_EMAIL", "noreply@sandwitches.local"),
		GotifyURL:                 os.Getenv("GOTIFY_URL"),
		GotifyToken:               os.Getenv("GOTIFY_TOKEN"),
		UmamiHost:                 os.Getenv("UMAMI_HOST"),
		UmamiWebsiteID:            os.Getenv("UMAMI_WEBSITE_ID"),
		LanguageCode:              envOr("LANGUAGE_CODE", "en"),
		BaseDir:                   baseDir,
		APICORSOrigins:            os.Getenv("API_CORS_ORIGINS"),
		APIRateLimitRate:          envFloat("API_RATE_LIMIT_RATE", 10),
		APIRateLimitBurst:         envFloat("API_RATE_LIMIT_BURST", 100),
		LegalControllerName:       os.Getenv("LEGAL_CONTROLLER_NAME"),
		LegalContactEmail:         os.Getenv("LEGAL_CONTACT_EMAIL"),
		LegalContactAddress:       os.Getenv("LEGAL_CONTACT_ADDRESS"),
		LegalHostingRegion:        os.Getenv("LEGAL_HOSTING_REGION"),
		LegalBaseURL:              os.Getenv("LEGAL_BASE_URL"),
		LegalDPOEmail:             os.Getenv("LEGAL_DPO_EMAIL"),
		LegalSupervisoryAuthority: os.Getenv("LEGAL_SUPERVISORY_AUTHORITY"),

		DemoUserUsername: strings.ToLower(strings.TrimSpace(envOr("DEMO_USER_USERNAME", "demo"))),
		DemoUserPassword: envOr("DEMO_USER_PASSWORD", "demo"),
		DemoUserEmail:    strings.ToLower(strings.TrimSpace(envOr("DEMO_USER_EMAIL", "demo@sandwitches.local"))),
	}
	cfg.DemoUserEnabled = envBool("DEMO_USER_ENABLED", isDemoHost(envOr("LEGAL_BASE_URL", "https://sandwitches.vandijke.xyz")))

	if cfg.DatabaseFile == "" {
		cfg.DatabaseFile = filepath.Join(baseDir, "db.sqlite3")
	}
	if cfg.MediaRoot == "" {
		cfg.MediaRoot = filepath.Join(baseDir, "media")
	}
	if cfg.LogFile == "" {
		cfg.LogFile = filepath.Join(cfg.MediaRoot, "sandwitches.log")
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

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

// envBool accepts the usual true/false spellings and keeps def when unset.
func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return def
}

// isDemoHost reports whether the configured base URL is the hosted
// Sandwitches instance. Only that instance gets a demo login by default.
func isDemoHost(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Hostname(), "sandwitches.vandijke.xyz")
}
