# GoHTML Template Helpers

FluxGo registers these helpers globally before compiling layouts and pages. They
work in both full-page and HTMX partial renders. No import or bootstrap
registration is required.

```go
views, err := view.New(view.Config{Root: "views"})
```

Go's built-in template functions such as `and`, `or`, `not`, `eq`, `ne`, `lt`,
`le`, `gt`, `ge`, `len`, `index`, `slice`, `printf`, `html`, and `urlquery`
remain available.

## String helpers

| Helper | Example | Result |
| --- | --- | --- |
| `upper` | `{{upper "flux"}}` | `FLUX` |
| `lower` | `{{lower "GO"}}` | `go` |
| `trim` | `{{trim "  hello  "}}` | `hello` |
| `capitalize` | `{{capitalize "flux"}}` | `Flux` |
| `contains` | `{{contains "go" "fluxgo"}}` | `true` |
| `hasPrefix` | `{{hasPrefix "flux" "fluxgo"}}` | `true` |
| `hasSuffix` | `{{hasSuffix "go" "fluxgo"}}` | `true` |
| `trimPrefix` | `{{trimPrefix "/" "/users"}}` | `users` |
| `trimSuffix` | `{{trimSuffix ".gohtml" "home.gohtml"}}` | `home` |
| `replace` | `{{replace " " "-" "hello go"}}` | `hello-go` |
| `split` | `{{split "," "go,html"}}` | `[]string{"go","html"}` |
| `join` | `{{join ", " .Tags}}` | Joins a `[]string` |
| `slug` | `{{slug "Hello, Go!"}}` | `hello-go` |
| `truncate` | `{{truncate 10 .Description}}` | At most 10 characters |
| `queryEscape` | `{{queryEscape .Search}}` | URL query-safe text |

Arguments are pipeline-friendly because the value is normally last:

```gohtml
{{.Name | trim | upper}}
{{.Title | replace " " "-"}}
{{.Description | truncate 80}}
```

## Defaults and conditions

```gohtml
{{default "Anonymous" .Name}}
{{coalesce .Nickname .Name "Anonymous"}}
{{ternary "Active" "Inactive" .Active}}

{{/* Pipeline form */}}
{{.Active | ternary "Active" "Inactive"}}
```

`default` and `coalesce` treat `nil`, `false`, numeric zero, empty strings,
empty arrays, slices, and maps as empty.

## Collection helpers

```gohtml
{{/* Build values inline */}}
{{$roles := list "admin" "editor"}}
{{if in .Role $roles}}Allowed{{end}}

{{$user := dict "name" "Milan" "active" true}}
{{$user.name}}

{{range keys .Settings}}
	<p>{{.}}</p>
{{end}}
```

`keys` accepts maps with string keys and returns sorted keys. `dict` requires
string keys and an even number of key/value arguments. Invalid input stops
template execution and returns an error to the handler.

## Number helpers

```gohtml
{{add 10 5}}       {{/* 15 */}}
{{sub 10 5}}       {{/* 5 */}}
{{mul 10 5}}       {{/* 50 */}}
{{div 10 5}}       {{/* 2 */}}
{{mod 10 3}}       {{/* 1 */}}
{{inc .Index}}
{{dec .Page}}
{{even .Index}}
{{odd .Index}}

{{range seq 1 5}}
	<span>{{.}}</span>
{{end}}
```

`seq` includes both endpoints and supports descending ranges (`seq 5 1`). It
is limited to 10,000 values. Division or modulo by zero returns a template
execution error.

## Date and time helpers

The `date` helper accepts `time.Time` or `*time.Time`:

```gohtml
{{date "2006-01-02" .CreatedAt}}
{{date "02 Jan 2006, 15:04" .UpdatedAt}}
{{date "15:04:05" (now)}}
```

Date layouts use Go's reference time format: `Mon Jan 2 15:04:05 MST 2006`.

## Adding a global helper

Add the implementation under `internal/helpers`, then register its template
name in `TemplateFuncs`:

```go
func greeting(name string) string {
	return "Hello " + name
}
```

```go
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"greeting": greeting,
		// Existing helpers...
	}
}
```

It becomes available automatically in every `.gohtml` file:

```gohtml
<h1>{{greeting .Name}}</h1>
```

Helpers should be fast, deterministic formatting functions. Keep database
queries, HTTP calls, and application state changes in handlers or services.
`html/template` continues to escape normal helper output automatically; avoid
returning `template.HTML` unless the content is fully trusted.
