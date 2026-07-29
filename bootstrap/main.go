package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fluxgo/app/handlers"
	"fluxgo/config"
	"fluxgo/internal/auth"
	"fluxgo/internal/csrf"
	"fluxgo/internal/database"
	AuthMail "fluxgo/internal/mail"
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

	if environment.Database.RunMigrations {
		if err := database.RunMigrations(db, database.DefaultMigrations()); err != nil {
			log.Fatalf("run database migrations: %v", err)
		}
	}

	views, err := view.New(view.Config{Root: environment.ViewsRoot})
	if err != nil {
		log.Fatalf("boot views: %v", err)
	}
	Route.SetRenderer(views)

	sessionStore := session.NewDatabaseStore(db)
	if err := sessionStore.DeleteExpired(time.Now()); err != nil {
		log.Printf("clean expired sessions: %v", err)
	}
	sessions := session.New(session.Config{
		CookieName: environment.SessionCookie,
		Lifetime:   environment.SessionLifetime,
		Secure:     environment.SessionSecure,
	}, sessionStore)

	csrfProtection := csrf.New(csrf.Config{})

	Routes.Middleware(environment.SessionSecure)
	loginLimiter := auth.NewLoginLimiter(db, 5, 15*time.Minute, 15*time.Minute)
	authHandlers, err := handlers.NewAuthHandler(
		db,
		environment.AppURL,
		AuthMail.LogMailer{},
		loginLimiter,
	)
	if err != nil {
		log.Fatalf("initialize authentication: %v", err)
	}
	Routes.Web(
		authHandlers,
		handlers.Health(db),
		sessions.Middleware,
		csrfProtection.Middleware,
	)

	server := &http.Server{
		Addr:              environment.ServerAddr,
		Handler:           Route.HTTPHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go cleanupExpiredSessions(shutdownSignal, sessionStore)

	go func() {
		log.Printf("%s running in %s at %s", environment.AppName, environment.AppEnv, environment.ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve HTTP: %v", err)
		}
	}()

	<-shutdownSignal.Done()
	log.Print("shutting down server")
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func cleanupExpiredSessions(ctx context.Context, store *session.DatabaseStore) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := store.DeleteExpired(now); err != nil {
				log.Printf("clean expired sessions: %v", err)
			}
		}
	}
}
