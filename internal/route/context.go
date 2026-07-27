package route

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Context contains everything a route handler needs for one request.
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	renderer Renderer
	values   map[string]any
	session  Session
	shared   Data
}

// Session provides request-scoped access to the framework session.
type Session interface {
	ID() string
	Get(string) (any, bool)
	Has(string) bool
	Set(string, any) error
	Pull(string) (any, bool, error)
	Delete(string) error
	Clear() error
	Regenerate() error
	Destroy() error
}

// JSON writes a JSON response with the provided status code.
func (c *Context) JSON(status int, value any) error {
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(status)
	return json.NewEncoder(c.Response).Encode(value)
}

// OK writes a successful JSON response.
func (c *Context) OK(value any) error {
	return c.JSON(http.StatusOK, value)
}

// Created writes a 201 JSON response.
func (c *Context) Created(value any) error {
	return c.JSON(http.StatusCreated, value)
}

// Text writes a plain-text response.
func (c *Context) Text(status int, content string) error {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(content))
	return err
}

// NoContent writes a 204 response with no body.
func (c *Context) NoContent() error {
	c.Response.WriteHeader(http.StatusNoContent)
	return nil
}

// HTML writes an HTML response with the provided status code.
func (c *Context) HTML(status int, content string) error {
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	_, err := c.Response.Write([]byte(content))
	return err
}

// View writes a successful HTML response.
func (c *Context) View(content string) error {
	return c.HTML(http.StatusOK, content)
}

// Render executes a precompiled view selected for the current request.
func (c *Context) Render(name string, data any) error {
	return c.RenderStatus(http.StatusOK, name, data)
}

// RenderStatus executes a precompiled view with an explicit response status.
func (c *Context) RenderStatus(status int, name string, data any) error {
	if c.renderer == nil {
		return fmt.Errorf("route: view renderer is not configured")
	}
	if values, ok := data.(Data); ok && len(c.shared) > 0 {
		merged := make(Data, len(c.shared)+len(values))
		for key, value := range c.shared {
			merged[key] = value
		}
		for key, value := range values {
			merged[key] = value
		}
		data = merged
	}
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(status)
	return c.renderer.Render(c.Response, c.Request, name, data)
}

// Redirect sends an HTTP redirect. The status defaults to 302 Found.
func (c *Context) Redirect(destination string, status ...int) error {
	code := http.StatusFound
	if len(status) > 0 {
		code = status[0]
	}
	http.Redirect(c.Response, c.Request, destination, code)
	return nil
}

// BindJSON decodes the request body into value.
func (c *Context) BindJSON(value any) error {
	return json.NewDecoder(c.Request.Body).Decode(value)
}

// Param returns a path parameter registered with a pattern such as /users/{id}.
func (c *Context) Param(name string) string {
	return c.Request.PathValue(name)
}

// Query returns a URL query parameter.
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}

// QueryDefault returns a URL query parameter or a fallback when it is empty.
func (c *Context) QueryDefault(name, fallback string) string {
	if value := c.Query(name); value != "" {
		return value
	}
	return fallback
}

// Form returns a parsed form value.
func (c *Context) Form(name string) string {
	return c.Request.FormValue(name)
}

// Header returns a request header.
func (c *Context) Header(name string) string {
	return c.Request.Header.Get(name)
}

// SetHeader sets a response header.
func (c *Context) SetHeader(name, value string) {
	c.Response.Header().Set(name, value)
}

// BearerToken returns the Authorization bearer token, if present.
func (c *Context) BearerToken() string {
	authorization := strings.TrimSpace(c.Header("Authorization"))
	scheme, token, found := strings.Cut(authorization, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

// Cookie returns a cookie from the request.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a cookie to the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// DeleteCookie expires a cookie at the root path.
func (c *Context) DeleteCookie(name string) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		HttpOnly: true,
	})
}

// IsHTMX reports whether this request was initiated by HTMX.
func (c *Context) IsHTMX() bool {
	return strings.EqualFold(strings.TrimSpace(c.Header("HX-Request")), "true")
}

// IsJSON reports whether the request body has a JSON content type.
func (c *Context) IsJSON() bool {
	contentType := strings.ToLower(strings.TrimSpace(c.Header("Content-Type")))
	return contentType == "application/json" || strings.HasPrefix(contentType, "application/json;")
}

// Method returns the request method.
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the request URL path.
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// Set stores a request-scoped value, typically from middleware.
func (c *Context) Set(key string, value any) {
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}

// Get retrieves a request-scoped value.
func (c *Context) Get(key string) (any, bool) {
	value, exists := c.values[key]
	return value, exists
}

// MustGet retrieves a request-scoped value or panics when it is absent.
func (c *Context) MustGet(key string) any {
	value, exists := c.Get(key)
	if !exists {
		panic(fmt.Sprintf("route: context value %q is not set", key))
	}
	return value
}

// Session returns the session attached by the session middleware.
func (c *Context) Session() Session {
	return c.session
}

// SetSession attaches a session to this request context.
// It is intended for framework middleware.
func (c *Context) SetSession(session Session) {
	c.session = session
}

// Share makes a value available to Route.Data passed to Render.
func (c *Context) Share(key string, value any) {
	if c.shared == nil {
		c.shared = make(Data)
	}
	c.shared[key] = value
}

// Shared retrieves a shared view value.
func (c *Context) Shared(key string) (any, bool) {
	value, exists := c.shared[key]
	return value, exists
}

// BadRequest writes a standard 400 JSON error response.
func (c *Context) BadRequest(message string) error {
	return c.errorJSON(http.StatusBadRequest, message)
}

// Unauthorized writes a standard 401 JSON error response.
func (c *Context) Unauthorized(message string) error {
	return c.errorJSON(http.StatusUnauthorized, message)
}

// Forbidden writes a standard 403 JSON error response.
func (c *Context) Forbidden(message string) error {
	return c.errorJSON(http.StatusForbidden, message)
}

// NotFound writes a standard 404 JSON error response.
func (c *Context) NotFound(message string) error {
	return c.errorJSON(http.StatusNotFound, message)
}

// UnprocessableEntity writes a standard 422 JSON error response.
func (c *Context) UnprocessableEntity(message string) error {
	return c.errorJSON(http.StatusUnprocessableEntity, message)
}

// ServerError writes a standard 500 JSON error response.
func (c *Context) ServerError(message string) error {
	return c.errorJSON(http.StatusInternalServerError, message)
}

func (c *Context) errorJSON(status int, message string) error {
	return c.JSON(status, Data{"error": message})
}

// Status returns the response status code. It returns 200 before a response
// explicitly writes another status.
func (c *Context) Status() int {
	if writer, ok := c.Response.(*responseWriter); ok && writer.status != 0 {
		return writer.status
	}
	return http.StatusOK
}
