package route

import (
	"fluxgo/app/handlers"
	"fluxgo/app/middleware"
	Route "fluxgo/internal/route"
)

// Web registers browser-facing routes.
func Web(auth *handlers.AuthHandler) {
	Route.Get("/", auth.Home)

	guest := Route.Group("").Use(middleware.Guest)
	guest.Get("/register", auth.ShowRegister)
	guest.Post("/register", auth.Register)
	guest.Get("/login", auth.ShowLogin)
	guest.Post("/login", auth.Login)

	protected := Route.Group("").Use(middleware.Auth)
	protected.Get("/dashboard", auth.Dashboard)
	protected.Post("/logout", auth.Logout)
}
