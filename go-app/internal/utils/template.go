package utils

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
	"github.com/martynvdijke/sandwitches-go/internal/middleware"
)

type TemplateData struct {
	User       *database.User
	Flashes    []FlashMessage
	CSRFToken  string
	CSRFHidden template.HTML
	Error      string
	EinkMode   bool
	Data       gin.H
}

// IsEinkMode detects e-ink mode: URL param ?eink=1 > cookie eink_mode=1 > user theme.
func IsEinkMode(c *gin.Context, user *database.User) bool {
	if c.Query("eink") == "1" {
		return true
	}
	if cookie, err := c.Cookie("eink_mode"); err == nil && cookie == "1" {
		return true
	}
	return user != nil && user.Theme == "eink"
}

func NewTemplateData(c *gin.Context) *TemplateData {
	td := &TemplateData{
		User:       middleware.GetUser(c),
		Flashes:    GetFlashes(c),
		CSRFToken:  "",
		CSRFHidden: "",
		Data:       make(gin.H),
	}
	td.EinkMode = IsEinkMode(c, td.User)
	if tok, exists := c.Get("csrf_token"); exists {
		td.CSRFToken = tok.(string)
		td.CSRFHidden = template.HTML(`<input type="hidden" name="csrf_token" value="` + tok.(string) + `">`)
	}
	return td
}

func (td *TemplateData) With(key string, value interface{}) *TemplateData {
	td.Data[key] = value
	return td
}

func (td *TemplateData) ToGinH() gin.H {
	h := gin.H{
		"user":    td.User,
		"flashes": td.Flashes,
	}
	if td.CSRFHidden != "" {
		h["csrf_token"] = td.CSRFHidden
		h["csrf_token_value"] = td.CSRFToken
	}
	if td.Error != "" {
		h["error"] = td.Error
	}
	h["eink_mode"] = td.EinkMode
	for k, v := range td.Data {
		h[k] = v
	}
	return h
}
