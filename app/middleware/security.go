package middleware

import Route "fluxgo/internal/route"

// SecurityHeaders applies conservative browser security defaults.
func SecurityHeaders(hsts bool) Route.Middleware {
	return func(next Route.Handler) Route.Handler {
		return func(c *Route.Context) error {
			c.SetHeader("Content-Security-Policy",
				"default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'")
			c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
			c.SetHeader("X-Content-Type-Options", "nosniff")
			c.SetHeader("X-Frame-Options", "DENY")
			c.SetHeader("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			if hsts {
				c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			return next(c)
		}
	}
}
