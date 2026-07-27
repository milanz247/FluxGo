package route

import (
	"net/http"
	"slices"
	"strings"
)

// RouteGroup registers routes under a shared URL prefix.
type RouteGroup struct {
	engine     *Engine
	prefix     string
	middleware []Middleware
}

// Use adds middleware that runs on every route registered on the group
// after this call. Register middleware before the routes it should wrap.
func (g *RouteGroup) Use(mw ...Middleware) *RouteGroup {
	g.middleware = append(g.middleware, mw...)
	return g
}

func (g *RouteGroup) Handle(method, path string, handler Handler) {
	for i := len(g.middleware) - 1; i >= 0; i-- {
		handler = g.middleware[i](handler)
	}
	g.engine.Handle(method, g.path(path), handler)
}

func (g *RouteGroup) Get(path string, handler Handler) {
	g.Handle(http.MethodGet, path, handler)
}

func (g *RouteGroup) Post(path string, handler Handler) {
	g.Handle(http.MethodPost, path, handler)
}

func (g *RouteGroup) Put(path string, handler Handler) {
	g.Handle(http.MethodPut, path, handler)
}

func (g *RouteGroup) Delete(path string, handler Handler) {
	g.Handle(http.MethodDelete, path, handler)
}

func (g *RouteGroup) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		engine:     g.engine,
		prefix:     g.prefix + cleanPrefix(prefix),
		middleware: slices.Clone(g.middleware),
	}
}

func (g *RouteGroup) path(path string) string {
	if path == "" || path == "/" {
		if g.prefix == "" {
			return "/"
		}
		return g.prefix
	}
	return g.prefix + "/" + strings.TrimLeft(path, "/")
}
