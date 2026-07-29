package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	AppMiddleware "fluxgo/app/middleware"
	Route "fluxgo/internal/route"
)

func TestSecurityHeaders(t *testing.T) {
	engine := Route.New()
	engine.Use(AppMiddleware.SecurityHeaders(true))
	engine.Get("/", func(c *Route.Context) error {
		return c.NoContent()
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, header := range []string{
		"Content-Security-Policy",
		"Referrer-Policy",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Permissions-Policy",
		"Strict-Transport-Security",
	} {
		if response.Header().Get(header) == "" {
			t.Fatalf("expected %s header", header)
		}
	}
}
