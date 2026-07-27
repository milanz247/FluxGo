// Package session provides server-side HTTP sessions.
package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	Route "fluxgo/internal/route"
)

// Config controls session lifetime and cookie security.
type Config struct {
	CookieName string
	Lifetime   time.Duration
	Path       string
	Domain     string
	Secure     bool
	SameSite   http.SameSite
}

// Manager loads sessions and attaches them to route contexts.
type Manager struct {
	store  Store
	config Config
	now    func() time.Time
	mu     sync.Mutex
}

// New creates a session manager. A nil store uses MemoryStore.
func New(config Config, store Store) *Manager {
	if store == nil {
		store = NewMemoryStore()
	}
	if config.CookieName == "" {
		config.CookieName = "flux_session"
	}
	if config.Lifetime <= 0 {
		config.Lifetime = 2 * time.Hour
	}
	if config.Path == "" {
		config.Path = "/"
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}

	return &Manager{
		store:  store,
		config: config,
		now:    time.Now,
	}
}

// Middleware loads or creates a session before the handler runs.
func (manager *Manager) Middleware(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		current, err := manager.start(c.Response, c.Request)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}
		c.SetSession(current)
		return next(c)
	}
}

func (manager *Manager) start(w http.ResponseWriter, r *http.Request) (*requestSession, error) {
	now := manager.now()
	if cookie, err := r.Cookie(manager.config.CookieName); err == nil {
		values, exists, loadErr := manager.store.Load(cookie.Value, now)
		if loadErr != nil {
			return nil, loadErr
		}
		if exists {
			return &requestSession{
				id:      cookie.Value,
				values:  values,
				manager: manager,
				writer:  w,
			}, nil
		}
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}
	current := &requestSession{
		id:      id,
		values:  make(map[string]any),
		manager: manager,
		writer:  w,
	}
	if err := manager.store.Save(id, current.values, now.Add(manager.config.Lifetime)); err != nil {
		return nil, err
	}
	manager.writeCookie(w, id, now.Add(manager.config.Lifetime))
	return current, nil
}

func (manager *Manager) writeCookie(w http.ResponseWriter, id string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    id,
		Path:     manager.config.Path,
		Domain:   manager.config.Domain,
		MaxAge:   int(manager.config.Lifetime.Seconds()),
		Expires:  expiresAt,
		Secure:   manager.config.Secure,
		HttpOnly: true,
		SameSite: manager.config.SameSite,
	})
}

func (manager *Manager) deleteCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    "",
		Path:     manager.config.Path,
		Domain:   manager.config.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   manager.config.Secure,
		HttpOnly: true,
		SameSite: manager.config.SameSite,
	})
}

func newID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate session ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

type requestSession struct {
	id        string
	values    map[string]any
	manager   *Manager
	writer    http.ResponseWriter
	destroyed bool
}

func (session *requestSession) ID() string {
	return session.id
}

func (session *requestSession) Get(key string) (any, bool) {
	value, exists := session.values[key]
	return value, exists
}

func (session *requestSession) Has(key string) bool {
	_, exists := session.Get(key)
	return exists
}

func (session *requestSession) Set(key string, value any) error {
	if session.destroyed {
		return fmt.Errorf("session is destroyed")
	}
	session.values[key] = value
	return session.persist()
}

func (session *requestSession) Pull(key string) (any, bool, error) {
	value, exists := session.values[key]
	if !exists {
		return nil, false, nil
	}
	delete(session.values, key)
	return value, true, session.persist()
}

func (session *requestSession) Delete(key string) error {
	if session.destroyed {
		return fmt.Errorf("session is destroyed")
	}
	delete(session.values, key)
	return session.persist()
}

func (session *requestSession) Clear() error {
	if session.destroyed {
		return fmt.Errorf("session is destroyed")
	}
	session.values = make(map[string]any)
	return session.persist()
}

func (session *requestSession) Regenerate() error {
	if session.destroyed {
		return fmt.Errorf("session is destroyed")
	}

	session.manager.mu.Lock()
	defer session.manager.mu.Unlock()

	id, err := newID()
	if err != nil {
		return err
	}
	if err := session.manager.store.Delete(session.id); err != nil {
		return err
	}
	session.id = id
	expiresAt := session.manager.now().Add(session.manager.config.Lifetime)
	if err := session.manager.store.Save(session.id, session.values, expiresAt); err != nil {
		return err
	}
	session.manager.writeCookie(session.writer, session.id, expiresAt)
	return nil
}

func (session *requestSession) Destroy() error {
	if session.destroyed {
		return nil
	}

	session.manager.mu.Lock()
	defer session.manager.mu.Unlock()

	if err := session.manager.store.Delete(session.id); err != nil {
		return err
	}
	session.values = nil
	session.destroyed = true
	session.manager.deleteCookie(session.writer)
	return nil
}

func (session *requestSession) persist() error {
	session.manager.mu.Lock()
	defer session.manager.mu.Unlock()

	expiresAt := session.manager.now().Add(session.manager.config.Lifetime)
	return session.manager.store.Save(session.id, session.values, expiresAt)
}
