package main

import (
	"context"
	"errors"
	"log/slog"
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
	"fluxgo/internal/logging"
	AuthMail "fluxgo/internal/mail"
	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
	"fluxgo/internal/view"
	Routes "fluxgo/route"
)

func main() {
	environment, err := config.Load(".env")
	if err != nil {
		slog.Error("load environment", "error", err)
		os.Exit(1)
	}

	logger := logging.New(logging.Config{Level: environment.LogLevel, Format: environment.LogFormat})
	slog.SetDefault(logger)

	db, err := database.ConnectMySQL(environment.Database)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("access database connection", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if environment.Database.RunMigrations {
		if err := database.RunMigrations(db, database.DefaultMigrations()); err != nil {
			slog.Error("run database migrations", "error", err)
			os.Exit(1)
		}
	}

	views, err := view.New(view.Config{Root: environment.ViewsRoot})
	if err != nil {
		slog.Error("boot views", "error", err)
		os.Exit(1)
	}
	Route.SetRenderer(views)

	sessionStore := session.NewDatabaseStore(db)
	if err := sessionStore.DeleteExpired(time.Now()); err != nil {
		slog.Error("clean expired sessions", "error", err)
	}
	sessions := session.New(session.Config{
		CookieName: environment.SessionCookie,
		Lifetime:   environment.SessionLifetime,
		Secure:     environment.SessionSecure,
	}, sessionStore)

	csrfProtection := csrf.New(csrf.Config{})

	Routes.Middleware(environment.SessionSecure, logger)
	loginLimiter := auth.NewLoginLimiter(db, 5, 15*time.Minute, 15*time.Minute)
	authHandlers, err := handlers.NewAuthHandler(
		db,
		environment.AppURL,
		AuthMail.LogMailer{},
		loginLimiter,
	)
	if err != nil {
		slog.Error("initialize authentication", "error", err)
		os.Exit(1)
	}
	Routes.Web(
		authHandlers,
		handlers.NewProductHandler(db),
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
		slog.Info("server starting", "app", environment.AppName, "env", environment.AppEnv, "addr", environment.ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serve HTTP", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownSignal.Done()
	slog.Info("shutting down server")
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
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
				slog.Error("clean expired sessions", "error", err)
			}
		}
	}
}
