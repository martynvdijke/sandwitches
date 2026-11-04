package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/account"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
)

// LegalData is the identity of the instance operator plus the deployment facts
// the published policy is allowed to assert. Each field is either an env value
// or read from the live settings row, so the text cannot claim a processor the
// running server does not use.
type LegalData struct {
	Controller     string
	ContactEmail   string
	ContactAddress string
	HostingRegion  string
	BaseURL        string
	SessionCookie  string

	Analytics    bool
	AnalyticsURL string
	Traces       bool
	Mail         bool
	Gotify       bool
	AI           bool
	OIDC         bool
}

func legalEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// legalData collects the policy facts once per request.
func legalData() LegalData {
	name := "Sandwitches"
	var setting database.Setting
	if err := database.DB.First(&setting).Error; err == nil && strings.TrimSpace(setting.SiteName) != "" {
		name = strings.TrimSpace(setting.SiteName)
	}

	d := LegalData{
		Controller:    "the operator of this " + name + " instance",
		HostingRegion: "the European Union",
		BaseURL:       "https://sandwitches.vandijke.xyz",
		SessionCookie: middleware.SessionCookieName,
	}

	if v := legalEnv("LEGAL_CONTROLLER_NAME", ""); v != "" {
		d.Controller = v
	}
	d.ContactEmail = legalEnv("LEGAL_CONTACT_EMAIL", strings.TrimSpace(setting.Email))
	d.ContactAddress = legalEnv("LEGAL_CONTACT_ADDRESS", "")
	d.HostingRegion = legalEnv("LEGAL_HOSTING_REGION", d.HostingRegion)
	d.BaseURL = strings.TrimRight(legalEnv("LEGAL_BASE_URL", d.BaseURL), "/")

	d.AnalyticsURL = strings.TrimSpace(os.Getenv("UMAMI_HOST"))
	d.Analytics = d.AnalyticsURL != "" && strings.TrimSpace(os.Getenv("UMAMI_WEBSITE_ID")) != ""

	d.Traces = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != "" ||
		setting.OTelEnabled || strings.TrimSpace(setting.OTelEndpoint) != ""

	d.Mail = os.Getenv("SEND_EMAIL") == "true" ||
		strings.TrimSpace(os.Getenv("SMTP_HOST")) != "" ||
		(setting.SMTPEnabled && strings.TrimSpace(setting.SMTPHost) != "")

	d.Gotify = strings.TrimSpace(os.Getenv("GOTIFY_URL")) != "" ||
		strings.TrimSpace(setting.GotifyURL) != ""
	d.AI = strings.TrimSpace(setting.AIConnectionPoint) != ""
	d.OIDC = loadOIDCConfig().valid()

	return d
}

// legalTemplateData builds template data without touching the session store, so
// reading a public legal page never issues a session cookie.
func legalTemplateData(c *gin.Context) *utils.TemplateData {
	td := &utils.TemplateData{
		User:       middleware.GetUser(c),
		Flashes:    nil,
		CSRFToken:  "",
		CSRFHidden: "",
		Data:       make(gin.H),
	}
	td.EinkMode = utils.IsEinkMode(c, td.User)
	return td
}

// PrivacyPage serves the public privacy policy. It stays reachable without a
// session because both app stores and data-protection law require the policy to
// be readable before anyone signs up.
func PrivacyPage(c *gin.Context) {
	td := legalTemplateData(c)
	c.HTML(http.StatusOK, "privacy.html", td.With("legal", legalData()).ToGinH())
}

// TermsPage serves the acceptable-use terms for the same reason.
func TermsPage(c *gin.Context) {
	td := legalTemplateData(c)
	c.HTML(http.StatusOK, "terms.html", td.With("legal", legalData()).ToGinH())
}

// AccountDeletePage is the deletion route Google Play asks every app with
// accounts to publish. It is readable while signed out, because store review
// and a user who can no longer sign in both need to reach it without a session.
func AccountDeletePage(c *gin.Context) {
	td := legalTemplateData(c)
	c.HTML(http.StatusOK, "account_delete.html", td.
		With("legal", legalData()).
		With("deleted", c.Query("deleted") == "1").
		ToGinH())
}

// DeleteAccount is the signed-in HTML path. It requires a CSRF token (the form
// is inside the CSRF group) and the account's own username typed as
// confirmation.
func DeleteAccount(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login?next=/settings")
		return
	}

	confirm := strings.TrimSpace(c.PostForm("confirm"))
	if confirm != user.Username {
		utils.AddFlash(c, "error", "Type your username exactly to confirm account deletion.")
		c.Redirect(http.StatusFound, "/settings")
		return
	}

	err := account.New(database.DB).Delete(user.ID)
	switch err {
	case nil:
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/account/delete?deleted=1")
	case account.ErrLastSuperuser:
		utils.AddFlash(c, "error", "This is the last superuser account. Promote another account before deleting it.")
		c.Redirect(http.StatusFound, "/settings")
	case account.ErrNotFound:
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.Redirect(http.StatusFound, "/account/delete?deleted=1")
	default:
		utils.AddFlash(c, "error", "The account could not be deleted. Please try again.")
		c.Redirect(http.StatusFound, "/settings")
	}
}
