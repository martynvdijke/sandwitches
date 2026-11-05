package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/account"
	"github.com/martynvdijke/sandwitches-go/internal/config"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
	"github.com/martynvdijke/sandwitches-go/internal/utils"
)

// legalCfg is the deployment configuration published on the legal pages. It is
// set once in main. When nil (unit tests, unusual embedders) legalData falls
// back to environment variables and neutral defaults.
var legalCfg *config.Config

// SetLegalConfig publishes the instance configuration used by the legal pages.
func SetLegalConfig(cfg *config.Config) { legalCfg = cfg }

// LegalData is the identity of the instance operator plus the deployment facts
// the published policy is allowed to assert. Each field is either a config
// value or read from the live settings row, so the text cannot claim a
// processor the running server does not use.
type LegalData struct {
	Controller           string
	ContactEmail         string
	ContactAddress       string
	DPOEmail             string
	SupervisoryAuthority string
	HostingRegion        string
	BaseURL              string
	SessionCookie        string

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

// legalValue prefers the configured value, then the environment, then the
// fallback, so tests keep working without calling SetLegalConfig.
func legalValue(configured, key, fallback string) string {
	if v := strings.TrimSpace(configured); v != "" {
		return v
	}
	return legalEnv(key, fallback)
}

func legalProcessorEnv(configured, key string) string {
	if legalCfg != nil {
		if v := strings.TrimSpace(configured); v != "" {
			return v
		}
	}
	return strings.TrimSpace(os.Getenv(key))
}

// legalData collects the policy facts once per request.
func legalData() LegalData {
	name := "Sandwitches"
	var setting database.Setting
	if err := database.DB.First(&setting).Error; err == nil && strings.TrimSpace(setting.SiteName) != "" {
		name = strings.TrimSpace(setting.SiteName)
	}

	d := LegalData{
		Controller:           "the operator of this " + name + " instance",
		HostingRegion:        "the European Union",
		BaseURL:              "https://sandwitches.vandijke.xyz",
		SessionCookie:        middleware.SessionCookieName,
		SupervisoryAuthority: "your local data protection authority",
	}

	var (
		cfgController, cfgContact, cfgAddress, cfgDPO, cfgAuthority string
		cfgHosting, cfgBase                                         string
		cfgUmamiHost, cfgUmamiID                                    string
	)
	if legalCfg != nil {
		cfgController = legalCfg.LegalControllerName
		cfgContact = legalCfg.LegalContactEmail
		cfgAddress = legalCfg.LegalContactAddress
		cfgDPO = legalCfg.LegalDPOEmail
		cfgAuthority = legalCfg.LegalSupervisoryAuthority
		cfgHosting = legalCfg.LegalHostingRegion
		cfgBase = legalCfg.LegalBaseURL
		cfgUmamiHost = legalCfg.UmamiHost
		cfgUmamiID = legalCfg.UmamiWebsiteID
	}

	if v := legalValue(cfgController, "LEGAL_CONTROLLER_NAME", ""); v != "" {
		d.Controller = v
	}
	d.ContactEmail = legalValue(cfgContact, "LEGAL_CONTACT_EMAIL", strings.TrimSpace(setting.Email))
	d.ContactAddress = legalValue(cfgAddress, "LEGAL_CONTACT_ADDRESS", "")
	d.DPOEmail = legalValue(cfgDPO, "LEGAL_DPO_EMAIL", "")
	d.SupervisoryAuthority = legalValue(cfgAuthority, "LEGAL_SUPERVISORY_AUTHORITY", d.SupervisoryAuthority)
	d.HostingRegion = legalValue(cfgHosting, "LEGAL_HOSTING_REGION", d.HostingRegion)
	d.BaseURL = strings.TrimRight(legalValue(cfgBase, "LEGAL_BASE_URL", d.BaseURL), "/")

	d.AnalyticsURL = legalProcessorEnv(cfgUmamiHost, "UMAMI_HOST")
	d.Analytics = d.AnalyticsURL != "" && legalProcessorEnv(cfgUmamiID, "UMAMI_WEBSITE_ID") != ""

	d.Traces = legalProcessorEnv("", "OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		legalProcessorEnv("", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" ||
		legalProcessorEnv("", "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") != "" ||
		legalProcessorEnv("", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT") != "" ||
		setting.OTelEnabled || strings.TrimSpace(setting.OTelEndpoint) != ""

	d.Mail = legalProcessorEnv("", "SEND_EMAIL") == "true" ||
		legalProcessorEnv("", "SMTP_HOST") != "" ||
		(setting.SMTPEnabled && strings.TrimSpace(setting.SMTPHost) != "")

	d.Gotify = legalProcessorEnv("", "GOTIFY_URL") != "" ||
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
	td.Analytics = utils.AnalyticsEnabled() && (td.User == nil || !td.User.AnalyticsOptOut)
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
	case account.ErrDemoAccount:
		utils.AddFlash(c, "error", "The demo account cannot be deleted.")
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
