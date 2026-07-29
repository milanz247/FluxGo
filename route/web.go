package route

import (
	"fluxgo/app/handlers"
	"fluxgo/app/middleware"
	Route "fluxgo/internal/route"
)

// Web registers browser-facing routes.
func Web(
	auth *handlers.AuthHandler,
	product *handlers.ProductHandler,
	health Route.Handler,
	sessionMiddleware Route.Middleware,
	csrfMiddleware Route.Middleware,
) {
	Route.Get("/health", health)

	web := Route.Group("").Use(sessionMiddleware, csrfMiddleware)
	web.Get("/", auth.Home)
	web.Get("/email/verify", auth.VerifyEmail)

	guest := web.Group("").Use(middleware.Guest)
	guest.Get("/register", auth.ShowRegister)
	guest.Post("/register", auth.Register)
	guest.Get("/login", auth.ShowLogin)
	guest.Post("/login", auth.Login)
	guest.Get("/forgot-password", auth.ShowForgotPassword)
	guest.Post("/forgot-password", auth.ForgotPassword)
	guest.Get("/reset-password", auth.ShowResetPassword)
	guest.Post("/reset-password", auth.ResetPassword)

	protected := web.Group("").Use(middleware.Auth)
	protected.Get("/dashboard", auth.Dashboard)
	protected.Post("/logout", auth.Logout)
	protected.Post("/email/verification-notification", auth.ResendVerification)

	products := protected.Group("/products")
	products.Get("", product.Index)
	products.Get("/create", product.ShowCreate)
	products.Post("", product.Store)
	products.Get("/{id}/edit", product.ShowEdit)
	products.Post("/{id}", product.Update)
	products.Post("/{id}/delete", product.Delete)
}
