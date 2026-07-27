package route

import (
	"fluxgo/app/handlers"
	Route "fluxgo/internal/route"
)

// Web registers browser-facing routes.
func Web() {
	Route.Get("/", handlers.WelcomeHandler)
	Route.Post("/check-value", handlers.CheckValue)
}
