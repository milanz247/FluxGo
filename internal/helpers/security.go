package helpers

import (
	"html/template"
)

// csrfField returns the standard hidden CSRF input for an application form.
func csrfField(token string) template.HTML {
	escaped := template.HTMLEscapeString(token)
	return template.HTML(`<input type="hidden" name="_token" value="` + escaped + `">`)
}
