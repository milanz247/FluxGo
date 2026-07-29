# Error Handling

Handlers return a plain `error`. The engine decides what the client sees:

- A `*Route.HTTPError` is rendered as JSON with its own status code.
- Any other error is logged and rendered as a generic `500` — the
  underlying message is never leaked to the client.
- A panic inside a handler or middleware is recovered automatically and
  treated the same as a `500`; it can never take the server down.

```go
func (h *UserHandler) Show(c *Route.Context) error {
	user, err := h.repository.Find(c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		return Route.NotFoundError("that user does not exist")
	}
	if err != nil {
		return Route.InternalServerError("could not load user", err)
	}
	return c.OK(user)
}
```

## Constructors

```go
Route.BadRequestError(message)          // 400
Route.UnauthorizedError(message)        // 401
Route.ForbiddenError(message)           // 403
Route.NotFoundError(message)            // 404
Route.UnprocessableEntityError(message) // 422
Route.InternalServerError(message, err) // 500, err is logged but never shown
Route.NewHTTPError(status, message)     // any other status
```

`message` is what the client receives:

```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"that user does not exist"}
```

## Wrapping a cause for logs only

`InternalServerError` and `NewHTTPError(...).Wrap(err)` attach an underlying
error that shows up in logs (via `error.Error()` and `errors.Unwrap`) without
ever reaching the response body — use this to keep internal detail (SQL
errors, stack traces, third-party responses) out of client-facing messages:

```go
return Route.InternalServerError("could not process payment", err)
```

## Checking a returned error's status

Middleware that needs to know what status a handler's error will produce —
useful for logging before the engine has written the response — can call:

```go
status := Route.StatusForError(err) // an HTTPError's own status, or 500
```

## Panics

A handler that panics (nil pointer, out-of-range index, a third-party
library panicking on bad input) is recovered by the engine and converted
into a `500` response. The client never sees the panic value; it is logged
via `error`. No middleware needs to opt into this — it applies to every
route.

## HTTPError vs `c.BadRequest`/`c.NotFound`/etc.

`Context` also has direct-write helpers documented in
[`handlers.md`](handlers.md#standard-json-errors) — `c.BadRequest(...)`,
`c.NotFound(...)`, and friends. Use whichever fits where the error is
detected:

- **`c.NotFound(...)` etc.** — the handler has `c` in scope and wants to
  write the response immediately.
- **`Route.NotFoundError(...)` etc.** — the error originates somewhere
  without access to `c` (a repository, a service function several calls
  deep) and needs to propagate up through a plain `error` return until the
  engine renders it.

Both produce the same `{"error":"..."}` JSON shape and status code.

## HTML views vs JSON errors

Form-based handlers that render an HTML page on validation failure (see
[Validation](validation.md)) generally don't need `HTTPError` — they
call `c.RenderStatus` directly so they can re-render the form with the
user's input and a field-level message. `HTTPError` is aimed at JSON APIs
and any handler where a plain error status is the right response.
