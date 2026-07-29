# MySQL and GORM

FluxGo uses GORM with the official MySQL driver. Database settings are loaded
from `.env`:

```dotenv
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=fluxgo
DB_USERNAME=root
DB_PASSWORD=
DB_CHARSET=utf8mb4
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME_MINUTES=60
DB_RUN_MIGRATIONS=true
```

Create the development database before starting the application:

```sql
CREATE DATABASE fluxgo
	CHARACTER SET utf8mb4
	COLLATE utf8mb4_unicode_ci;
```

Then run:

```sh
flux dev
```

The application verifies the connection with `Ping` during startup. Pool sizes
and connection lifetime are configured through `database/sql`.

## Versioned migrations

When `DB_RUN_MIGRATIONS=true`, startup applies pending migrations from the
ordered registry in `internal/database/migrations.go`. Applied versions are
recorded in `schema_migrations`, so every migration runs once.

```go
database.RunMigrations(db, database.DefaultMigrations())
```

The migration history creates users, persistent sessions, password reset
tokens, verification tokens, and login rate-limit buckets. It also includes the
one-time removal of the obsolete `phone` column from the CRUD prototype.

## MVH database access

The application intentionally uses a small Model-View-Handler structure. Auth
handlers receive `*gorm.DB` and use it directly with the HTTP request context;
there is no repository or service layer in this example.

The DSN enables `parseTime` for `time.Time` model fields and uses `utf8mb4`.
Passwords are not logged by the application.
