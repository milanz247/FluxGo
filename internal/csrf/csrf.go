// Package csrf provides session-backed cross-site request forgery protection.
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	Route "fluxgo/internal/route"
)

const (
	defaultSessionKey = "_csrf_token"
	sessionIDKey      = "_csrf_session_id"
)

// Config controls token input names and excluded paths.
type Config struct {
	FieldName  string
	HeaderName string
	Except     []string
}

// Protection validates unsafe requests against a session-bound token.
type Protection struct {
	config Config
}

// New creates CSRF protection with secure defaults.
func New(config Config) *Protection {
	if config.FieldName == "" {
		config.FieldName = "_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	return &Protection{config: config}
}

// Middleware generates a token and validates it on unsafe HTTP methods.
// Session middleware must be registered before this middleware.
func (protection *Protection) Middleware(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		current := c.Session()
		if current == nil {
			return fmt.Errorf("csrf: session middleware is not registered")
		}

		token, err := protection.ensureToken(current)
		if err != nil {
			return fmt.Errorf("csrf: generate token: %w", err)
		}
		c.Share("CSRFToken", token)

		if isSafeMethod(c.Method()) || protection.excluded(c.Path()) {
			return next(c)
		}

		submitted := c.Header(protection.config.HeaderName)
		if submitted == "" {
			submitted = c.Form(protection.config.FieldName)
		}
		if !tokensMatch(token, submitted) {
			return c.Forbidden("CSRF token mismatch")
		}

		return next(c)
	}
}

func (protection *Protection) ensureToken(current Route.Session) (string, error) {
	tokenValue, hasToken := current.Get(defaultSessionKey)
	boundID, hasBoundID := current.Get(sessionIDKey)
	token, tokenOK := tokenValue.(string)
	sessionID, sessionIDOK := boundID.(string)

	if hasToken && tokenOK && token != "" &&
		hasBoundID && sessionIDOK && sessionID == current.ID() {
		return token, nil
	}

	token, err := newToken()
	if err != nil {
		return "", err
	}
	if err := current.Set(defaultSessionKey, token); err != nil {
		return "", err
	}
	if err := current.Set(sessionIDKey, current.ID()); err != nil {
		return "", err
	}
	return token, nil
}

func (protection *Protection) excluded(path string) bool {
	for _, pattern := range protection.config.Except {
		pattern = strings.TrimSpace(pattern)
		if pattern == path {
			return true
		}
		if strings.HasSuffix(pattern, "*") &&
			strings.HasPrefix(path, strings.TrimSuffix(pattern, "*")) {
			return true
		}
	}
	return false
}

func newToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func tokensMatch(expected, submitted string) bool {
	if expected == "" || submitted == "" || len(expected) != len(submitted) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(submitted)) == 1
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}
