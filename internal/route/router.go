package route

import (
	"fluxgo/internal/middleware"
	"log"
	"net/http"
	"strings"
)

type Handler func(*Context) error

// Renderer renders a named view for the current request.
type Renderer interface {
	Render(http.ResponseWriter, *http.Request, string, any) error
}

// Middleware wraps a Handler to run logic before or after it.
type Middleware = middleware.Middleware[Handler]

// Engine registers routes and serves HTTP requests.
type Engine struct {
	mux        *http.ServeMux
	middleware *middleware.Engine[Handler]
	renderer   Renderer
}

// SetRenderer configures the view renderer used by route contexts.
func (e *Engine) SetRenderer(renderer Renderer) {
	e.renderer = renderer
}

func New() *Engine {
	return &Engine{
		mux:        http.NewServeMux(),
		middleware: middleware.New[Handler](),
	}
}

// Use adds middleware that runs on every request handled by the engine.
func (e *Engine) Use(mw ...Middleware) {
	e.middleware.Use(mw...)
}

func (e *Engine) Handle(method, path string, handler Handler) {
	if path == "/" {
		// Bare "/" in ServeMux matches every otherwise-unmatched path;
		// "/{$}" matches the root exactly so unknown paths get a 404.
		path = "/{$}"
	}
	e.mux.Handle(method+" "+path, e.wrap(handler))
}

func (e *Engine) Get(path string, handler Handler) {
	e.Handle(http.MethodGet, path, handler)
}

func (e *Engine) Post(path string, handler Handler) {
	e.Handle(http.MethodPost, path, handler)
}

func (e *Engine) Put(path string, handler Handler) {
	e.Handle(http.MethodPut, path, handler)
}

func (e *Engine) Delete(path string, handler Handler) {
	e.Handle(http.MethodDelete, path, handler)
}

func (e *Engine) Patch(path string, handler Handler) {
	e.Handle(http.MethodPatch, path, handler)
}

func (e *Engine) Head(path string, handler Handler) {
	e.Handle(http.MethodHead, path, handler)
}

func (e *Engine) Options(path string, handler Handler) {
	e.Handle(http.MethodOptions, path, handler)
}

// Match registers the same handler for each supplied HTTP method.
func (e *Engine) Match(methods []string, path string, handler Handler) {
	for _, method := range uniqueMethods(methods) {
		e.Handle(method, path, handler)
	}
}

// Any registers a handler for all commonly used HTTP methods.
func (e *Engine) Any(path string, handler Handler) {
	e.Match(standardMethods, path, handler)
}

// Redirect registers a GET endpoint that redirects to another URL.
func (e *Engine) Redirect(path, destination string, status ...int) {
	e.Get(path, redirectHandler(destination, redirectStatus(status)))
}

// Group creates a route group that shares a URL prefix.
func (e *Engine) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		engine: e,
		prefix: cleanPrefix(prefix),
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.mux.ServeHTTP(w, r)
}

func (e *Engine) wrap(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w}
		ctx := &Context{Response: rw, Request: r, renderer: e.renderer}

		if err := e.middleware.Wrap(handler)(ctx); err != nil {
			log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
			if !rw.started {
				http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}
}

// responseWriter tracks whether the response has started, so the engine
// never writes a second status line after a handler has begun responding.
type responseWriter struct {
	http.ResponseWriter
	started bool
	status  int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.started {
		return
	}
	w.started = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func cleanPrefix(prefix string) string {
	if prefix == "" || prefix == "/" {
		return ""
	}
	return "/" + strings.Trim(prefix, "/")
}

var standardMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
}

func uniqueMethods(methods []string) []string {
	result := make([]string, 0, len(methods))
	seen := make(map[string]struct{}, len(methods))

	for _, method := range methods {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" {
			continue
		}
		if _, exists := seen[method]; exists {
			continue
		}
		seen[method] = struct{}{}
		result = append(result, method)
	}

	return result
}

func redirectStatus(status []int) int {
	if len(status) == 0 {
		return http.StatusFound
	}
	return status[0]
}

func redirectHandler(destination string, status int) Handler {
	return func(c *Context) error {
		http.Redirect(c.Response, c.Request, destination, status)
		return nil
	}
}
