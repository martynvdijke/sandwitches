package utils

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
)

var mdRenderer = goldmark.New()

func RenderMarkdown(input string) template.HTML {
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(input), &buf); err != nil {
		return template.HTML(input)
	}
	return template.HTML(buf.String())
}
