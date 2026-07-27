package handlers

import Route "fluxgo/internal/route"

// WelcomeHandler handles the application welcome page.
func WelcomeHandler(c *Route.Context) error {
	return c.View(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>FluxGo</title>
</head>
<body>
	<h1>Hello from FluxGo!</h1>
	<p>Your route engine is running.</p>
</body>
</html>`)
}
