// Package render provides a gin HTMLRender that buffers template execution
// and surfaces execution errors as HTTP 500 responses instead of silently
// writing a truncated page.
//
// gin's default render.HTML calls ExecuteTemplate directly on the response
// writer and discards the returned error: a template bug produces a 200
// response with partial HTML and no log entry. This renderer executes into a
// buffer first, so an error means nothing has been written yet and can be
// reported as a clean 500 plus a log line.
package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin/render"
)

// HTMLRender satisfies gin's render.HTMLRender interface.
type HTMLRender struct {
	Template *template.Template
}

// Instance creates a render for one template invocation.
func (r HTMLRender) Instance(name string, data interface{}) render.Render {
	return HTML{Template: r.Template, Name: name, Data: data}
}

// HTML is a single template execution.
type HTML struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

// Render executes the template into a buffer and only then writes it out.
func (r HTML) Render(w http.ResponseWriter) error {
	writeContentType(w, "text/html; charset=utf-8")

	var buf bytes.Buffer
	var err error
	if r.Name == "" {
		err = r.Template.Execute(&buf, r.Data)
	} else {
		err = r.Template.ExecuteTemplate(&buf, r.Name, r.Data)
	}
	if err != nil {
		log.Printf("template error rendering %q: %v", r.Name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

// WriteContentType sets the response Content-Type.
func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}

func writeContentType(w http.ResponseWriter, value string) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", value)
	}
}
