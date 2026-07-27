package main

import (
	"log"
	"net/http"

	"fluxgo/config"
	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
	"fluxgo/internal/view"
	Routes "fluxgo/route"
)

func main() {
	environment, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load environment: %v", err)
	}

	views, err := view.New(view.Config{Root: environment.ViewsRoot})
	if err != nil {
		log.Fatalf("boot views: %v", err)
	}
	Route.SetRenderer(views)

	sessions := session.New(session.Config{
		CookieName: environment.SessionCookie,
		Lifetime:   environment.SessionLifetime,
		Secure:     environment.SessionSecure,
	}, nil)
	Route.Use(sessions.Middleware)

	Routes.Middleware()
	Routes.Web()
	Routes.API()

	log.Printf(
		"%s running in %s at %s",
		environment.AppName,
		environment.AppEnv,
		environment.ServerAddr,
	)
	if err := http.ListenAndServe(environment.ServerAddr, Route.HTTPHandler()); err != nil {
		log.Fatal(err)
	}
}
