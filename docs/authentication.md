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

In production, serve the application exclusively over HTTPS and set:

```dotenv
SESSION_SECURE=true
```
