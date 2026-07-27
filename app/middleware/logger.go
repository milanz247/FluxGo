// Package middleware contains middleware specific to this application.
package middleware

import (
	"log"
	"net/http"
	"time"

	Route "fluxgo/internal/route"
)

// Logger records the method, path, response status, and request duration.
func Logger(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		started := time.Now()
		err := next(c)
		status := c.Status()
		if err != nil && status == http.StatusOK {
			status = http.StatusInternalServerError
		}

		log.Printf(
			"%s %s status=%d duration=%s",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			time.Since(started),
		)

		return err
	}
}
