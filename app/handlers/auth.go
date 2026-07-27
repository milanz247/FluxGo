package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"fluxgo/app/models"
	"fluxgo/internal/auth"
	Route "fluxgo/internal/route"
	"gorm.io/gorm"
)

const (
	minimumPasswordLength = 8
	maximumPasswordLength = 1024
)

// AuthHandler handles registration, login, logout, and the dashboard.
type AuthHandler struct {
	database          *gorm.DB
	dummyPasswordHash string
}

// NewAuthHandler creates authentication handlers using GORM directly.
func NewAuthHandler(database *gorm.DB) (*AuthHandler, error) {
	dummyHash, err := auth.HashPassword("not-a-real-user-password")
	if err != nil {
		return nil, err
	}
	return &AuthHandler{database: database, dummyPasswordHash: dummyHash}, nil
}

// Home redirects visitors to the appropriate entry page.
func (handler *AuthHandler) Home(c *Route.Context) error {
	if c.Session().Has("user_id") {
		return c.Redirect("/dashboard", http.StatusSeeOther)
	}
	return c.Redirect("/login", http.StatusSeeOther)
}

// ShowRegister displays the registration form.
func (handler *AuthHandler) ShowRegister(c *Route.Context) error {
	return c.Render("register", Route.Data{"Title": "Create account"})
}

// Register validates and creates an account.
func (handler *AuthHandler) Register(c *Route.Context) error {
	name := strings.TrimSpace(c.Form("name"))
	email := normalizeEmail(c.Form("email"))
	password := c.Form("password")
	confirmation := c.Form("password_confirmation")

	data := Route.Data{
		"Title":    "Create account",
		"OldName":  name,
		"OldEmail": email,
	}
	if name == "" || !validEmail(email) {
		data["Error"] = "Enter a valid name and email address."
		return c.RenderStatus(http.StatusUnprocessableEntity, "register", data)
	}
	if len(password) < minimumPasswordLength || len(password) > maximumPasswordLength {
		data["Error"] = "Password must be between 8 and 1024 characters."
		return c.RenderStatus(http.StatusUnprocessableEntity, "register", data)
	}
	if password != confirmation {
		data["Error"] = "Password confirmation does not match."
		return c.RenderStatus(http.StatusUnprocessableEntity, "register", data)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user := models.User{Name: name, Email: email, PasswordHash: passwordHash}
	if err := handler.database.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			data["Error"] = "An account already exists for that email address."
			return c.RenderStatus(http.StatusUnprocessableEntity, "register", data)
		}
		return err
	}

	if err := c.Session().Regenerate(); err != nil {
		return err
	}
	if err := c.Session().Set("user_id", user.ID); err != nil {
		return err
	}
	return c.Redirect("/dashboard", http.StatusSeeOther)
}

// ShowLogin displays the login form.
func (handler *AuthHandler) ShowLogin(c *Route.Context) error {
	return c.Render("login", Route.Data{"Title": "Sign in"})
}

// Login authenticates an account and regenerates its session.
func (handler *AuthHandler) Login(c *Route.Context) error {
	email := normalizeEmail(c.Form("email"))
	password := c.Form("password")
	data := Route.Data{"Title": "Sign in", "OldEmail": email}

	var user models.User
	err := handler.database.WithContext(c.Request.Context()).
		Where("email = ?", email).
		First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	passwordHash := user.PasswordHash
	if err != nil {
		passwordHash = handler.dummyPasswordHash
	}
	if !auth.VerifyPassword(password, passwordHash) || err != nil {
		data["Error"] = "The email address or password is incorrect."
		return c.RenderStatus(http.StatusUnprocessableEntity, "login", data)
	}

	if err := c.Session().Regenerate(); err != nil {
		return err
	}
	if err := c.Session().Set("user_id", user.ID); err != nil {
		return err
	}
	return c.Redirect("/dashboard", http.StatusSeeOther)
}

// Dashboard displays the authenticated user's dashboard.
func (handler *AuthHandler) Dashboard(c *Route.Context) error {
	userID, _ := c.Session().Get("user_id")
	var user models.User
	if err := handler.database.WithContext(c.Request.Context()).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = c.Session().Destroy()
			return c.Redirect("/login", http.StatusSeeOther)
		}
		return err
	}
	return c.Render("dashboard", Route.Data{
		"Title": "Dashboard",
		"User":  user,
	})
}

// Logout destroys the authenticated session.
func (handler *AuthHandler) Logout(c *Route.Context) error {
	if err := c.Session().Destroy(); err != nil {
		return err
	}
	return c.Redirect("/login", http.StatusSeeOther)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
