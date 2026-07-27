package route

import (
	"fluxgo/app/handlers"
	Route "fluxgo/internal/route"
)

// Web registers browser-facing routes.
func Web(users *handlers.UserHandler) {
	Route.Get("/", handlers.WelcomeHandler)
	Route.Post("/check-value", handlers.CheckValue)

	Route.Get("/users", users.Index)
	Route.Post("/users", users.Store)
	Route.Get("/users/{id}", users.Show)
	Route.Get("/users/{id}/edit", users.Edit)
	Route.Post("/users/{id}/update", users.Update)
	Route.Post("/users/{id}/delete", users.Delete)
}
