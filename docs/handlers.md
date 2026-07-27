# Handler Context

FluxGo's context helpers are thin wrappers around `net/http`. They reduce
repetitive request and response code without hiding Go's standard behavior.

## Request values

```go
id := c.Param("id")
search := c.Query("search")
page := c.QueryDefault("page", "1")
email := c.Form("email")
authorization := c.Header("Authorization")
token := c.BearerToken()
```

`Form` uses `http.Request.FormValue`, so it supports URL-encoded and multipart
form values. Use `c.Request` directly when parse errors or uploaded files need
explicit handling.

## Request information

```go
c.Method()
c.Path()
c.IsHTMX()
c.IsJSON()
```

The underlying request remains available as `c.Request`.

The globally registered session is available through `c.Session()`. See
[`sessions.md`](sessions.md) for session configuration and operations.

CSRF middleware shares `CSRFToken` automatically with `Route.Data` used for
views. See [`csrf.md`](csrf.md) for forms, HTMX, headers, and exclusions.

## Responses

```go
return c.OK(Route.Data{"user": user})
return c.Created(Route.Data{"user": user})
return c.JSON(http.StatusAccepted, payload)
return c.Text(http.StatusOK, "healthy")
return c.NoContent()
return c.Render("home", Route.Data{"Title": "Home"})
return c.Redirect("/login")
return c.Redirect("/login", http.StatusTemporaryRedirect)
```

Response headers and cookies:

```go
c.SetHeader("X-Request-ID", requestID)

c.SetCookie(&http.Cookie{
	Name:     "session",
	Value:    token,
	Path:     "/",
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteLaxMode,
})

cookie, err := c.Cookie("session")
c.DeleteCookie("session")
```

## Standard JSON errors

Each helper returns `{"error":"message"}` with the matching HTTP status:

```go
return c.BadRequest("invalid request")             // 400
return c.Unauthorized("authentication required")   // 401
return c.Forbidden("access denied")                // 403
return c.NotFound("user not found")                // 404
return c.UnprocessableEntity("validation failed") // 422
return c.ServerError("internal error")              // 500
```

## Middleware values

Middleware can pass request-scoped values to handlers:

```go
func Auth(next Route.Handler) Route.Handler {
	return func(c *Route.Context) error {
		user, err := authenticate(c.BearerToken())
		if err != nil {
			return c.Unauthorized("invalid token")
		}

		c.Set("user", user)
		return next(c)
	}
}
```

Retrieve optional or required values in the handler:

```go
user, exists := c.Get("user")
user := c.MustGet("user")
```

`MustGet` panics when the key is absent, so use it only when middleware
guarantees that the value has been set. Store request-specific values only;
shared application state belongs in services or dependencies.
