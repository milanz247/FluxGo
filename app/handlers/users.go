package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"fluxgo/app/models"
	"fluxgo/app/store"
	Route "fluxgo/internal/route"
)

// UserRepository defines the persistence required by user handlers.
type UserRepository interface {
	All(context.Context) ([]models.User, error)
	Find(context.Context, uint) (models.User, error)
	Create(context.Context, string, string) (models.User, error)
	Update(context.Context, uint, string, string) (models.User, error)
	Delete(context.Context, uint) error
}

// UserHandler handles user CRUD requests.
type UserHandler struct {
	users UserRepository
}

// NewUserHandler creates user CRUD handlers.
func NewUserHandler(users UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

// Index displays the user table and create form.
func (handler *UserHandler) Index(c *Route.Context) error {
	users, err := handler.users.All(c.Request.Context())
	if err != nil {
		return err
	}
	return c.Render("users", Route.Data{
		"Title": "Users",
		"Users": users,
	})
}

// Store creates a user from submitted form values.
func (handler *UserHandler) Store(c *Route.Context) error {
	name := strings.TrimSpace(c.Form("name"))
	phone := strings.TrimSpace(c.Form("phone"))
	if name == "" || phone == "" {
		return c.UnprocessableEntity("name and phone are required")
	}

	user, err := handler.users.Create(c.Request.Context(), name, phone)
	if err != nil {
		return err
	}
	if c.IsHTMX() {
		return c.Render("user-row", Route.Data{"User": user})
	}
	return c.Redirect("/users", http.StatusSeeOther)
}

// Show displays one user.
func (handler *UserHandler) Show(c *Route.Context) error {
	user, err := handler.findUser(c)
	if err != nil {
		return userLookupError(c, err)
	}
	return c.Render("user", Route.Data{"Title": "View User", "User": user})
}

// Edit displays the edit form for one user.
func (handler *UserHandler) Edit(c *Route.Context) error {
	user, err := handler.findUser(c)
	if err != nil {
		return userLookupError(c, err)
	}
	return c.Render("user-edit", Route.Data{"Title": "Edit User", "User": user})
}

// Update updates an existing user.
func (handler *UserHandler) Update(c *Route.Context) error {
	id, err := userID(c)
	if err != nil {
		return c.BadRequest("invalid user ID")
	}
	name := strings.TrimSpace(c.Form("name"))
	phone := strings.TrimSpace(c.Form("phone"))
	if name == "" || phone == "" {
		return c.UnprocessableEntity("name and phone are required")
	}

	if _, err := handler.users.Update(c.Request.Context(), id, name, phone); err != nil {
		return userStoreError(c, err)
	}
	return c.Redirect("/users", http.StatusSeeOther)
}

// Delete removes an existing user.
func (handler *UserHandler) Delete(c *Route.Context) error {
	id, err := userID(c)
	if err != nil {
		return c.BadRequest("invalid user ID")
	}
	if err := handler.users.Delete(c.Request.Context(), id); err != nil {
		return userStoreError(c, err)
	}
	if c.IsHTMX() {
		return c.Text(http.StatusOK, "")
	}
	return c.Redirect("/users", http.StatusSeeOther)
}

func (handler *UserHandler) findUser(c *Route.Context) (models.User, error) {
	id, err := userID(c)
	if err != nil {
		return models.User{}, err
	}
	return handler.users.Find(c.Request.Context(), id)
}

func userStoreError(c *Route.Context, err error) error {
	if errors.Is(err, store.ErrUserNotFound) {
		return c.NotFound("user not found")
	}
	return err
}

func userLookupError(c *Route.Context, err error) error {
	if errors.Is(err, store.ErrUserNotFound) {
		return c.NotFound("user not found")
	}
	return c.BadRequest("invalid user ID")
}

func userID(c *Route.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("invalid user ID")
	}
	return uint(id), err
}
