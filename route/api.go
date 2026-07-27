package route

import (
	"fluxgo/app/handlers"
	Route "fluxgo/internal/route"
)

// API registers all routes under the /api prefix.
func API() {
	api := Route.Group("/api")

	api.Get("/hello", handlers.Hello)
	api.Get("/users/{id}", handlers.ShowUser)
}
