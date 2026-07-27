package session

import "time"

// Store persists server-side session values.
type Store interface {
	Load(id string, now time.Time) (map[string]any, bool, error)
	Save(id string, values map[string]any, expiresAt time.Time) error
	Delete(id string) error
}
