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
	Data       gin.H
}

func NewTemplateData(c *gin.Context) *TemplateData {
	td := &TemplateData{
		User:       middleware.GetUser(c),
		Flashes:    GetFlashes(c),
		CSRFToken:  "",
		CSRFHidden: "",
		Data:       make(gin.H),
	}
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
	for k, v := range td.Data {
		h[k] = v
	}
	return h
}
