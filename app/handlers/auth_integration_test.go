package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"fluxgo/app/handlers"
	AppMiddleware "fluxgo/app/middleware"
	"fluxgo/internal/auth"
	"fluxgo/internal/csrf"
	"fluxgo/internal/database"
	Route "fluxgo/internal/route"
	"fluxgo/internal/session"
	"fluxgo/internal/view"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var csrfPattern = regexp.MustCompile(`name="_token" value="([^"]+)"`)

type captureMailer struct {
	verificationLink string
	resetLink        string
}

func (mailer *captureMailer) SendVerification(_ string, link string) error {
	mailer.verificationLink = link
	return nil
}

func (mailer *captureMailer) SendPasswordReset(_ string, link string) error {
	mailer.resetLink = link
	return nil
}

type authApplication struct {
	engine   *Route.Engine
	mailer   *captureMailer
	database *gorm.DB
}

func TestRegisterVerifyLogoutAndLoginFlow(t *testing.T) {
	app := newAuthApplication(t)
	var cookie *http.Cookie

	registerPage := app.request(http.MethodGet, "/register", nil, cookie)
	cookie = responseCookie(registerPage, cookie)
	csrfToken := extractCSRF(t, registerPage.Body.String())

	register := app.request(http.MethodPost, "/register", url.Values{
		"_token":                {csrfToken},
		"name":                  {"Milan"},
		"email":                 {"milan@example.com"},
		"password":              {"strong-password"},
		"password_confirmation": {"strong-password"},
	}, cookie)
	if register.Code != http.StatusSeeOther || register.Header().Get("Location") != "/dashboard" {
		t.Fatalf("unexpected register response: %d %s", register.Code, register.Body.String())
	}
	cookie = responseCookie(register, cookie)
	if app.mailer.verificationLink == "" {
		t.Fatal("expected a verification email")
	}

	dashboard := app.request(http.MethodGet, "/dashboard", nil, cookie)
	if dashboard.Code != http.StatusOK || !strings.Contains(dashboard.Body.String(), "milan@example.com") {
		t.Fatalf("unexpected dashboard: %d %s", dashboard.Code, dashboard.Body.String())
	}
	if !strings.Contains(dashboard.Body.String(), "Email verified:</strong> No") {
		t.Fatal("expected unverified account status")
	}

	verificationURL, err := url.Parse(app.mailer.verificationLink)
	if err != nil {
		t.Fatal(err)
	}
	verify := app.request(http.MethodGet, verificationURL.RequestURI(), nil, cookie)
	if verify.Code != http.StatusSeeOther {
		t.Fatalf("unexpected verification response: %d", verify.Code)
	}

	verifiedDashboard := app.request(http.MethodGet, "/dashboard", nil, cookie)
	if !strings.Contains(verifiedDashboard.Body.String(), "Email verified:</strong> Yes") {
		t.Fatal("expected verified account status")
	}
	logoutToken := extractCSRF(t, verifiedDashboard.Body.String())
	logout := app.request(http.MethodPost, "/logout", url.Values{"_token": {logoutToken}}, cookie)
	if logout.Code != http.StatusSeeOther || logout.Header().Get("Location") != "/login" {
		t.Fatalf("unexpected logout response: %d", logout.Code)
	}

	loginPage := app.request(http.MethodGet, "/login", nil, nil)
	loginCookie := responseCookie(loginPage, nil)
	login := app.request(http.MethodPost, "/login", url.Values{
		"_token":   {extractCSRF(t, loginPage.Body.String())},
		"email":    {"milan@example.com"},
		"password": {"strong-password"},
	}, loginCookie)
	if login.Code != http.StatusSeeOther || login.Header().Get("Location") != "/dashboard" {
		t.Fatalf("unexpected login response: %d %s", login.Code, login.Body.String())
	}
}

