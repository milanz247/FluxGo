# Validation

`fluxgo/internal/validation` validates structs against `validate` struct
tags, so a handler describes its input requirements once instead of writing
a chain of `if` checks.

```go
type registerInput struct {
	Name                 string `validate:"required" label:"Name"`
	Email                string `validate:"required,email" label:"Email"`
	Password             string `validate:"required,min=8,max=1024" label:"Password"`
	PasswordConfirmation string `validate:"required,eqfield=Password" label:"Password confirmation"`
}

input := registerInput{Name: name, Email: email, Password: password, PasswordConfirmation: confirmation}
if err := validation.Validate(input); err != nil {
	var errs validation.Errors
	errors.As(err, &errs)
	return renderAuthError(c, "auth/register", data, errs.First("Email"))
}
```

`Validate` accepts a struct or a pointer to one. It returns `nil` when every
rule passes, or `validation.Errors` (a `map[string][]string]` keyed by field
label) when any rule fails.

## Rules

| Rule | Applies to | Meaning |
| --- | --- | --- |
| `required` | any | not the zero value |
| `email` | string | a valid, unambiguous email address |
| `url` | string | a valid absolute URL |
| `numeric` | string | digits only |
| `alpha` | string | letters only |
| `alphanumeric` | string | letters and digits only |
| `min=N` | string, number, slice/map | length/value/size at least N |
| `max=N` | string, number, slice/map | length/value/size at most N |
| `len=N` | string, slice/map | exact length |
| `oneof=a\|b\|c` | any | value is one of the listed options |
| `eqfield=Field` | any | value equals a sibling field's value |

Combine rules with commas: `validate:"required,min=8,max=1024"`. Empty
fields skip every rule except `required`, so `min`/`max`/`email`/etc. only
fire once a value is present — pair them with `required` when a field is
mandatory.

## Field labels

Error messages use the Go field name by default. Add a `label` tag to show
users something friendlier:

```go
Confirm string `validate:"required,eqfield=Password" label:"Password confirmation"`
```

## Reading errors

`validation.Errors` supports:

```go
errs.Has("Email")      // bool
errs.First("Email")    // first message, or ""
errs.Error()            // every message joined into one string
```

Handlers typically pick the first relevant field to show as a single error,
in a priority order that matches the form layout:

```go
func firstError(errs validation.Errors, fields ...string) string {
	for _, field := range fields {
		if message := errs.First(field); message != "" {
			return field + " " + message + "."
		}
	}
	return "The submitted data is invalid."
}
```

See `app/handlers/auth.go` for the full pattern used by registration and
password reset.
