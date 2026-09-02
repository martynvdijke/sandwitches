package utils

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

var mdParser = parser.New()
var mdRenderer = html.New()

func RenderMarkdown(input string) template.HTML {
	var buf bytes.Buffer
	src := []byte(input)
	doc := mdParser.Parse(src)
	if err := mdRenderer.Render(&buf, src, doc); err != nil {
		return template.HTML(input)
	}
	return template.HTML(buf.String())
}
