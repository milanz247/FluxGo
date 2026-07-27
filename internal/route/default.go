package route

import "net/http"

var defaultEngine = New()

// Get registers a GET route on the default engine.
func Get(path string, handler Handler) {
	defaultEngine.Get(path, handler)
}

// Post registers a POST route on the default engine.
func Post(path string, handler Handler) {
	defaultEngine.Post(path, handler)
}

// Put registers a PUT route on the default engine.
func Put(path string, handler Handler) {
	defaultEngine.Put(path, handler)
}

// Delete registers a DELETE route on the default engine.
func Delete(path string, handler Handler) {
	defaultEngine.Delete(path, handler)
}

// Use adds middleware to the default engine.
func Use(mw ...Middleware) {
	defaultEngine.Use(mw...)
}

// Group creates a route group on the default engine.
func Group(prefix string) *RouteGroup {
	return defaultEngine.Group(prefix)
}

// HTTPHandler returns the default engine as a standard HTTP handler.
func HTTPHandler() http.Handler {
	return defaultEngine
}
