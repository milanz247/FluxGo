# Sessions

FluxGo provides globally available, server-side sessions. The browser cookie
contains only a cryptographically random session ID; session values remain in
the configured server-side store.

The default bootstrap uses the GORM database store:

```go
sessionStore := session.NewDatabaseStore(db)
sessions := session.New(session.Config{
	CookieName: environment.SessionCookie,
	Lifetime:   environment.SessionLifetime,
	Secure:     environment.SessionSecure,
}, sessionStore)

Route.Use(sessions.Middleware)
```

Sessions therefore survive hot reloads and application restarts. Passing `nil`
still creates a `MemoryStore`, which is useful for isolated unit tests.

## Configuration

```dotenv
SESSION_COOKIE=flux_session
SESSION_LIFETIME_MINUTES=120
SESSION_SECURE=false
```

Set `SESSION_SECURE=true` when the application is served over HTTPS. Session
cookies are `HttpOnly`, use `SameSite=Lax`, and default to the `/` path.

## Handler usage

Set and retrieve values:

```go
func Login(c *Route.Context) error {
	if err := c.Session().Set("user_id", user.ID); err != nil {
		return err
	}

	// Prevent session fixation after authentication.
	if err := c.Session().Regenerate(); err != nil {
		return err
	}

	return c.Redirect("/dashboard")
}

func Dashboard(c *Route.Context) error {
	userID, authenticated := c.Session().Get("user_id")
	if !authenticated {
		return c.Redirect("/login")
	}

	return c.Render("dashboard", Route.Data{"UserID": userID})
}
```

## Available operations

```go
session := c.Session()

session.ID()
session.Get("key")
session.Has("key")
session.Set("key", value)
session.Delete("key")
session.Clear()
session.Regenerate()
session.Destroy()
```

`Pull` retrieves and immediately removes a value, which is useful for flash
messages:

```go
if err := c.Session().Set("success", "Profile saved"); err != nil {
	return err
}

message, exists, err := c.Session().Pull("success")
```

Always mutate sessions before writing/rendering the response because cookie
headers cannot be changed after the response has started.

## Custom stores

A store must safely support concurrent calls:

```go
type Store interface {
	Load(id string, now time.Time) (map[string]any, bool, error)
	Save(id string, values map[string]any, expiresAt time.Time) error
	Delete(id string) error
}
```

Session IDs are generated with `crypto/rand` and contain 256 bits of entropy.
Regenerate the ID after login or any privilege change, and destroy the session
on logout.
