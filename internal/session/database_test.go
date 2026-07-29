package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseSessionSurvivesManagerRestart(t *testing.T) {
	database, err := gorm.Open(sqlite.Open("file:session-restart?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrator().CreateTable(&session.DatabaseRecord{}); err != nil {
		t.Fatal(err)
	}
	store := session.NewDatabaseStore(database)

	firstEngine := Route.New()
	firstManager := session.New(session.Config{CookieName: "session"}, store)
	firstEngine.Use(firstManager.Middleware)
	firstEngine.Get("/set", func(c *Route.Context) error {
		return c.Session().Set("user_id", uint(42))
	})
	setResponse := httptest.NewRecorder()
	firstEngine.ServeHTTP(setResponse, httptest.NewRequest(http.MethodGet, "/set", nil))
	cookie := setResponse.Result().Cookies()[0]

	// A new manager represents an application restart or hot reload.
	secondEngine := Route.New()
	secondManager := session.New(session.Config{CookieName: "session"}, store)
	secondEngine.Use(secondManager.Middleware)
	secondEngine.Get("/get", func(c *Route.Context) error {
		userID, exists := c.Session().Get("user_id")
		return c.OK(Route.Data{"exists": exists, "user_id": userID})
	})
	getRequest := httptest.NewRequest(http.MethodGet, "/get", nil)
	getRequest.AddCookie(cookie)
	getResponse := httptest.NewRecorder()
	secondEngine.ServeHTTP(getResponse, getRequest)

	if getResponse.Body.String() != "{\"exists\":true,\"user_id\":42}\n" {
		t.Fatalf("session did not survive restart: %s", getResponse.Body.String())
	}
}
