package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
)

func TestSessionPersistsAcrossRequests(t *testing.T) {
	manager := session.New(session.Config{
		CookieName: "test_session",
		Lifetime:   time.Hour,
	}, nil)
	engine := Route.New()
	engine.Use(manager.Middleware)

	engine.Get("/set", func(c *Route.Context) error {
		return c.Session().Set("user", "Milan")
	})
	engine.Get("/get", func(c *Route.Context) error {
		value, exists := c.Session().Get("user")
		return c.OK(Route.Data{"exists": exists, "value": value})
	})

	setResponse := httptest.NewRecorder()
	engine.ServeHTTP(setResponse, httptest.NewRequest(http.MethodGet, "/set", nil))
	cookies := setResponse.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected a session cookie, got %d", len(cookies))
	}
	if !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected secure cookie defaults, got %#v", cookies[0])
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/get", nil)
	getRequest.AddCookie(cookies[0])
	getResponse := httptest.NewRecorder()
	engine.ServeHTTP(getResponse, getRequest)

	if got := getResponse.Body.String(); got != "{\"exists\":true,\"value\":\"Milan\"}\n" {
		t.Fatalf("unexpected session response %q", got)
	}
}

func TestSessionPullRegenerateAndDestroy(t *testing.T) {
	manager := session.New(session.Config{CookieName: "session"}, nil)
	engine := Route.New()
	engine.Use(manager.Middleware)

	engine.Get("/set", func(c *Route.Context) error {
		if err := c.Session().Set("flash", "Saved"); err != nil {
			return err
		}
		return c.Text(http.StatusOK, c.Session().ID())
	})
	engine.Get("/pull", func(c *Route.Context) error {
		value, exists, err := c.Session().Pull("flash")
		if err != nil {
			return err
		}
		return c.OK(Route.Data{"exists": exists, "value": value})
	})
	engine.Get("/regenerate", func(c *Route.Context) error {
		oldID := c.Session().ID()
		if err := c.Session().Regenerate(); err != nil {
			return err
		}
		return c.OK(Route.Data{"changed": oldID != c.Session().ID()})
	})
	engine.Get("/destroy", func(c *Route.Context) error {
		return c.Session().Destroy()
	})

	setResponse := httptest.NewRecorder()
	engine.ServeHTTP(setResponse, httptest.NewRequest(http.MethodGet, "/set", nil))
	cookie := setResponse.Result().Cookies()[0]

	pullResponse := performWithCookie(engine, "/pull", cookie)
	if pullResponse.Body.String() != "{\"exists\":true,\"value\":\"Saved\"}\n" {
		t.Fatalf("unexpected pull response %q", pullResponse.Body.String())
	}
	pullAgain := performWithCookie(engine, "/pull", cookie)
	if pullAgain.Body.String() != "{\"exists\":false,\"value\":null}\n" {
		t.Fatalf("expected pulled value to be removed, got %q", pullAgain.Body.String())
	}

	regenerateResponse := performWithCookie(engine, "/regenerate", cookie)
	if regenerateResponse.Body.String() != "{\"changed\":true}\n" {
		t.Fatalf("expected session ID to change")
	}
	regeneratedCookies := regenerateResponse.Result().Cookies()
	if len(regeneratedCookies) != 1 || regeneratedCookies[0].Value == cookie.Value {
		t.Fatalf("expected a regenerated cookie")
	}

	destroyResponse := performWithCookie(engine, "/destroy", regeneratedCookies[0])
	destroyedCookies := destroyResponse.Result().Cookies()
	if len(destroyedCookies) != 1 || destroyedCookies[0].MaxAge != -1 {
		t.Fatalf("expected session cookie deletion")
	}
}

func TestMemoryStoreExpiresSessions(t *testing.T) {
	store := session.NewMemoryStore()
	now := time.Now()
	if err := store.Save("expired", map[string]any{"value": true}, now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}

	_, exists, err := store.Load("expired", now)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected expired session to be removed")
	}
}

func performWithCookie(engine *Route.Engine, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}
