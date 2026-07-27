// Package store contains application data stores.
package store

import (
	"context"
	"errors"
	"strings"

	"fluxgo/app/models"
	"gorm.io/gorm"
)

// ErrUserNotFound indicates that a user ID does not exist.
var ErrUserNotFound = errors.New("user not found")

// UserStore persists users with GORM.
type UserStore struct {
	database *gorm.DB
}

// NewUserStore creates a GORM-backed user repository.
func NewUserStore(database *gorm.DB) *UserStore {
	return &UserStore{database: database}
}

// All returns users ordered by ID.
func (store *UserStore) All(ctx context.Context) ([]models.User, error) {
	var users []models.User
	result := store.database.WithContext(ctx).Order("id ASC").Find(&users)
	return users, result.Error
}

// Find returns a user by ID.
func (store *UserStore) Find(ctx context.Context, id uint) (models.User, error) {
	var user models.User
	err := store.database.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrUserNotFound
	}
	return user, err
}

// Create stores a new user.
func (store *UserStore) Create(ctx context.Context, name, phone string) (models.User, error) {
	user := models.User{
		Name:  strings.TrimSpace(name),
		Phone: strings.TrimSpace(phone),
	}
	err := store.database.WithContext(ctx).Create(&user).Error
	return user, err
}

// Update replaces the editable fields of a user.
func (store *UserStore) Update(
	ctx context.Context,
	id uint,
	name string,
	phone string,
) (models.User, error) {
	result := store.database.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":  strings.TrimSpace(name),
			"phone": strings.TrimSpace(phone),
		})
	if result.Error != nil {
		return models.User{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.User{}, ErrUserNotFound
	}
	return store.Find(ctx, id)
}

// Delete removes a user.
func (store *UserStore) Delete(ctx context.Context, id uint) error {
	result := store.database.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
