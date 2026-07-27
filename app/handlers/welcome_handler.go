package handlers

import Route "fluxgo/internal/route"

// WelcomeHandler handles the application welcome page.
func WelcomeHandler(c *Route.Context) error {
	return c.Render("home", map[string]string{
		"Title":   "FluxGo",
		"Heading": "Hello from FluxGo!",
	})
}

// CheckValue receives the form value and returns an HTMX fragment.
func CheckValue(c *Route.Context) error {
	return c.Render("value-result", map[string]any{
		"Value":  c.Request.FormValue("value"),
		"Result": true,
	})
}
