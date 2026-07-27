package route

import (
	"encoding/json"
	"net/http"
)

// Context contains everything a route handler needs for one request.
type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
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

// BindJSON decodes the request body into value.
func (c *Context) BindJSON(value any) error {
	return json.NewDecoder(c.Request.Body).Decode(value)
}

// Param returns a path parameter registered with a pattern such as /users/{id}.
func (c *Context) Param(name string) string {
	return c.Request.PathValue(name)
}
