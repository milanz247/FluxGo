package route_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	Route "fluxgo/internal/route"
)

func TestRequestContextHelpers(t *testing.T) {
	engine := Route.New()
	engine.Post("/users/{id}", func(c *Route.Context) error {
		if c.Param("id") != "42" {
			t.Fatalf("unexpected path parameter %q", c.Param("id"))
		}
		if c.QueryDefault("page", "1") != "2" {
			t.Fatalf("unexpected query value")
		}
		if c.Form("name") != "Milan" {
			t.Fatalf("unexpected form value")
		}
		if c.BearerToken() != "secret" {
			t.Fatalf("unexpected bearer token %q", c.BearerToken())
		}
		if !c.IsHTMX() || c.IsJSON() {
			t.Fatalf("unexpected request detection result")
		}
		if c.Method() != http.MethodPost || c.Path() != "/users/42" {
			t.Fatalf("unexpected method or path")
		}

		c.Set("user", "Milan")
		if value, exists := c.Get("user"); !exists || value != "Milan" {
			t.Fatalf("unexpected context value")
		}
		if c.MustGet("user") != "Milan" {
			t.Fatalf("unexpected required context value")
		}

		return c.NoContent()
	})

	body := strings.NewReader("name=Milan")
	request := httptest.NewRequest(http.MethodPost, "/users/42?page=2", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "bearer secret")
	request.Header.Set("HX-Request", "true")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent || response.Body.Len() != 0 {
		t.Fatalf("unexpected response: status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestResponseContextHelpers(t *testing.T) {
	engine := Route.New()
	engine.Get("/text", func(c *Route.Context) error {
		c.SetHeader("X-FluxGo", "true")
		c.SetCookie(&http.Cookie{Name: "session", Value: "token"})
		return c.Text(http.StatusAccepted, "accepted")
	})
	engine.Get("/redirect", func(c *Route.Context) error {
		return c.Redirect("/text", http.StatusTemporaryRedirect)
	})
	engine.Get("/error", func(c *Route.Context) error {
		return c.UnprocessableEntity("invalid input")
	})

	textResponse := httptest.NewRecorder()
	engine.ServeHTTP(textResponse, httptest.NewRequest(http.MethodGet, "/text", nil))
	if textResponse.Code != http.StatusAccepted || textResponse.Body.String() != "accepted" {
		t.Fatalf("unexpected text response")
	}
	if textResponse.Header().Get("X-FluxGo") != "true" || len(textResponse.Result().Cookies()) != 1 {
		t.Fatalf("expected response header and cookie")
	}

	redirectResponse := httptest.NewRecorder()
	engine.ServeHTTP(redirectResponse, httptest.NewRequest(http.MethodGet, "/redirect", nil))
	if redirectResponse.Code != http.StatusTemporaryRedirect || redirectResponse.Header().Get("Location") != "/text" {
		t.Fatalf("unexpected redirect response")
	}

	errorResponse := httptest.NewRecorder()
	engine.ServeHTTP(errorResponse, httptest.NewRequest(http.MethodGet, "/error", nil))
	if errorResponse.Code != http.StatusUnprocessableEntity ||
		errorResponse.Body.String() != "{\"error\":\"invalid input\"}\n" {
		t.Fatalf("unexpected error response: %s", errorResponse.Body.String())
	}
}

func TestJSONAndCookieRequestDetection(t *testing.T) {
	engine := Route.New()
	engine.Post("/json", func(c *Route.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return err
		}
		if !c.IsJSON() || cookie.Value != "token" {
			t.Fatalf("expected JSON request and session cookie")
		}
		c.DeleteCookie("session")
		return c.OK(Route.Data{"ok": true})
	})

	request := httptest.NewRequest(http.MethodPost, "/json", nil)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.AddCookie(&http.Cookie{Name: "session", Value: "token"})
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Fatalf("expected cookie deletion response")
	}
}
