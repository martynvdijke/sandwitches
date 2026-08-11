package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTMLRenderInstance(t *testing.T) {
	tmpl := template.Must(template.New("page").Parse("hello {{.Name}}"))
	r := HTMLRender{Template: tmpl}.Instance("page", map[string]string{"Name": "world"})
	html, ok := r.(HTML)
	if !ok {
		t.Fatalf("Instance returned %T, want render.HTML", r)
	}
	if html.Name != "page" {
		t.Fatalf("Name = %q, want %q", html.Name, "page")
	}
}

func TestHTMLRenderSuccess(t *testing.T) {
	tmpl := template.Must(template.New("page").Parse("hello {{.Name}}"))
	r := HTMLRender{Template: tmpl}.Instance("page", map[string]string{"Name": "world"})

	rec := httptest.NewRecorder()
	if err := r.Render(rec); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "hello world" {
		t.Fatalf("body = %q, want %q", got, "hello world")
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
}

// TestHTMLRenderExecutionError is the regression test for the homepage
// truncation bug: when a template dereferences a nil pointer (e.g.
// {{ .User.Username }} with User == nil), gin's default renderer wrote a
// partial 200 page. Our renderer must surface the error as a 500 and must
// not write partial content.
func TestHTMLRenderExecutionError(t *testing.T) {
	tmpl := template.Must(template.New("page").Parse(
		`before{{ .User.Username }}after`,
	))
	r := HTMLRender{Template: tmpl}.Instance("page", map[string]interface{}{
		"User": (*struct{ Username string })(nil),
	})

	rec := httptest.NewRecorder()
	if err := r.Render(rec); err == nil {
		t.Fatal("Render returned nil error, want template execution error")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if body := rec.Body.String(); strings.Contains(body, "before") {
		t.Fatalf("partial content leaked into 500 response: %q", body)
	}
}

func TestWriteContentTypeOnlySetsWhenEmpty(t *testing.T) {
	tmpl := template.Must(template.New("page").Parse("x"))

	// Fresh recorder: header gets set.
	rec := httptest.NewRecorder()
	HTML{Template: tmpl, Name: "page"}.WriteContentType(rec)
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}

	// Header already present: must not be overwritten.
	rec2 := httptest.NewRecorder()
	rec2.Header().Set("Content-Type", "application/custom")
	HTML{Template: tmpl, Name: "page"}.WriteContentType(rec2)
	if ct := rec2.Header().Get("Content-Type"); ct != "application/custom" {
		t.Fatalf("Content-Type = %q, want existing value preserved", ct)
	}
}

func TestRenderWithoutNameExecutesRoot(t *testing.T) {
	tmpl := template.Must(template.New("root").Parse("root output"))
	r := HTML{Template: tmpl, Name: ""}

	rec := httptest.NewRecorder()
	if err := r.Render(rec); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if got := rec.Body.String(); got != "root output" {
		t.Fatalf("body = %q, want %q", got, "root output")
	}
}
