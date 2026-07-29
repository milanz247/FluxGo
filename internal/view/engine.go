// Package view provides precompiled, HTMX-aware HTML template rendering.
package view

import (
	"fluxgo/internal/helpers"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	defaultExtension = ".gohtml"
	partialSuffix    = "_partial"
)

// Config controls where templates are loaded from.
type Config struct {
	// Root is the views directory containing layouts/ and pages/.
	Root string
	// Extension is the template file extension. It defaults to ".gohtml".
	Extension string
	// Funcs contains helpers registered before templates are parsed.
	Funcs template.FuncMap
}

// Engine holds immutable, precompiled templates safe for concurrent rendering.
type Engine struct {
	templates map[string]*template.Template
	extension string
}

// New compiles the main layout and every page into full and partial versions.
// It performs all template disk I/O; Render only reads the in-memory map.
func New(config Config) (*Engine, error) {
	extension := normalizeExtension(config.Extension)
	funcs := helpers.TemplateFuncs()
	for name, helper := range config.Funcs {
		funcs[name] = helper
	}

	root := config.Root
	if strings.TrimSpace(root) == "" {
		root = "views"
	}

	layoutPath := filepath.Join(root, "layouts", "main"+extension)
	pagesRoot := filepath.Join(root, "pages")
	pagePaths, err := findPages(pagesRoot, extension)
	if err != nil {
		return nil, fmt.Errorf("find page templates: %w", err)
	}
	if len(pagePaths) == 0 {
		return nil, fmt.Errorf("no page templates found in %q", pagesRoot)
	}

	engine := &Engine{
		templates: make(map[string]*template.Template, len(pagePaths)*2),
		extension: extension,
	}

	for _, pagePath := range pagePaths {
		relative, err := filepath.Rel(pagesRoot, pagePath)
		if err != nil {
			return nil, fmt.Errorf("resolve page template %q: %w", pagePath, err)
		}
		name := strings.TrimSuffix(filepath.ToSlash(relative), extension)

		full, err := template.New(name).Funcs(funcs).ParseFiles(layoutPath, pagePath)
		if err != nil {
			return nil, fmt.Errorf("compile full template %q: %w", name, err)
		}

		partial, err := template.New(name + partialSuffix).Funcs(funcs).ParseFiles(pagePath)
		if err != nil {
			return nil, fmt.Errorf("compile partial template %q: %w", name, err)
		}

		engine.templates[name] = full
		engine.templates[name+partialSuffix] = partial
	}

	return engine, nil
}

// Render selects a full or partial precompiled template based on HX-Request.
func (e *Engine) Render(w http.ResponseWriter, r *http.Request, name string, data any) error {
	name = e.cleanName(name)
	isHTMX := strings.EqualFold(strings.TrimSpace(r.Header.Get("HX-Request")), "true")

	key := name
	templateName := "layout"
	if isHTMX {
		key += partialSuffix
		templateName = "content"
	}

	compiled, ok := e.templates[key]
	if !ok {
		return fmt.Errorf("template %q is not loaded", key)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return compiled.ExecuteTemplate(w, templateName, data)
}

func (e *Engine) cleanName(name string) string {
	name = filepath.ToSlash(strings.TrimSpace(name))
	name = strings.TrimSuffix(name, e.extension)
	return strings.TrimSuffix(name, partialSuffix)
}

// findPages walks the pages directory recursively so pages can be organized
// into feature subfolders, e.g. pages/auth/login.gohtml renders as "auth/login".
func findPages(pagesRoot, extension string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(pagesRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != extension {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func normalizeExtension(extension string) string {
	extension = strings.TrimSpace(extension)
	if extension == "" {
		return defaultExtension
	}
	if !strings.HasPrefix(extension, ".") {
		return "." + extension
	}
	return extension
}
