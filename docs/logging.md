# Logging

FluxGo logs through Go's standard structured logger, `log/slog`. The
`fluxgo/internal/logging` package builds the application's logger from
configuration and provides a request-logging middleware.

## Configuration

```env
LOG_LEVEL=info   # debug, info, warn, error
LOG_FORMAT=text  # text or json
```

`bootstrap/main.go` builds the logger once at startup and installs it as the
package-wide default, so `slog.Info`/`slog.Error`/etc. anywhere in the
application use it automatically:

```go
logger := logging.New(logging.Config{Level: environment.LogLevel, Format: environment.LogFormat})
slog.SetDefault(logger)
```

Use `LOG_FORMAT=json` in production so logs are easy to ship to a log
aggregator; `text` is more readable during local development.

## Request logging

`logging.Middleware` records the method, path, response status, and
duration of every request as one structured entry:

```go
Route.Use(logging.Middleware(logger))
```

Passing `nil` resolves `slog.Default()` at request time, so it always
reflects whatever logger is currently installed:

```go
Route.Use(logging.Middleware(nil))
```

Successful requests log at `info`; requests whose handler returned an error
log at `error` with an `error` field attached. The status recorded matches
what the client actually receives — including the status carried by an
[`HTTPError`](errors.md), not a blanket `500`.

```
level=INFO msg="request handled" method=GET path=/dashboard status=200 duration=1.2ms
level=ERROR msg="request failed" method=POST path=/login status=422 duration=800µs error="..."
```

## Logging elsewhere in the application

Anywhere outside a request — background jobs, the session cleanup ticker in
`bootstrap/main.go`, one-off scripts — use the standard `log/slog`
package-level functions directly:

```go
slog.Info("cache warmed", "entries", count)
slog.Error("payment webhook failed", "error", err)
```

Because `slog.SetDefault` was called at startup, these calls share the same
level and output configuration as request logs without needing a logger
passed around explicitly.
