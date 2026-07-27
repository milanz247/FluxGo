package route_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	Route "fluxgo/internal/route"
)

func TestGroupedRoute(t *testing.T) {
	engine := Route.New()
	api := engine.Group("/api")

	api.Get("/users/{id}", func(c *Route.Context) error {
		return c.OK(map[string]string{"id": c.Param("id")})
	})

	request := httptest.NewRequest(http.MethodGet, "/api/users/42", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if got, want := response.Body.String(), "{\"id\":\"42\"}\n"; got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func TestRootRouteIsExactMatch(t *testing.T) {
	engine := Route.New()
	engine.Get("/", func(c *Route.Context) error {
		return c.View("home")
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected / to return %d, got %d", http.StatusOK, response.Code)
	}

	response = httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected /missing to return %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestMiddlewareOrder(t *testing.T) {
	var order []string
	record := func(name string) Route.Middleware {
		return func(next Route.Handler) Route.Handler {
			return func(c *Route.Context) error {
				order = append(order, name)
				return next(c)
			}
		}
	}

	engine := Route.New()
	engine.Use(record("engine"))

	api := engine.Group("/api").Use(record("group"))
	api.Get("/ping", func(c *Route.Context) error {
		order = append(order, "handler")
		return c.OK(map[string]string{"pong": "true"})
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/ping", nil))

	want := []string{"engine", "group", "handler"}
	if len(order) != len(want) {
		t.Fatalf("expected call order %v, got %v", want, order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("expected call order %v, got %v", want, order)
		}
	}
}

func TestErrorAfterResponseStartedKeepsStatus(t *testing.T) {
	engine := Route.New()
	engine.Get("/broken", func(c *Route.Context) error {
		if err := c.View("partial"); err != nil {
			return err
		}
		return errors.New("failure after writing")
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/broken", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected the original %d status to be kept, got %d", http.StatusOK, response.Code)
	}
	if got, want := response.Body.String(), "partial"; got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func TestErrorBeforeResponseStartedReturns500(t *testing.T) {
	engine := Route.New()
	engine.Get("/fail", func(c *Route.Context) error {
		return errors.New("boom")
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/fail", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}
