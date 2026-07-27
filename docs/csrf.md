# CSRF Protection

FluxGo globally protects state-changing browser requests with a synchronizer
token stored in the server-side session. It follows the standard pattern used
by frameworks such as Laravel:

- `GET`, `HEAD`, `OPTIONS`, and `TRACE` are allowed without validation.
- Other methods require the session token in `_token` or `X-CSRF-Token`.
- Tokens are generated using `crypto/rand` and compared in constant time.
- The token changes after the session ID is regenerated.

Session middleware must run before CSRF middleware. The default bootstrap
already registers them in the correct order:

```go
Route.Use(sessions.Middleware)
Route.Use(csrfProtection.Middleware)
```

## GoHTML forms

Use the global `csrfField` helper in every state-changing form:

```gohtml
<form method="POST" action="/profile">
	{{csrfField .CSRFToken}}

	<input name="name" value="{{.Name}}">
	<button type="submit">Save</button>
</form>
```

It generates:

```html
<input type="hidden" name="_token" value="...">
```

CSRF middleware automatically shares `CSRFToken` with `Route.Data`, so handlers
do not need to pass it manually:

```go
return c.Render("profile", Route.Data{
	"Name": user.Name,
})
```

The same hidden field works with HTMX forms because HTMX submits successful
form controls with the request.

## Header-based requests

JavaScript clients may send the token in a header:

```http
X-CSRF-Token: token-value
```

A page can expose it in a meta tag:

```gohtml
<meta name="csrf-token" content="{{.CSRFToken}}">
```

## Excluded routes

External webhooks cannot know a user's session token. Exclude narrowly scoped
paths when creating the middleware:

```go
csrfProtection := csrf.New(csrf.Config{
	Except: []string{
		"/webhooks/stripe",
		"/webhooks/github/*",
	},
})
```

An ending `*` performs a prefix match. Never exclude normal authenticated
browser routes. Verify webhook signatures independently.

## Failure response

Missing or incorrect tokens receive:

```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{"error":"CSRF token mismatch"}
```

Applications that authenticate exclusively through bearer tokens do not
usually require CSRF protection because browsers do not attach those
credentials automatically. Such APIs should use a separate route group or
narrow exclusions; cookie-authenticated APIs still require CSRF protection.
