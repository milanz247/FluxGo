package main

import (
	"log"
	"net/http"

	"fluxgo/app/handlers"
	"fluxgo/app/models"
	"fluxgo/config"
	"fluxgo/internal/csrf"
	"fluxgo/internal/database"
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

	db, err := database.ConnectMySQL(environment.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("access database connection: %v", err)
	}
	defer sqlDB.Close()

	if environment.Database.AutoMigrate {
		if err := database.Migrate(db, &models.User{}); err != nil {
			log.Fatalf("migrate database: %v", err)
		}
		if err := database.DropColumnIfExists(db, &models.User{}, "phone"); err != nil {
			log.Fatalf("clean legacy user schema: %v", err)
		}
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

	csrfProtection := csrf.New(csrf.Config{})
	Route.Use(csrfProtection.Middleware)

	Routes.Middleware()
	authHandlers, err := handlers.NewAuthHandler(db)
	if err != nil {
		log.Fatalf("initialize authentication: %v", err)
	}
	Routes.Web(authHandlers)

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
