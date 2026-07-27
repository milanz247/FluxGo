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
DB_AUTO_MIGRATE=true
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

## Models and migration

The user model is in `app/models/user.go`. When `DB_AUTO_MIGRATE=true`, startup
runs:

```go
database.Migrate(db, &models.User{})
```

GORM `AutoMigrate` is convenient during development. For production systems,
set `DB_AUTO_MIGRATE=false` and use reviewed, versioned migrations.

The current development migration also removes the obsolete `phone` column
from the earlier CRUD example. This one-time cleanup allows the authentication
model (`name`, `email`, and `password_hash`) to write to an existing development
`users` table.

## MVH database access

The application intentionally uses a small Model-View-Handler structure. Auth
handlers receive `*gorm.DB` and use it directly with the HTTP request context;
there is no repository or service layer in this example.

The DSN enables `parseTime` for `time.Time` model fields and uses `utf8mb4`.
Passwords are not logged by the application.
