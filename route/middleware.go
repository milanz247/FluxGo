package route

import (
	AppMiddleware "fluxgo/app/middleware"
	Route "fluxgo/internal/route"
)

// Middleware registers application-wide middleware.
func Middleware() {
	Route.Use(AppMiddleware.Logger)
}
