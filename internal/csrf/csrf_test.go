package csrf_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"fluxgo/internal/csrf"
	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
)

func TestCSRFProtectsUnsafeRequests(t *testing.T) {
	engine := protectedEngine(csrf.Config{})
	engine.Get("/form", tokenHandler)
	engine.Post("/submit", func(c *Route.Context) error {
		return c.NoContent()
	})

	token, cookie := fetchToken(t, engine)

	validForm := url.Values{"_token": {token}}
	validRequest := httptest.NewRequest(
		http.MethodPost,
		"/submit",
		strings.NewReader(validForm.Encode()),
	)
	validRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	validRequest.AddCookie(cookie)
	validResponse := httptest.NewRecorder()
	engine.ServeHTTP(validResponse, validRequest)
	if validResponse.Code != http.StatusNoContent {
		t.Fatalf("expected valid token status %d, got %d", http.StatusNoContent, validResponse.Code)
	}

	invalidRequest := httptest.NewRequest(http.MethodPost, "/submit", nil)
	invalidRequest.AddCookie(cookie)
	invalidResponse := httptest.NewRecorder()
	engine.ServeHTTP(invalidResponse, invalidRequest)
	if invalidResponse.Code != http.StatusForbidden {
		t.Fatalf("expected missing token status %d, got %d", http.StatusForbidden, invalidResponse.Code)
	}
}

func TestCSRFSupportsRequestHeader(t *testing.T) {
	engine := protectedEngine(csrf.Config{})
	engine.Get("/form", tokenHandler)
	engine.Patch("/update", func(c *Route.Context) error {
		return c.NoContent()
	})

	token, cookie := fetchToken(t, engine)
	request := httptest.NewRequest(http.MethodPatch, "/update", nil)
	request.Header.Set("X-CSRF-Token", token)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected header token to pass, got %d", response.Code)
	}
}

func TestCSRFExcludedPaths(t *testing.T) {
	engine := protectedEngine(csrf.Config{Except: []string{"/webhooks/*"}})
	engine.Post("/webhooks/payment", func(c *Route.Context) error {
		return c.NoContent()
	})

	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/webhooks/payment", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected excluded path to pass, got %d", response.Code)
	}
}

func TestCSRFTokenChangesAfterSessionRegeneration(t *testing.T) {
	engine := protectedEngine(csrf.Config{})
	engine.Get("/form", tokenHandler)
	engine.Get("/regenerate", func(c *Route.Context) error {
		if err := c.Session().Regenerate(); err != nil {
			return err
		}
		return c.NoContent()
	})

	firstToken, cookie := fetchToken(t, engine)
	regenerateRequest := httptest.NewRequest(http.MethodGet, "/regenerate", nil)
	regenerateRequest.AddCookie(cookie)
	regenerateResponse := httptest.NewRecorder()
	engine.ServeHTTP(regenerateResponse, regenerateRequest)
	regeneratedCookie := regenerateResponse.Result().Cookies()[0]

	tokenRequest := httptest.NewRequest(http.MethodGet, "/form", nil)
	tokenRequest.AddCookie(regeneratedCookie)
	tokenResponse := httptest.NewRecorder()
	engine.ServeHTTP(tokenResponse, tokenRequest)
	secondToken := tokenResponse.Body.String()

	if firstToken == secondToken {
		t.Fatal("expected CSRF token to change after session regeneration")
	}
}

func protectedEngine(config csrf.Config) *Route.Engine {
	engine := Route.New()
	sessions := session.New(session.Config{CookieName: "test_session"}, nil)
	protection := csrf.New(config)
	engine.Use(sessions.Middleware, protection.Middleware)
	return engine
}

func tokenHandler(c *Route.Context) error {
	token, exists := c.Shared("CSRFToken")
	if !exists {
		return c.ServerError("token not shared")
	}
	return c.Text(http.StatusOK, token.(string))
}

func fetchToken(t *testing.T, engine *Route.Engine) (string, *http.Cookie) {
	t.Helper()
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/form", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("fetch token returned %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected session cookie")
	}
	return response.Body.String(), cookies[0]
}
