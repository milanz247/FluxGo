# Environment Configuration

FluxGo reads configuration from the root `.env` file during startup. Operating
system environment variables take precedence, which allows production
deployments to inject settings without a file.

```dotenv
APP_NAME=FluxGo
APP_ENV=local
APP_URL=http://localhost:8080
SERVER_ADDR=:8080
VIEWS_ROOT=views
SESSION_COOKIE=flux_session
SESSION_LIFETIME_MINUTES=120
SESSION_SECURE=false
LOG_LEVEL=info
LOG_FORMAT=text
```

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_NAME` | `FluxGo` | Application name used in logs |
| `APP_ENV` | `local` | Current environment name |
| `APP_URL` | `http://localhost:8080` | Base URL used in email links |
| `SERVER_ADDR` | `:8080` | Address passed to `http.ListenAndServe` |
| `VIEWS_ROOT` | `views` | Root containing `layouts/` and `pages/` |
| `SESSION_COOKIE` | `flux_session` | Session cookie name |
| `SESSION_LIFETIME_MINUTES` | `120` | Server-side session lifetime |
| `SESSION_SECURE` | `false` | Restrict the cookie to HTTPS |
| `LOG_LEVEL` | `info` | Minimum level logged: `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `text` | Log encoding: `text` or `json`. See [`logging.md`](logging.md) |

Database settings are documented separately in
[`database.md`](database.md). Keep real database passwords in `.env`; the file
is ignored by Git.

Values may be unquoted, single-quoted, or double-quoted. Blank lines, comments,
and the optional `export` prefix are supported:

```dotenv
# Local development
export APP_ENV="local"
```

The real `.env` is ignored by Git. Copy `.env.example` when setting up a new
environment:

```sh
cp .env.example .env
```
