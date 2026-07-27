# Router Helpers

FluxGo exposes the same route registration helpers on `Engine`, `RouteGroup`,
and the package-level default router.

## HTTP methods

```go
Route.Get("/users", handlers.Index)
Route.Post("/users", handlers.Store)
Route.Put("/users/{id}", handlers.Replace)
Route.Patch("/users/{id}", handlers.Update)
Route.Delete("/users/{id}", handlers.Delete)
Route.Head("/health", handlers.Health)
Route.Options("/users", handlers.Options)
```

## Register multiple methods

`Match` normalizes method names, ignores blank methods, and removes duplicates:

```go
Route.Match(
	[]string{http.MethodGet, http.MethodPost},
	"/session",
	handlers.Session,
)
```

`Any` registers `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, and `OPTIONS`:

```go
Route.Any("/webhook", handlers.Webhook)
```

Prefer explicit methods for normal application endpoints. `Any` is most useful
for diagnostics, catch-all integrations, and endpoints where the method is
intentionally interpreted inside the handler.

## Redirects

The status defaults to `302 Found`:

```go
Route.Redirect("/old", "/new")
Route.Redirect("/legacy", "/current", http.StatusMovedPermanently)
```

## Route groups

Every helper is available on groups and respects group prefixes and middleware:

```go
admin := Route.Group("/admin")
admin.Use(middleware.RequireAdmin)

admin.Patch("/users/{id}", handlers.UpdateUser)
admin.Match([]string{"GET", "POST"}, "/search", handlers.Search)
admin.Redirect("/home", "/admin/dashboard")
```

## View data

Use `Route.Data` instead of declaring a different map type for every view. Its
values may contain strings, booleans, numbers, slices, structs, or any other
template data:

```go
return c.Render("home", Route.Data{
	"Title":   "FluxGo",
	"Heading": "Hello from FluxGo!",
	"Admin":   true,
})
```

Data can also be assembled or merged fluently when that makes conditional view
data clearer:

```go
data := Route.Data{}.
	Set("Title", "Dashboard").
	Merge(sharedData)

return c.Render("dashboard", data)
```
