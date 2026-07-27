package route

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Context contains everything a route handler needs for one request.
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	renderer Renderer
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
	if c.renderer == nil {
		return fmt.Errorf("route: view renderer is not configured")
	}
	return c.renderer.Render(c.Response, c.Request, name, data)
}

// BindJSON decodes the request body into value.
func (c *Context) BindJSON(value any) error {
	return json.NewDecoder(c.Request.Body).Decode(value)
}

// Param returns a path parameter registered with a pattern such as /users/{id}.
func (c *Context) Param(name string) string {
	return c.Request.PathValue(name)
}

// Status returns the response status code. It returns 200 before a response
// explicitly writes another status.
func (c *Context) Status() int {
	if writer, ok := c.Response.(*responseWriter); ok && writer.status != 0 {
		return writer.status
	}
	return http.StatusOK
}
