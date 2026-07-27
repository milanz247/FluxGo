package route

import "net/http"

var defaultEngine = New()

// SetRenderer configures the view renderer on the default engine.
func SetRenderer(renderer Renderer) {
	defaultEngine.SetRenderer(renderer)
}

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

// Patch registers a PATCH route on the default engine.
func Patch(path string, handler Handler) {
	defaultEngine.Patch(path, handler)
}

// Head registers a HEAD route on the default engine.
func Head(path string, handler Handler) {
	defaultEngine.Head(path, handler)
}

// Options registers an OPTIONS route on the default engine.
func Options(path string, handler Handler) {
	defaultEngine.Options(path, handler)
}

// Match registers a route for each supplied HTTP method.
func Match(methods []string, path string, handler Handler) {
	defaultEngine.Match(methods, path, handler)
}

// Any registers a route for all commonly used HTTP methods.
func Any(path string, handler Handler) {
	defaultEngine.Any(path, handler)
}

// Redirect registers a GET endpoint that redirects to another URL.
func Redirect(path, destination string, status ...int) {
	defaultEngine.Redirect(path, destination, status...)
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
