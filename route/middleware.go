package route

import (
	"log/slog"

	AppMiddleware "fluxgo/app/middleware"
	"fluxgo/internal/logging"
	Route "fluxgo/internal/route"
)

// Middleware registers application-wide middleware.
func Middleware(hsts bool, logger *slog.Logger) {
	Route.Use(AppMiddleware.SecurityHeaders(hsts), logging.Middleware(logger))
}
