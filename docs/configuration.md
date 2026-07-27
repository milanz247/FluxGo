# Environment Configuration

FluxGo reads configuration from the root `.env` file during startup. Operating
system environment variables take precedence, which allows production
deployments to inject settings without a file.

```dotenv
APP_NAME=FluxGo
APP_ENV=local
SERVER_ADDR=:8080
VIEWS_ROOT=views
```

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_NAME` | `FluxGo` | Application name used in logs |
| `APP_ENV` | `local` | Current environment name |
| `SERVER_ADDR` | `:8080` | Address passed to `http.ListenAndServe` |
| `VIEWS_ROOT` | `views` | Root containing `layouts/` and `pages/` |

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
