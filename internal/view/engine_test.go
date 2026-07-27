package view_test

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fluxgo/internal/view"
)

func TestRenderSelectsFullAndPartialTemplates(t *testing.T) {
	root := createViews(t, ".gohtml")
	engine, err := view.New(view.Config{Root: root})
	if err != nil {
		t.Fatalf("boot template engine: %v", err)
	}

	tests := []struct {
		name string
		htmx bool
		want string
	}{
		{name: "full", want: "<html><body>Hello Milan</body></html>"},
		{name: "partial", htmx: true, want: "Hello Milan"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.htmx {
				request.Header.Set("HX-Request", "true")
			}
			response := httptest.NewRecorder()

			if err := engine.Render(response, request, "home.gohtml", "Milan"); err != nil {
				t.Fatalf("render template: %v", err)
			}
			if got := response.Body.String(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCustomExtension(t *testing.T) {
	root := createViews(t, ".tpl")
	engine, err := view.New(view.Config{Root: root, Extension: "tpl"})
	if err != nil {
		t.Fatalf("boot template engine: %v", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := engine.Render(response, request, "home", "FluxGo"); err != nil {
		t.Fatalf("render template: %v", err)
	}
	if !strings.Contains(response.Body.String(), "Hello FluxGo") {
		t.Fatalf("unexpected response %q", response.Body.String())
	}
}

func TestTemplateHelpersAreAvailableToFullAndPartialTemplates(t *testing.T) {
	root := createViews(t, ".gohtml")
	pagePath := filepath.Join(root, "pages", "home.gohtml")
	if err := os.WriteFile(
		pagePath,
		[]byte(`{{define "content"}}{{shout .}}{{end}}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	engine, err := view.New(view.Config{
		Root: root,
		Funcs: template.FuncMap{
			"shout": strings.ToUpper,
		},
	})
	if err != nil {
		t.Fatalf("boot template engine: %v", err)
	}

	for _, htmx := range []bool{false, true} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		if htmx {
			request.Header.Set("HX-Request", "true")
		}
		response := httptest.NewRecorder()

		if err := engine.Render(response, request, "home", "hello"); err != nil {
			t.Fatalf("render template: %v", err)
		}
		if !strings.Contains(response.Body.String(), "HELLO") {
			t.Fatalf("helper output missing from %q", response.Body.String())
		}
	}
}

func TestGlobalFrameworkHelpersAreRegisteredAutomatically(t *testing.T) {
	root := createViews(t, ".gohtml")
	pagePath := filepath.Join(root, "pages", "home.gohtml")
	if err := os.WriteFile(
		pagePath,
		[]byte(`{{define "content"}}{{upper .}}{{end}}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	engine, err := view.New(view.Config{Root: root})
	if err != nil {
		t.Fatalf("boot template engine: %v", err)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := engine.Render(response, request, "home", "global helper"); err != nil {
		t.Fatalf("render template: %v", err)
	}
	if !strings.Contains(response.Body.String(), "GLOBAL HELPER") {
		t.Fatalf("global helper output missing from %q", response.Body.String())
	}
}

func createViews(t *testing.T, extension string) string {
	t.Helper()

	root := t.TempDir()
	layouts := filepath.Join(root, "layouts")
	pages := filepath.Join(root, "pages")
	if err := os.MkdirAll(layouts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(pages, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(layouts, "main"+extension),
		[]byte(`{{define "layout"}}<html><body>{{template "content" .}}</body></html>{{end}}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(pages, "home"+extension),
		[]byte(`{{define "content"}}Hello {{.}}{{end}}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return root
}