func TestForgotAndResetPasswordFlow(t *testing.T) {
	app := newAuthApplication(t)
	cookie := registerUser(t, app, "reset@example.com", "old-password")

	// Destroy the authenticated session before using guest-only reset routes.
	dashboard := app.request(http.MethodGet, "/dashboard", nil, cookie)
	logout := app.request(http.MethodPost, "/logout", url.Values{
		"_token": {extractCSRF(t, dashboard.Body.String())},
	}, cookie)
	if logout.Code != http.StatusSeeOther {
		t.Fatalf("logout failed: %d", logout.Code)
	}

	forgotPage := app.request(http.MethodGet, "/forgot-password", nil, nil)
	guestCookie := responseCookie(forgotPage, nil)
	forgot := app.request(http.MethodPost, "/forgot-password", url.Values{
		"_token": {extractCSRF(t, forgotPage.Body.String())},
		"email":  {"reset@example.com"},
	}, guestCookie)
	if forgot.Code != http.StatusOK || app.mailer.resetLink == "" {
		t.Fatalf("password reset request failed: %d", forgot.Code)
	}

	resetURL, err := url.Parse(app.mailer.resetLink)
	if err != nil {
		t.Fatal(err)
	}
	resetPage := app.request(http.MethodGet, resetURL.RequestURI(), nil, guestCookie)
	reset := app.request(http.MethodPost, "/reset-password", url.Values{
		"_token":                {extractCSRF(t, resetPage.Body.String())},
		"token":                 {resetURL.Query().Get("token")},
		"password":              {"new-password"},
		"password_confirmation": {"new-password"},
	}, guestCookie)
	if reset.Code != http.StatusSeeOther || reset.Header().Get("Location") != "/login?status=password-reset" {
		t.Fatalf("password reset failed: %d %s", reset.Code, reset.Body.String())
	}
}

func TestSecurityHeadersAndHealth(t *testing.T) {
	app := newAuthApplication(t)
	response := app.request(http.MethodGet, "/health", nil, nil)
	if response.Code != http.StatusOK || response.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("expected healthy response with security headers: %d", response.Code)
	}
}

func newAuthApplication(t *testing.T) *authApplication {
	t.Helper()
	connection, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.RunMigrations(connection, database.DefaultMigrations()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	// Migrations are idempotent.
	if err := database.RunMigrations(connection, database.DefaultMigrations()); err != nil {
		t.Fatalf("rerun migrations: %v", err)
	}

	views, err := view.New(view.Config{Root: "../../views"})
	if err != nil {
		t.Fatal(err)
	}
	engine := Route.New()
	engine.SetRenderer(views)
	sessions := session.New(session.Config{CookieName: "test_session"}, session.NewDatabaseStore(connection))
	protection := csrf.New(csrf.Config{})
	engine.Use(sessions.Middleware, protection.Middleware, AppMiddleware.SecurityHeaders(false))

	mailer := &captureMailer{}
	limiter := auth.NewLoginLimiter(connection, 5, 15*time.Minute, 15*time.Minute)
	authHandler, err := handlers.NewAuthHandler(connection, "http://example.test", mailer, limiter)
	if err != nil {
		t.Fatal(err)
	}
	engine.Get("/register", authHandler.ShowRegister)
	engine.Post("/register", authHandler.Register)
	engine.Get("/login", authHandler.ShowLogin)
	engine.Post("/login", authHandler.Login)
	engine.Get("/forgot-password", authHandler.ShowForgotPassword)
	engine.Post("/forgot-password", authHandler.ForgotPassword)
	engine.Get("/reset-password", authHandler.ShowResetPassword)
	engine.Post("/reset-password", authHandler.ResetPassword)
	engine.Get("/email/verify", authHandler.VerifyEmail)
	engine.Get("/dashboard", authHandler.Dashboard)
	engine.Post("/logout", authHandler.Logout)
	engine.Get("/health", handlers.Health(connection))
	return &authApplication{engine: engine, mailer: mailer, database: connection}
}

func registerUser(t *testing.T, app *authApplication, email, password string) *http.Cookie {
	t.Helper()
	page := app.request(http.MethodGet, "/register", nil, nil)
	cookie := responseCookie(page, nil)
	response := app.request(http.MethodPost, "/register", url.Values{
		"_token":                {extractCSRF(t, page.Body.String())},
		"name":                  {"Test User"},
		"email":                 {email},
		"password":              {password},
		"password_confirmation": {password},
	}, cookie)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("register user failed: %d %s", response.Code, response.Body.String())
	}
	return responseCookie(response, cookie)
}

func (app *authApplication) request(
	method string,
	path string,
	form url.Values,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	var body *strings.Reader
	if form == nil {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(form.Encode())
	}
	request := httptest.NewRequest(method, path, body)
	if form != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	app.engine.ServeHTTP(response, request)
	return response
}

func extractCSRF(t *testing.T, body string) string {
	t.Helper()
	match := csrfPattern.FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("CSRF field missing from response: %s", body)
	}
	return match[1]
}

func responseCookie(response *httptest.ResponseRecorder, fallback *http.Cookie) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.MaxAge >= 0 && cookie.Value != "" {
			return cookie
		}
	}
	return fallback
}
