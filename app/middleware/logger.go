// Package middleware contains middleware specific to this application.
package middleware

import (
	"log"
	"time"

	Route "fluxgo/internal/route"
)

// Logger records the method, path, response status, and request duration.
func Logger(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		started := time.Now()
		err := next(c)

		log.Printf(
			"%s %s status=%d duration=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Status(),
			time.Since(started),
		)

		return err
	}
}
