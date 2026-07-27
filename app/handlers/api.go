package handlers

import Route "fluxgo/internal/route"

// Hello handles the API greeting endpoint.
func Hello(c *Route.Context) error {
	return c.OK(map[string]string{
		"message": "Hello from the API!",
	})
}

// ShowUser handles an API request for a user by ID.
func ShowUser(c *Route.Context) error {
	return c.OK(map[string]string{
		"id": c.Param("id"),
	})
}
