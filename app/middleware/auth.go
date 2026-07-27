package middleware

import (
	"net/http"

	Route "fluxgo/internal/route"
)

// Auth requires a logged-in user.
func Auth(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		if c.Session() == nil || !c.Session().Has("user_id") {
			return c.Redirect("/login", http.StatusSeeOther)
		}
		return next(c)
	}
}

// Guest keeps authenticated users out of login and registration pages.
func Guest(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		if c.Session() != nil && c.Session().Has("user_id") {
			return c.Redirect("/dashboard", http.StatusSeeOther)
		}
		return next(c)
	}
}
