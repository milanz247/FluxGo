package logging

import (
	"log/slog"
	"net/http"
	"time"

	Route "fluxgo/internal/route"
)

// Middleware records the method, path, response status, and duration of
// every request as a structured log entry. A nil logger uses slog.Default(),
// resolved per request so it reflects whatever logger is current when the
// request is handled.
func Middleware(logger *slog.Logger) Route.Middleware {
	return func(next Route.Handler) Route.Handler {
		return func(c *Route.Context) error {
			active := logger
			if active == nil {
				active = slog.Default()
			}

			started := time.Now()
			err := next(c)
			status := c.Status()
			if err != nil && status == http.StatusOK {
				status = Route.StatusForError(err)
			}

			attrs := []any{
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration", time.Since(started),
			}
			if err != nil {
				attrs = append(attrs, "error", err)
				active.Error("request failed", attrs...)
			} else {
				active.Info("request handled", attrs...)
			}

			return err
		}
	}
}
