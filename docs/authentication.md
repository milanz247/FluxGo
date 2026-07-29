# Authentication

The application uses a compact Model-View-Handler structure:

- Model: `app/models/user.go`
- Views: `views/pages/login.gohtml`, `register.gohtml`, and `dashboard.gohtml`
- Handler: `app/handlers/auth.go`
- Middleware: `app/middleware/auth.go`

## Routes

```text
GET  /             Redirect based on authentication state
GET  /register     Guest registration form
POST /register     Create an account
GET  /login        Guest login form
POST /login        Authenticate
GET  /dashboard    Authenticated dashboard
POST /logout       Destroy the session
GET  /forgot-password
POST /forgot-password
GET  /reset-password
POST /reset-password
GET  /email/verify
POST /email/verification-notification
```

Registration and login redirect to `/dashboard`. Guest middleware prevents
authenticated users from returning to login/register, while auth middleware
protects the dashboard and logout.

## Passwords

Passwords are stored with PBKDF2-HMAC-SHA256 using:

- 600,000 iterations
- A unique 128-bit random salt
- A 256-bit derived key
- Constant-time hash comparison

Plaintext passwords are never stored. Login uses a dummy password hash when an
email does not exist to reduce timing differences that could reveal registered
email addresses.

## Session security

The session ID is regenerated after registration and login to prevent session
fixation. Logout destroys both the server-side session and browser cookie.
Every state-changing form includes the CSRF token.

## Brute-force protection

Login failures are grouped by a SHA-256 hash of client IP and normalized email.
Five failures within 15 minutes block the bucket for 15 minutes. Buckets live
in MySQL, so hot reloads do not reset the limit. Blocked responses use HTTP 429
and include `Retry-After`.

## Email verification and password reset

Verification links expire after 24 hours. Password reset links expire after one
hour. Only SHA-256 token hashes are stored, and tokens are single-use.

Development uses `LogMailer`, which prints links to the server log. Replace the
`Mailer` implementation with an SMTP or transactional-email adapter before
production.

## Security headers and health

The global security middleware sends CSP, clickjacking, MIME-sniffing, referrer,
and permissions-policy headers. HSTS is enabled when `SESSION_SECURE=true`.
`GET /health` checks the database and returns 200 or 503. The HTTP server uses
timeouts and allows ten seconds for graceful shutdown.

In production, serve the application exclusively over HTTPS and set:

```dotenv
SESSION_SECURE=true
```
